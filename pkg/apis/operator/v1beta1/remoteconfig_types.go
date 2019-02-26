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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SignCert struct {
	CA    []byte
	Root  []byte
	Key   []byte
	Chain []byte
}

type IstioService struct {
	Name          string   `json:"name"`
	LabelSelector string   `json:"labelSelector"`
	IPs           []string `json:"podIPs,omitempty"`
}

// RemoteConfigSpec defines the desired state of RemoteConfig
type RemoteConfigSpec struct {
	IncludeIPRanges string         `json:"includeIPRanges,omitempty"`
	ExcludeIPRanges string         `json:"excludeIPRanges,omitempty"`
	EnabledServices []IstioService `json:"enabledServices"`
	// List of namespaces to label with sidecar auto injection enabled
	AutoInjectionNamespaces []string `json:"autoInjectionNamespaces,omitempty"`
	// ControlPlaneSecurityEnabled control plane services are communicating through mTLS
	ControlPlaneSecurityEnabled bool `json:"controlPlaneSecurityEnabled,omitempty"`

	signCert SignCert
}

func (spec RemoteConfigSpec) SetSignCert(signCert SignCert) RemoteConfigSpec {
	spec.signCert = signCert
	return spec
}

func (spec RemoteConfigSpec) GetSignCert() SignCert {
	return spec.signCert
}

// RemoteConfigStatus defines the observed state of RemoteConfig
type RemoteConfigStatus struct {
	Status       ConfigState
	ErrorMessage string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RemoteConfig is the Schema for the remoteconfigs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type RemoteConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteConfigSpec   `json:"spec,omitempty"`
	Status RemoteConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RemoteConfigList contains a list of RemoteConfig
type RemoteConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteConfig{}, &RemoteConfigList{})
}
