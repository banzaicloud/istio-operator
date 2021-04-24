module github.com/banzaicloud/istio-operator/v2

go 1.15

require (
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	istio.io/api v0.0.0-20210318170531-e6e017e575c5
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.6.5
)
