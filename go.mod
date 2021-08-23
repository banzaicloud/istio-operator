module github.com/banzaicloud/istio-operator/v2

go 1.16

require (
	cloud.google.com/go v0.60.0 // indirect
	emperror.dev/errors v0.8.0
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/banzaicloud/istio-operator/v2/api v0.0.1
	github.com/banzaicloud/k8s-objectmatcher v1.5.2
	github.com/banzaicloud/operator-tools v0.24.1-0.20210823142428-2095ccb7f5e6
	github.com/fatih/color v1.12.0 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/golang/protobuf v1.5.2
	github.com/gonvenience/ytbx v1.4.2
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/homeport/dyff v1.4.3
	github.com/kylelemons/godebug v1.1.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	go.opencensus.io v0.22.4 // indirect
	go.uber.org/zap v1.18.1
	golang.org/x/tools v0.1.3 // indirect
	istio.io/client-go v1.11.0
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.5
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/banzaicloud/istio-operator/v2/api => ./api
	github.com/banzaicloud/istio-operator/v2/static => ./static

	// needs a fork to support int64/uint64 marshalling to integers
	github.com/gogo/protobuf => github.com/waynz0r/protobuf v1.3.3-0.20210811122234-64636cae0910
)
