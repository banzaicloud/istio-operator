/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// IstioMesh is the Schema for the mesh API
type IstioMesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   *IstioMeshSpec  `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status IstioMeshStatus `json:"status,omitempty"`
}

func (m *IstioMesh) SetStatus(status ConfigState, errorMessage string) {
	m.Status.Status = status
	m.Status.ErrorMessage = errorMessage
}

func (m *IstioMesh) GetStatus() IstioMeshStatus {
	return m.Status
}

func (m *IstioMesh) GetSpec() *IstioMeshSpec {
	if m.Spec != nil {
		return m.Spec
	}

	return nil
}

// +kubebuilder:object:root=true

// IstioMeshList contains a list of IstioMesh
type IstioMeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []IstioMesh `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&IstioMesh{}, &IstioMeshList{})
}
