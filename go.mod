module github.com/rancher/os2

go 1.16

require (
	github.com/mattn/go-isatty v0.0.12
	github.com/pin/tftp v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher/fleet/pkg/apis v0.0.0-20210927195558-4aaa778d23dd
	github.com/rancher/lasso v0.0.0-20210709145333-6c6cd7fd6607
	github.com/rancher/rancher/pkg/apis v0.0.0-20211013185633-a636bda2a00e
	github.com/rancher/steve v0.0.0-20210922195510-7224dc21013d
	github.com/rancher/system-upgrade-controller/pkg/apis v0.0.0-20210929162341-5e6e996d9486
	github.com/rancher/wrangler v0.8.7
	github.com/sirupsen/logrus v1.7.0
	github.com/tredoe/osutil v1.0.5
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	gopkg.in/pin/tftp.v2 v2.1.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0-beta.0
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client-go => k8s.io/client-go v0.22.2
