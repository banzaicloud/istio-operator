module github.com/banzaicloud/istio-operator/v2

go 1.15

require (
	emperror.dev/errors v0.8.0
	github.com/banzaicloud/istio-operator/v2/api v0.0.1
	github.com/banzaicloud/operator-tools v0.21.2-0.20210422182251-13765652229a
	github.com/go-logr/logr v0.4.0
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546
	golang.org/x/mod v0.4.0 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/banzaicloud/istio-operator/v2/api => ./api
