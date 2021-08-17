module github.com/banzaicloud/istio-operator/v2

go 1.16

require (
	cloud.google.com/go v0.60.0 // indirect
	emperror.dev/errors v0.8.0
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/banzaicloud/istio-operator/v2/api v0.0.1
	github.com/banzaicloud/k8s-objectmatcher v1.5.1
	github.com/banzaicloud/operator-tools v0.23.5-0.20210817100953-3db992cc06ee
	github.com/fatih/color v1.12.0 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/gonvenience/ytbx v1.4.2
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/homeport/dyff v1.4.3
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0
	github.com/lib/pq v1.9.0 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.opencensus.io v0.22.4 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sys v0.0.0-20210603125802-9665404d3644 // indirect
	golang.org/x/tools v0.1.3 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	honnef.co/go/tools v0.2.0 // indirect
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

	// needs a fork to support int64/uint64 marshalling to integers
	github.com/gogo/protobuf => github.com/waynz0r/protobuf v1.3.3-0.20210811122234-64636cae0910
)
