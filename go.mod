module github.com/burmilla/os

go 1.26.0

replace github.com/cloudfoundry-incubator/candiedyaml => github.com/burmilla/candiedyaml v0.0.0-20190123020550-457a86017e98

replace github.com/codegangsta/cli => github.com/burmilla/cli-1 v1.18.1-0.20160809210721-d2b9ba9c38eb

replace github.com/docker/containerd => github.com/burmilla/containerd v0.2.1-0.20210501144658-c5695fa78c8c

replace github.com/docker/docker => github.com/burmilla/docker v0.0.0-20210501162625-1447a2463bc3

replace github.com/opencontainers/runc => github.com/burmilla/runc v0.0.5-0.20160531034625-edc34c4a8c1e

replace github.com/vishvananda/netlink => github.com/burmilla/netlink v0.0.0-20191127081807-b76f71f1d337

replace github.com/docker/distribution => ./third_party/docker/distribution

require (
	github.com/SvenDowideit/cpuid v0.0.0-20170710000921-dfdb6dba69f4
	github.com/cloudfoundry-incubator/candiedyaml v0.0.0-00010101000000-000000000000
	github.com/codegangsta/cli v0.0.0-00010101000000-000000000000
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/engine-api v0.3.3
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.5.0
	github.com/docker/libnetwork v0.5.6
	github.com/docker/machine v0.3.0-rc1.0.20150728163540-4a8e93ac9bc2
	github.com/fatih/structs v0.0.0-20160807235529-dc3312cb1a45
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/j-keck/arping v0.0.0-20150107081615-4f4d2c8983a1
	github.com/packethost/packngo v0.1.0
	github.com/pin/tftp v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/ryanuber/go-glob v0.0.0-20140617171557-0067a9abd927
	github.com/sigma/vmw-guestinfo v0.0.0-20160204083807-95dd4126d6e8
	github.com/sigma/vmw-ovflib v0.0.0-20150909143614-a99a06f1158f
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.4
	github.com/tredoe/term v0.0.0-20161130133337-e551c64f56c0
	github.com/vishvananda/netlink v0.0.0-00010101000000-000000000000
)

require (
	github.com/Microsoft/go-winio v0.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/docker/containerd v0.0.0-00010101000000-000000000000 // indirect
	github.com/docker/distribution v0.0.0-00010101000000-000000000000 // indirect
	github.com/gorilla/context v0.0.0-20140604161150-14f550f51af5 // indirect
	github.com/gorilla/mux v0.0.0-20140926153814-e444e69cbd2e // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v0.0.0-00010101000000-000000000000 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sigma/bdoor v0.0.0-20160624225336-b9c82b7b3c0b // indirect
	github.com/vbatts/tar-split v0.9.11 // indirect
	github.com/vishvananda/netns v0.0.0-20170219233438-54f0e4339ce7 // indirect
	github.com/vmware/vmw-guestinfo v0.0.0-20220317130741-510905f0efa3 // indirect
	golang.org/x/sync v0.3.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/compose-spec/compose-go v1.20.2
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/onsi/ginkgo v1.2.0 // indirect
	github.com/onsi/gomega v1.1.0 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/smartystreets/goconvey v1.6.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
	golang.org/x/net v0.0.0-20210428185706-aea814203247 // indirect
	golang.org/x/sys v0.1.0 // indirect
)

replace github.com/docker/libnetwork => ./third_party/docker/libnetwork

replace github.com/xeipuuv/gojsonschema => github.com/xeipuuv/gojsonschema v0.0.0-20170528113821-0c8571ac0ce1

replace github.com/xeipuuv/gojsonpointer => github.com/xeipuuv/gojsonpointer v0.0.0-20170225233418-6fe8760cad35

replace github.com/xeipuuv/gojsonreference => github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c

replace github.com/docker/go-connections => ./third_party/docker/go-connections
