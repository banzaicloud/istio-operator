module github.com/banzaicloud/istio-operator/v2

go 1.16

require (
	emperror.dev/errors v0.8.0
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/banzaicloud/istio-operator/v2/api v0.0.1
	github.com/banzaicloud/k8s-objectmatcher v1.5.1
	github.com/banzaicloud/operator-tools v0.23.5-0.20210809115043-5a718e66d1bf
	github.com/go-logr/logr v0.4.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	golang.org/x/tools v0.1.3 // indirect
	istio.io/client-go v1.10.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/banzaicloud/istio-operator/v2/api => ./api
	github.com/banzaicloud/istio-operator/v2/static => ./static
)
