module github.com/banzaicloud/istio-operator/api/v2

go 1.18

require (
	github.com/golang/protobuf v1.5.2
	google.golang.org/genproto v0.0.0-20220628213854-d9e0b6570c03
	google.golang.org/protobuf v1.28.0
	istio.io/api v0.0.0-20220817131511-59047e057639
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	sigs.k8s.io/controller-runtime v0.12.3
)

require (
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

// needs a fork to support istio operator v2 api int64/uint64 marshalling to integers
replace github.com/golang/protobuf => github.com/luciferinlove/protobuf v0.0.0-20220913214010-c63936d75066
