/*
Copyright 2020 Banzai Cloud.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RemoteIstioSpec defines the desired state of RemoteIstio
type RemoteIstioSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of RemoteIstio. Edit RemoteIstio_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// RemoteIstioStatus defines the observed state of RemoteIstio
type RemoteIstioStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// RemoteIstio is the Schema for the remoteistios API
type RemoteIstio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteIstioSpec   `json:"spec,omitempty"`
	Status RemoteIstioStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RemoteIstioList contains a list of RemoteIstio
type RemoteIstioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteIstio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteIstio{}, &RemoteIstioList{})
}
