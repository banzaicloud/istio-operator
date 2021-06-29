module github.com/banzaicloud/istio-operator/v2

go 1.15

require (
	emperror.dev/errors v0.8.0
	github.com/banzaicloud/istio-operator/v2/api v0.0.1
	github.com/banzaicloud/istio-operator/v2/static v0.0.1
	github.com/banzaicloud/operator-tools v0.21.2-0.20210422182251-13765652229a
	github.com/go-logr/logr v0.4.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	istio.io/client-go v1.10.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	github.com/banzaicloud/istio-operator/v2/api => ./api
	github.com/banzaicloud/istio-operator/v2/static => ./static
)
