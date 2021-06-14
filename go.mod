module github.com/banzaicloud/istio-operator

go 1.16

require (
	emperror.dev/errors v0.8.0
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/istio-client-go v0.0.9
	github.com/banzaicloud/istio-operator/pkg/apis v0.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.5.0
	github.com/caddyserver/caddy v1.0.5
	github.com/coreos/go-semver v0.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/goph/emperror v0.17.2
	github.com/lestrrat-go/jwx v1.0.6
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200627165143-92b8a710ab6c
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/banzaicloud/istio-operator/pkg/apis => ./pkg/apis
