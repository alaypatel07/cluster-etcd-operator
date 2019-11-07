package targetconfigcontroller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-etcd-operator/pkg/operator/operatorclient"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/v430_00_assets"
	"github.com/openshift/cluster-etcd-operator/pkg/version"
)

const workQueueKey = "key"

type TargetConfigController struct {
	targetImagePullSpec   string
	operatorImagePullSpec string

	operatorClient v1helpers.StaticPodOperatorClient

	kubeClient      kubernetes.Interface
	configMapLister corev1listers.ConfigMapLister
	eventRecorder   events.Recorder

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface
}

func NewTargetConfigController(targetImagePullSpec, operatorImagePullSpec string, operatorClient v1helpers.StaticPodOperatorClient, kubeInformersForOpenshiftEtcdNamespace informers.SharedInformerFactory, kubeInformersForNamespaces v1helpers.KubeInformersForNamespaces, kubeClient kubernetes.Interface, eventRecorder events.Recorder) *TargetConfigController {
	c := &TargetConfigController{
		targetImagePullSpec:   targetImagePullSpec,
		operatorImagePullSpec: operatorImagePullSpec,

		operatorClient:  operatorClient,
		kubeClient:      kubeClient,
		configMapLister: kubeInformersForNamespaces.ConfigMapLister(),
		eventRecorder:   eventRecorder.WithComponentSuffix("target-config-controller"),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TargetConfigController"),
	}

	operatorClient.Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftEtcdNamespace.Rbac().V1().Roles().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftEtcdNamespace.Rbac().V1().RoleBindings().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftEtcdNamespace.Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftEtcdNamespace.Core().V1().Secrets().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftEtcdNamespace.Core().V1().ServiceAccounts().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForOpenshiftEtcdNamespace.Core().V1().Services().Informer().AddEventHandler(c.eventHandler())

	// we react to some config changes
	kubeInformersForNamespaces.InformersFor(operatorclient.GlobalUserSpecifiedConfigNamespace).Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForNamespaces.InformersFor(operatorclient.GlobalMachineSpecifiedConfigNamespace).Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForNamespaces.InformersFor(operatorclient.OperatorNamespace).Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())
	kubeInformersForNamespaces.InformersFor(operatorclient.TargetNamespace).Core().V1().ConfigMaps().Informer().AddEventHandler(c.eventHandler())

	return c
}

func (c TargetConfigController) sync() error {
	operatorSpec, _, _, err := c.operatorClient.GetStaticPodOperatorState()
	if err != nil {
		return err
	}

	switch operatorSpec.ManagementState {
	case operatorv1.Managed:
	case operatorv1.Unmanaged:
		return nil
	case operatorv1.Removed:
		// TODO probably just fail
		return nil
	default:
		c.eventRecorder.Warningf("ManagementStateUnknown", "Unrecognized operator management state %q", operatorSpec.ManagementState)
		return nil
	}

	// block until config is observed and specific paths are present
	if err := isRequiredConfigPresent(operatorSpec.ObservedConfig.Raw); err != nil {
		c.eventRecorder.Warning("ConfigMissing", err.Error())
		return err
	}

	requeue, err := createTargetConfig(c, c.eventRecorder)
	if err != nil {
		return err
	}
	if requeue {
		return fmt.Errorf("synthetic requeue request")
	}

	return nil
}

func isRequiredConfigPresent(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("no observedConfig")
	}

	existingConfig := map[string]interface{}{}
	if err := json.NewDecoder(bytes.NewBuffer(config)).Decode(&existingConfig); err != nil {
		return fmt.Errorf("error parsing config, %v", err)
	}

	requiredPaths := [][]string{
		{"servingInfo", "namedCertificates"},
		{"storageConfig", "urls"},
	}
	for _, requiredPath := range requiredPaths {
		configVal, found, err := unstructured.NestedFieldNoCopy(existingConfig, requiredPath...)
		if err != nil {
			return fmt.Errorf("error reading %v from config, %v", strings.Join(requiredPath, "."), err)
		}
		if !found {
			return fmt.Errorf("%v missing from config", strings.Join(requiredPath, "."))
		}
		if configVal == nil {
			return fmt.Errorf("%v null in config", strings.Join(requiredPath, "."))
		}
		if configValSlice, ok := configVal.([]interface{}); ok && len(configValSlice) == 0 {
			return fmt.Errorf("%v empty in config", strings.Join(requiredPath, "."))
		}
		if configValString, ok := configVal.(string); ok && len(configValString) == 0 {
			return fmt.Errorf("%v empty in config", strings.Join(requiredPath, "."))
		}
	}
	return nil
}

// createTargetConfig takes care of creation of valid resources in a fixed name.  These are inputs to other control loops.
// returns whether or not requeue and if an error happened when updating status.  Normally it updates status itself.
func createTargetConfig(c TargetConfigController, recorder events.Recorder) (bool, error) {
	errors := []error{}

	directResourceResults := resourceapply.ApplyDirectly(c.kubeClient, c.eventRecorder, v430_00_assets.Asset,
		"v4.3.0/etcd/ns.yaml",
		"v4.3.0/etcd/svc.yaml",
	)

	for _, currResult := range directResourceResults {
		if currResult.Error != nil {
			errors = append(errors, fmt.Errorf("%q (%T): %v", currResult.File, currResult.Type, currResult.Error))
		}
	}

	//_, _, err := manageEtcdConfig(c.kubeClient.CoreV1(), recorder, operatorSpec)
	//if err != nil {
	//	errors = append(errors, fmt.Errorf("%q: %v", "configmap/config", err))
	//}
	_, _, err := managePod(c.kubeClient.CoreV1(), recorder, c.targetImagePullSpec, c.operatorImagePullSpec)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q: %v", "configmap/kube-apiserver-pod", err))
	}

	if len(errors) > 0 {
		condition := operatorv1.OperatorCondition{
			Type:    "TargetConfigControllerDegraded",
			Status:  operatorv1.ConditionTrue,
			Reason:  "SynchronizationError",
			Message: v1helpers.NewMultiLineAggregate(errors).Error(),
		}
		if _, _, err := v1helpers.UpdateStaticPodStatus(c.operatorClient, v1helpers.UpdateStaticPodConditionFn(condition)); err != nil {
			return true, err
		}
		return true, nil
	}

	condition := operatorv1.OperatorCondition{
		Type:   "TargetConfigControllerDegraded",
		Status: operatorv1.ConditionFalse,
	}
	if _, _, err := v1helpers.UpdateStaticPodStatus(c.operatorClient, v1helpers.UpdateStaticPodConditionFn(condition)); err != nil {
		return true, err
	}

	return false, nil
}

func managePod(client coreclientv1.ConfigMapsGetter, recorder events.Recorder, imagePullSpec, operatorImagePullSpec string) (*corev1.ConfigMap, bool, error) {
	required := resourceread.ReadPodV1OrDie(v430_00_assets.MustAsset("v4.3.0/etcd/pod.yaml"))
	// TODO: If the image pull spec is not specified, the "${IMAGE}" will be used as value and the pod will fail to start.
	images := map[string]string{
		"${IMAGE}":          imagePullSpec,
		"${OPERATOR_IMAGE}": operatorImagePullSpec,
	}
	if len(imagePullSpec) > 0 {
		for i := range required.Spec.Containers {
			for pat, img := range images {
				if required.Spec.Containers[i].Image == pat {
					required.Spec.Containers[i].Image = img
					break
				}
			}
		}
		for i := range required.Spec.InitContainers {
			for pat, img := range images {
				if required.Spec.InitContainers[i].Image == pat {
					required.Spec.InitContainers[i].Image = img
					break
				}
			}
		}
	}

	configMap := resourceread.ReadConfigMapV1OrDie(v430_00_assets.MustAsset("v4.3.0/etcd/pod-cm.yaml"))
	configMap.Data["pod.yaml"] = resourceread.WritePodV1OrDie(required)
	//configMap.Data["forceRedeploymentReason"] = operatorSpec.ForceRedeploymentReason
	configMap.Data["version"] = version.Get().String()
	return resourceapply.ApplyConfigMap(client, recorder, configMap)
}

// Run starts the kube-apiserver and blocks until stopCh is closed.
func (c *TargetConfigController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting TargetConfigController")
	defer klog.Infof("Shutting down TargetConfigController")

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *TargetConfigController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *TargetConfigController) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	err := c.sync()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
func (c *TargetConfigController) eventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(workQueueKey) },
	}
}
