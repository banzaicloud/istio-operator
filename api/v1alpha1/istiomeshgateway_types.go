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

// IstioMeshGateway is the Schema for the istiomeshgateways API
type IstioMeshGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   *IstioMeshGatewaySpec  `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status IstioMeshGatewayStatus `json:"status,omitempty"`
}

func (imgw *IstioMeshGateway) SetStatus(status ConfigState, errorMessage string) {
	imgw.Status.Status = status
	imgw.Status.ErrorMessage = errorMessage
}

func (imgw *IstioMeshGateway) GetStatus() IstioMeshGatewayStatus {
	return imgw.Status
}

func (imgw *IstioMeshGateway) GetSpec() *IstioMeshGatewaySpec {
	if imgw.Spec != nil {
		return imgw.Spec
	}

	return nil
}

type IstioMeshGatewayWithProperties struct {
	*IstioMeshGateway
	Properties IstioMeshGatewayProperties
}

func (p *IstioMeshGatewayWithProperties) SetDefaults() {
	annotations := p.IstioMeshGateway.GetSpec().GetDeployment().GetPodMetadata().GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if p.Properties.InjectionChecksum != "" {
		annotations[SidecarInjectionChecksumAnnotation] = p.Properties.InjectionChecksum
	}
	if p.Properties.MeshConfigChecksum != "" {
		annotations[MeshConfigChecksumAnnotation] = p.Properties.MeshConfigChecksum
	}
	if p.IstioMeshGateway.GetSpec().GetDeployment() == nil {
		p.IstioMeshGateway.GetSpec().Deployment = &BaseKubernetesResourceConfig{}
	}
	if p.IstioMeshGateway.GetSpec().GetDeployment().GetPodMetadata() == nil {
		p.IstioMeshGateway.GetSpec().GetDeployment().PodMetadata = &K8SObjectMeta{}
	}
	p.IstioMeshGateway.GetSpec().GetDeployment().GetPodMetadata().Annotations = annotations
}

type IstioMeshGatewayProperties struct {
	Revision                string
	EnablePrometheusMerge   *bool
	InjectionTemplate       string
	InjectionChecksum       string
	MeshConfigChecksum      string
	IstioControlPlane       *IstioControlPlane
	GenerateExternalService bool
}

func (p IstioMeshGatewayProperties) GetIstioControlPlane() *IstioControlPlane {
	if p.IstioControlPlane != nil {
		return p.IstioControlPlane
	}

	return &IstioControlPlane{}
}

// +kubebuilder:object:root=true

// IstioMeshGatewayList contains a list of IstioMeshGateway
type IstioMeshGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []IstioMeshGateway `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&IstioMeshGateway{}, &IstioMeshGatewayList{})
}
