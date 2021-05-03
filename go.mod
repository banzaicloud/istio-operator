module github.com/banzaicloud/istio-operator/v2

go 1.15

require (
	github.com/banzaicloud/istio-operator/v2/api v0.0.1
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/banzaicloud/istio-operator/v2/api => ./api
