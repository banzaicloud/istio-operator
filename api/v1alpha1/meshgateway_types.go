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

const (
	SidecarInjectionChecksumAnnotation = "sidecar.istio.servicemesh.cisco.com/injection-checksum"
	MeshConfigChecksumAnnotation       = "sidecar.istio.servicemesh.cisco.com/meshconfig-checksum"
)

// +kubebuilder:object:root=true

// MeshGateway is the Schema for the meshgateways API
type MeshGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   *MeshGatewaySpec  `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status MeshGatewayStatus `json:"status,omitempty"`
}

func (mgw *MeshGateway) SetStatus(status ConfigState, errorMessage string) {
	mgw.Status.Status = status
	mgw.Status.ErrorMessage = errorMessage
}

func (mgw *MeshGateway) GetStatus() MeshGatewayStatus {
	return mgw.Status
}

func (mgw *MeshGateway) GetSpec() *MeshGatewaySpec {
	if mgw.Spec != nil {
		return mgw.Spec
	}

	return nil
}

type MeshGatewayWithProperties struct {
	*MeshGateway
	Properties MeshGatewayProperties
}

func (p *MeshGatewayWithProperties) SetDefaults() {
	annotations := p.MeshGateway.GetSpec().GetDeployment().GetPodMetadata().GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if p.Properties.InjectionChecksum != "" {
		annotations[SidecarInjectionChecksumAnnotation] = p.Properties.InjectionChecksum
	}
	if p.Properties.MeshConfigChecksum != "" {
		annotations[MeshConfigChecksumAnnotation] = p.Properties.MeshConfigChecksum
	}
	if p.MeshGateway.GetSpec().GetDeployment() == nil {
		p.MeshGateway.GetSpec().Deployment = &BaseKubernetesResourceConfig{}
	}
	if p.MeshGateway.GetSpec().GetDeployment().GetPodMetadata() == nil {
		p.MeshGateway.GetSpec().GetDeployment().PodMetadata = &K8SObjectMeta{}
	}
	p.MeshGateway.GetSpec().GetDeployment().GetPodMetadata().Annotations = annotations
}

type MeshGatewayProperties struct {
	Revision              string
	EnablePrometheusMerge bool
	InjectionTemplate     string
	InjectionChecksum     string
	MeshConfigChecksum    string
}

// +kubebuilder:object:root=true

// MeshGatewayList contains a list of MeshGateway
type MeshGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []MeshGateway `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&MeshGateway{}, &MeshGatewayList{})
}
