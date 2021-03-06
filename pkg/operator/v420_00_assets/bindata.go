// Code generated by go-bindata.
// sources:
// bindata/v4.2.0/etcd/cm.yaml
// bindata/v4.2.0/etcd/defaultconfig.yaml
// bindata/v4.2.0/etcd/ns.yaml
// bindata/v4.2.0/etcd/operator-config.yaml
// bindata/v4.2.0/etcd/pod-cm.yaml
// bindata/v4.2.0/etcd/pod.yaml
// bindata/v4.2.0/etcd/sa.yaml
// bindata/v4.2.0/etcd/svc.yaml
// DO NOT EDIT!

package v420_00_assets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _v420EtcdCmYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: openshift-etcd
  name: config
data:
  config.yaml:
`)

func v420EtcdCmYamlBytes() ([]byte, error) {
	return _v420EtcdCmYaml, nil
}

func v420EtcdCmYaml() (*asset, error) {
	bytes, err := v420EtcdCmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/cm.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdDefaultconfigYaml = []byte(`apiVersion: kubecontrolplane.config.openshift.io/v1
kind: EtcdConfig
`)

func v420EtcdDefaultconfigYamlBytes() ([]byte, error) {
	return _v420EtcdDefaultconfigYaml, nil
}

func v420EtcdDefaultconfigYaml() (*asset, error) {
	bytes, err := v420EtcdDefaultconfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/defaultconfig.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdNsYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  annotations:
    openshift.io/node-selector: ""
  name: openshift-etcd
  labels:
    openshift.io/run-level: "0"
`)

func v420EtcdNsYamlBytes() ([]byte, error) {
	return _v420EtcdNsYaml, nil
}

func v420EtcdNsYaml() (*asset, error) {
	bytes, err := v420EtcdNsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/ns.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdOperatorConfigYaml = []byte(`apiVersion: operator.openshift.io/v1
kind: Etcd
metadata:
  name: cluster
spec:
  managementState: Managed
  # TODO this clearly needs to be fixed
  imagePullSpec: openshift/origin-hypershift:latest
`)

func v420EtcdOperatorConfigYamlBytes() ([]byte, error) {
	return _v420EtcdOperatorConfigYaml, nil
}

func v420EtcdOperatorConfigYaml() (*asset, error) {
	bytes, err := v420EtcdOperatorConfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/operator-config.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdPodCmYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: openshift-etcd
  name: etcd-pod
data:
  pod.yaml:
  forceRedeploymentReason:
  version:
`)

func v420EtcdPodCmYamlBytes() ([]byte, error) {
	return _v420EtcdPodCmYaml, nil
}

func v420EtcdPodCmYaml() (*asset, error) {
	bytes, err := v420EtcdPodCmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/pod-cm.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdPodYaml = []byte(`apiVersion: v1
kind: Pod
metadata:
  name: openshift-etcd
  namespace: openshift-etcd
  labels:
    app: etcd
    etcd: "true"
    revision: "REVISION"
spec:
  containers:
  - name: openshift-etcd
    image: ${IMAGE}
    imagePullPolicy: Always
    terminationMessagePolicy: FallbackToLogsOnError
    command: ["sleep"]
    args:
    - "7200"
    resources:
      requests:
        memory: 200Mi
        cpu: 100m
    ports:
      - containerPort: 10257
    volumeMounts:
    - mountPath: /etc/kubernetes/static-pod-resources
      name: resource-dir
#    livenessProbe:
#      httpGet:
#        scheme: HTTPS
#        port: 10257
#        path: healthz
#      initialDelaySeconds: 45
#      timeoutSeconds: 10
#    readinessProbe:
#      httpGet:
#        scheme: HTTPS
#        port: 10257
#        path: healthz
#      initialDelaySeconds: 10
#      timeoutSeconds: 10
  hostNetwork: true
  priorityClassName: system-node-critical
  tolerations:
  - operator: "Exists"
  volumes:
  - hostPath:
      path: /etc/kubernetes/static-pod-resources/etcd-pod-REVISION
    name: resource-dir

`)

func v420EtcdPodYamlBytes() ([]byte, error) {
	return _v420EtcdPodYaml, nil
}

func v420EtcdPodYaml() (*asset, error) {
	bytes, err := v420EtcdPodYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/pod.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdSaYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: openshift-etcd
  name: etcd-sa
`)

func v420EtcdSaYamlBytes() ([]byte, error) {
	return _v420EtcdSaYaml, nil
}

func v420EtcdSaYaml() (*asset, error) {
	bytes, err := v420EtcdSaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/sa.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _v420EtcdSvcYaml = []byte(`apiVersion: v1
kind: Service
metadata:
  namespace: openshift-etcd
  name: etcd
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: serving-cert
    prometheus.io/scrape: "true"
    prometheus.io/scheme: https
spec:
  selector:
    etcd: "true"
  ports:
  - name: https
    port: 443
    targetPort: 10257
`)

func v420EtcdSvcYamlBytes() ([]byte, error) {
	return _v420EtcdSvcYaml, nil
}

func v420EtcdSvcYaml() (*asset, error) {
	bytes, err := v420EtcdSvcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "v4.2.0/etcd/svc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"v4.2.0/etcd/cm.yaml":              v420EtcdCmYaml,
	"v4.2.0/etcd/defaultconfig.yaml":   v420EtcdDefaultconfigYaml,
	"v4.2.0/etcd/ns.yaml":              v420EtcdNsYaml,
	"v4.2.0/etcd/operator-config.yaml": v420EtcdOperatorConfigYaml,
	"v4.2.0/etcd/pod-cm.yaml":          v420EtcdPodCmYaml,
	"v4.2.0/etcd/pod.yaml":             v420EtcdPodYaml,
	"v4.2.0/etcd/sa.yaml":              v420EtcdSaYaml,
	"v4.2.0/etcd/svc.yaml":             v420EtcdSvcYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"v4.2.0": {nil, map[string]*bintree{
		"etcd": {nil, map[string]*bintree{
			"cm.yaml":              {v420EtcdCmYaml, map[string]*bintree{}},
			"defaultconfig.yaml":   {v420EtcdDefaultconfigYaml, map[string]*bintree{}},
			"ns.yaml":              {v420EtcdNsYaml, map[string]*bintree{}},
			"operator-config.yaml": {v420EtcdOperatorConfigYaml, map[string]*bintree{}},
			"pod-cm.yaml":          {v420EtcdPodCmYaml, map[string]*bintree{}},
			"pod.yaml":             {v420EtcdPodYaml, map[string]*bintree{}},
			"sa.yaml":              {v420EtcdSaYaml, map[string]*bintree{}},
			"svc.yaml":             {v420EtcdSvcYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
