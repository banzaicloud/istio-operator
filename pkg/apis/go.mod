module github.com/banzaicloud/istio-operator/pkg/apis

go 1.16

require (
	github.com/banzaicloud/istio-client-go v0.0.9
	github.com/banzaicloud/operator-tools v0.21.1
	github.com/onsi/gomega v1.10.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.6.2
)
