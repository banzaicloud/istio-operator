/*
Copyright 2021 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import "k8s.io/apimachinery/pkg/runtime/schema"

var serviceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "services",
}

var podGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "pods",
}

var deploymentGVR = schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "deployments",
}

var horizontalPodAutoscalerGVR = schema.GroupVersionResource{
	Group:    "autoscaling",
	Version:  "v1",
	Resource: "horizontalpodautoscalers",
}

var clusterRoleGVR = schema.GroupVersionResource{
	Group:    "rbac.authorization.k8s.io",
	Version:  "v1",
	Resource: "clusterroles",
}

var clusterRoleBindingGVR = schema.GroupVersionResource{
	Group:    "rbac.authorization.k8s.io",
	Version:  "v1",
	Resource: "clusterrolebindings",
}

var validatingWebhookConfigurationGVR = schema.GroupVersionResource{
	Group:    "admissionregistration.k8s.io",
	Version:  "v1",
	Resource: "validatingwebhookconfigurations",
}

var mutatingWebhookconfigurationGVR = schema.GroupVersionResource{
	Group:    "admissionregistration.k8s.io",
	Version:  "v1",
	Resource: "mutatingwebhookconfigurations",
}

var istioGVR = schema.GroupVersionResource{
	Group:    "istio.banzaicloud.io",
	Version:  "v1beta1",
	Resource: "istios",
}

var meshGatewayGVR = schema.GroupVersionResource{
	Group:    "istio.banzaicloud.io",
	Version:  "v1beta1",
	Resource: "meshgateways",
}
