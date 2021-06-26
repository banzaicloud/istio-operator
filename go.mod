module github.com/banzaicloud/istio-operator

go 1.14

require (
	emperror.dev/errors v0.8.0
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/istio-client-go v0.0.9
	github.com/banzaicloud/k8s-objectmatcher v1.5.0
	github.com/banzaicloud/operator-tools v0.21.3
	github.com/caddyserver/caddy v1.0.5
	github.com/coreos/go-semver v0.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
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
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.6.2
)
