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

type SortableIstioControlPlaneItems []IstioControlPlane

func (list SortableIstioControlPlaneItems) Len() int {
	return len(list)
}

func (list SortableIstioControlPlaneItems) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableIstioControlPlaneItems) Less(i, j int) bool {
	return list[i].CreationTimestamp.Time.Before(list[j].CreationTimestamp.Time)
}

// +kubebuilder:object:root=true

// IstioControlPlane is the Schema for the istiocontrolplanes API
// +kubebuilder:resource:path=istiocontrolplanes,shortName=icp;istiocp
type IstioControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *IstioControlPlaneSpec  `json:"spec,omitempty"`
	Status IstioControlPlaneStatus `json:"status,omitempty"`
}

func (icp *IstioControlPlane) GetSpec() *IstioControlPlaneSpec {
	if icp.Spec != nil {
		return icp.Spec
	}

	return nil
}

// +kubebuilder:object:root=true

// IstioControlPlaneList contains a list of IstioControlPlane
type IstioControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IstioControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IstioControlPlane{}, &IstioControlPlaneList{})
}
