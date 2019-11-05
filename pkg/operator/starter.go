package operator

import (
	"fmt"
	"time"

	"github.com/openshift/cluster-etcd-operator/pkg/operator/hostetcdendpointcontroller"

	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned"
	configv1informers "github.com/openshift/client-go/config/informers/externalversions"
	operatorversionedclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorv1informers "github.com/openshift/client-go/operator/informers/externalversions"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"github.com/openshift/cluster-etcd-operator/pkg/operator/bootstrapteardown"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/clustermembercontroller"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/configobservation/configobservercontroller"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/etcdcertsigner"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/operatorclient"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/resourcesynccontroller"
	operatorv1helpers "github.com/openshift/library-go/pkg/operator/v1helpers"
)

func RunOperator(ctx *controllercmd.ControllerContext) error {
	// This kube client use protobuf, do not use it for CR
	kubeClient, err := kubernetes.NewForConfig(ctx.ProtoKubeConfig)
	if err != nil {
		return err
	}
	operatorConfigClient, err := operatorversionedclient.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	configClient, err := configv1client.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(ctx.KubeConfig)
	if err != nil {
		return err
	}

	operatorConfigInformers := operatorv1informers.NewSharedInformerFactory(operatorConfigClient, 10*time.Minute)
	//operatorConfigInformers.ForResource()
	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(
		kubeClient,
		"",
		operatorclient.GlobalUserSpecifiedConfigNamespace,
		operatorclient.GlobalMachineSpecifiedConfigNamespace,
		operatorclient.TargetNamespace,
		operatorclient.OperatorNamespace,
		"openshift-etcd",
	)
	configInformers := configv1informers.NewSharedInformerFactory(configClient, 10*time.Minute)
	operatorClient := &operatorclient.OperatorClient{
		Informers: operatorConfigInformers,
		Client:    operatorConfigClient.OperatorV1(),
	}

	resourceSyncController, err := resourcesynccontroller.NewResourceSyncController(
		operatorClient,
		kubeInformersForNamespaces,
		kubeClient,
		ctx.EventRecorder,
	)
	if err != nil {
		return err
	}

	configObserver := configobservercontroller.NewConfigObserver(
		operatorClient,
		operatorConfigInformers,
		kubeInformersForNamespaces,
		configInformers,
		resourceSyncController,
		ctx.EventRecorder,
	)

	versionRecorder := status.NewVersionGetter()
	clusterOperator, err := configClient.ConfigV1().ClusterOperators().Get("openshift-etcd", metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, version := range clusterOperator.Status.Versions {
		versionRecorder.SetVersion(version.Name, version.Version)
	}
	versionRecorder.SetVersion("raw-internal", status.VersionForOperatorFromEnv())
	versionRecorder.SetVersion("operator", status.VersionForOperatorFromEnv())

	clusterOperatorStatus := status.NewClusterOperatorStatusController(
		"openshift-etcd",
		[]configv1.ObjectReference{
			{Group: "operator.openshift.io", Resource: "etcds", Name: "cluster"},
			{Resource: "namespaces", Name: operatorclient.GlobalUserSpecifiedConfigNamespace},
			{Resource: "namespaces", Name: operatorclient.GlobalMachineSpecifiedConfigNamespace},
			{Resource: "namespaces", Name: operatorclient.OperatorNamespace},
			{Resource: "namespaces", Name: operatorclient.TargetNamespace},
		},
		configClient.ConfigV1(),
		configInformers.Config().V1().ClusterOperators(),
		operatorClient,
		versionRecorder,
		ctx.EventRecorder,
	)
	clusterInfrastructure, err := configClient.ConfigV1().Infrastructures().Get("cluster", metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	etcdDiscoveryDomain := clusterInfrastructure.Status.EtcdDiscoveryDomain
	coreClient := clientset

	etcdCertSignerController := etcdcertsigner.NewEtcdCertSignerController(
		coreClient,
		operatorClient,
		kubeInformersForNamespaces.InformersFor("openshift-etcd"),
		ctx.EventRecorder,
	)
	hostEtcdEndpointController := hostetcdendpointcontroller.NewHostEtcdEndpointcontroller(
		coreClient,
		operatorClient,
		kubeInformersForNamespaces.InformersFor("openshift-etcd"),
		ctx.EventRecorder,
	)

	clusterMemberController := clustermembercontroller.NewClusterMemberController(
		coreClient,
		operatorClient,
		kubeInformersForNamespaces.InformersFor("openshift-etcd"),
		ctx.EventRecorder,
		etcdDiscoveryDomain,
	)
	operatorSpec, _, _, err := operatorClient.GetOperatorState()
	if err != nil {
		return err
	}

	operatorConfigInformers.Start(ctx.Done())
	kubeInformersForNamespaces.Start(ctx.Done())
	configInformers.Start(ctx.Done())

	go etcdCertSignerController.Run(1, ctx.Done())
	go hostEtcdEndpointController.Run(1, ctx.Done())
	go resourceSyncController.Run(1, ctx.Done())
	if operatorSpec.ManagementState == operatorv1.Managed {
		go clusterOperatorStatus.Run(1, ctx.Done())
	} else {
		versionCh := versionRecorder.VersionChangedChannel()
		go func() {
			for {
				select {
				case <-versionCh:
					// TODO find a better way of doing this
					versions := versionRecorder.GetVersions()
					originalClusterOperatorObj, err := configInformers.Config().V1().ClusterOperators().Lister().Get("openshift-etcd")
					if err != nil {
						klog.Errorf("error getting cluster operator %#v", err)
						break
					}
					clusterOperatorObj := originalClusterOperatorObj.DeepCopy()
					for operand, version := range versions {
						previousVersion := operatorv1helpers.SetOperandVersion(&clusterOperatorObj.Status.Versions, configv1.OperandVersion{Name: operand, Version: version})
						if previousVersion != version {
							// having this message will give us a marker in events when the operator updated compared to when the operand is updated
							klog.Infof("clusteroperator/%s version %q changed from %q to %q", "openshift-etcd", operand, previousVersion, version)
						}
					}
					if _, updateErr := configClient.ConfigV1().ClusterOperators().UpdateStatus(clusterOperatorObj); err != nil {
						klog.Errorf("error updage clusteroperator/%s %#v", "openshift-etcd", updateErr)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go configObserver.Run(1, ctx.Done())
	go clusterMemberController.Run(ctx.Done())
	go func() {
		err := bootstrapteardown.TearDownBootstrap(ctx.KubeConfig, clusterMemberController, operatorClient)
		if err != nil {
			klog.Fatalf("Error tearing down bootstrap: %#v", err)
		}
	}()

	<-ctx.Done()
	return fmt.Errorf("stopped")
}
