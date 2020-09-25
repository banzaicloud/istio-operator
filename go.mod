module github.com/banzaicloud/istio-operator

go 1.14

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/istio-client-go v0.0.9
	github.com/banzaicloud/k8s-objectmatcher v1.4.0
	github.com/coreos/go-semver v0.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/goph/emperror v0.17.2
	github.com/mholt/caddy v1.0.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.8.1
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200627165143-92b8a710ab6c
	github.com/stretchr/testify v1.6.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools v0.0.0-20200616195046-dc31b401abb5 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200603094226-e3079894b1e8 // indirect
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.4.0 // indirect
)
