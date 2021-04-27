module github.com/banzaicloud/istio-operator/v2

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	istio.io/api v0.0.0-20210406181827-2b71efe58165
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
