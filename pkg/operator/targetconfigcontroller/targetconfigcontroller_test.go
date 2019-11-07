package targetconfigcontroller

import (
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/events/eventstesting"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"reflect"
	"testing"
)

func Test_isRequiredConfigPresent(t *testing.T) {
	tests := []struct {
		name          string
		config        string
		wantErr       bool
		expectedError string
	}{
		{
			name: "unparseable",
			config: `{
		 "servingInfo": {
		}
		`,
			wantErr:       true,
			expectedError: "error parsing config",
		},
		{
			name:          "empty",
			config:        ``,
			wantErr:       true,
			expectedError: "no observedConfig",
		},
		{
			name: "nil-storage-urls",
			config: `{
		 "servingInfo": {
		   "namedCertificates": [
		     {
		       "certFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.crt",
		       "keyFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.key"
		     }
		   ]
		 },
		 "admission": {"pluginConfig": { "network.openshift.io/RestrictedEndpointsAdmission": {}}},
		 "storageConfig": {
		   "urls": null
		 }
		}
		`,
			wantErr:       true,
			expectedError: "storageConfig.urls null in config",
		},
		{
			name: "missing-storage-urls",
			config: `{
		 "servingInfo": {
		   "namedCertificates": [
		     {
		       "certFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.crt",
		       "keyFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.key"
		     }
		   ]
		 },
		 "storageConfig": {
		   "urls": []
		 }
		}
		`,
			wantErr:       true,
			expectedError: "storageConfig.urls empty in config",
		},
		{
			name: "empty-string-storage-urls",
			config: `{
  "servingInfo": {
    "namedCertificates": [
      {
        "certFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.crt",
        "keyFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.key"
      }
    ]
  },
  "storageConfig": {
    "urls": ""
  }
}
`,
			wantErr:       true,
			expectedError: "storageConfig.urls empty in config",
		},
		{
			name: "good",
			config: `{
		 "servingInfo": {
		   "namedCertificates": [
		     {
		       "certFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.crt",
		       "keyFile": "/etc/kubernetes/static-pod-certs/secrets/localhost-serving-cert-certkey/tls.key"
		     }
		   ]
		 },
		 "storageConfig": {
		   "urls": [ "val" ]
		 }
		}
		`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := isRequiredConfigPresent([]byte(tt.config)); (err != nil) != tt.wantErr {
				t.Errorf("isRequiredConfigPresent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createTargetConfig(t *testing.T) {
	fakeTargetController := TargetConfigController{
		targetImagePullSpec:   "",
		operatorImagePullSpec: "",
		operatorClient:        v1helpers.NewFakeStaticPodOperatorClient(nil, &operatorv1.StaticPodOperatorStatus{}, nil, nil),
		kubeClient:            fake.NewSimpleClientset(),
		eventRecorder:         eventstesting.NewTestingEventRecorder(t),
		queue:                 nil,
	}
	type args struct {
		c        TargetConfigController
		recorder events.Recorder
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				c:        fakeTargetController,
				recorder: fakeTargetController.eventRecorder,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createTargetConfig(tt.args.c, tt.args.recorder)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTargetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createTargetConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_managePod(t *testing.T) {
	fakeConfigMapsGetter := fake.NewSimpleClientset().CoreV1()
	type args struct {
		client                v1.ConfigMapsGetter
		recorder              events.Recorder
		imagePullSpec         string
		operatorImagePullSpec string
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.ConfigMap
		want1   bool
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				client:                fakeConfigMapsGetter,
				recorder:              eventstesting.NewTestingEventRecorder(t),
				imagePullSpec:         "foo",
				operatorImagePullSpec: "fooOperator",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := managePod(tt.args.client, tt.args.recorder, tt.args.imagePullSpec, tt.args.operatorImagePullSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("managePod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("managePod() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("managePod() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
