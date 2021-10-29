module github.com/rancher/os2

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/apiserver => k8s.io/apiserver v0.22.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.2
	k8s.io/code-generator => k8s.io/code-generator v0.22.2
	k8s.io/component-base => k8s.io/component-base v0.22.2
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.2
	k8s.io/controller-manager => k8s.io/controller-manager v0.22.2
	k8s.io/cri-api => k8s.io/cri-api v0.22.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.2
	k8s.io/kubectl => k8s.io/kubectl v0.22.2
	k8s.io/kubelet => k8s.io/kubelet v0.22.2
	k8s.io/kubernetes => k8s.io/kubernetes v1.22.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.2
	k8s.io/metrics => k8s.io/metrics v0.22.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.22.2
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.2
)

require (
	github.com/google/certificate-transparency-go v1.1.2
	github.com/google/go-attestation v0.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-isatty v0.0.12
	github.com/pin/tftp v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher/fleet/pkg/apis v0.0.0-20210927195558-4aaa778d23dd
	github.com/rancher/lasso v0.0.0-20210709145333-6c6cd7fd6607
	github.com/rancher/rancher/pkg/apis v0.0.0-20211013185633-a636bda2a00e
	github.com/rancher/rancherd v0.0.1-alpha9.0.20211028172625-bdf5642d62d5
	github.com/rancher/steve v0.0.0-20210922195510-7224dc21013d
	github.com/rancher/system-upgrade-controller/pkg/apis v0.0.0-20210929162341-5e6e996d9486
	github.com/rancher/wrangler v0.8.7
	github.com/sirupsen/logrus v1.8.1
	github.com/tredoe/osutil v1.0.5
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	gopkg.in/pin/tftp.v2 v2.1.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0-beta.0
	sigs.k8s.io/yaml v1.2.0
)
