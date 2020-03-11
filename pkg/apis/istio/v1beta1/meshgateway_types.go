/*
Copyright 2019 Banzai Cloud.

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatewayType string

const (
	GatewayTypeIngress GatewayType = "ingress"
	GatewayTypeEgress  GatewayType = "egress"
)

// MeshGatewaySpec defines the desired state of MeshGateway
type MeshGatewaySpec struct {
	MeshGatewayConfiguration `json:",inline"`
	// +kubebuilder:validation:MinItems=1
	Ports []corev1.ServicePort `json:"ports"`
	Type  GatewayType          `json:"type"`
}

type MeshGatewayConfiguration struct {
	BaseK8sResourceConfigurationWithHPAWithoutImage `json:",inline"`
	Labels                                          map[string]string `json:"labels,omitempty"`
	// +kubebuilder:validation:Enum=ClusterIP,NodePort,LoadBalancer
	ServiceType          corev1.ServiceType      `json:"serviceType,omitempty"`
	LoadBalancerIP       string                  `json:"loadBalancerIP,omitempty"`
	ServiceAnnotations   map[string]string       `json:"serviceAnnotations,omitempty"`
	ServiceLabels        map[string]string       `json:"serviceLabels,omitempty"`
	SDS                  GatewaySDSConfiguration `json:"sds,omitempty"`
	ApplicationPorts     string                  `json:"applicationPorts,omitempty"`
	RequestedNetworkView string                  `json:"requestedNetworkView,omitempty"`
	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

// MeshGatewayStatus defines the observed state of MeshGateway
type MeshGatewayStatus struct {
	Status         ConfigState `json:"Status,omitempty"`
	GatewayAddress []string    `json:"GatewayAddress,omitempty"`
	ErrorMessage   string      `json:"ErrorMessage,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeshGateway is the Schema for the meshgateways API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="Type of the gateway"
// +kubebuilder:printcolumn:name="Service Type",type="string",JSONPath=".spec.serviceType",description="Type of the service"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
// +kubebuilder:printcolumn:name="Ingress IPs",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateway addresses of the resource"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=meshgateways,shortName=mgw
type MeshGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshGatewaySpec   `json:"spec,omitempty"`
	Status MeshGatewayStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeshGatewayList contains a list of MeshGateway
type MeshGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeshGateway{}, &MeshGatewayList{})
}
