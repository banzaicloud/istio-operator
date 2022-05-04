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
	fmt "fmt"
	"strings"

	v1alpha1 "istio.io/api/mesh/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	RevisionedAutoInjectionLabel       = "istio.io/rev"
	DeprecatedAutoInjectionLabel       = "istio-injection"
	NamespaceInjectionSourceAnnotation = "controlplane.istio.servicemesh.cisco.com/namespace-injection-source"
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

func (icp *IstioControlPlane) SetStatus(status ConfigState, errorMessage string) {
	icp.Status.Status = status
	icp.Status.ErrorMessage = errorMessage
}

func (icp *IstioControlPlane) GetStatus() IstioControlPlaneStatus {
	return icp.Status
}

func (icp *IstioControlPlane) GetSpec() *IstioControlPlaneSpec {
	if icp.Spec != nil {
		return icp.Spec
	}

	return nil
}

func (icp *IstioControlPlane) Revision() string {
	return strings.ReplaceAll(icp.GetName(), ".", "-")
}

func (icp *IstioControlPlane) NamespacedRevision() string {
	return NamespacedRevision(icp.Revision(), icp.GetNamespace())
}

func (icp *IstioControlPlane) RevisionLabels() map[string]string {
	return map[string]string{
		RevisionedAutoInjectionLabel: icp.NamespacedRevision(),
	}
}

func (icp *IstioControlPlane) MeshExpansionGatewayLabels() map[string]string {
	return map[string]string{
		RevisionedAutoInjectionLabel: icp.NamespacedRevision(),
		"app":                        "istio-meshexpansion-gateway",
	}
}

func (icp *IstioControlPlane) WithRevision(s string) string {
	return fmt.Sprintf("%s-%s", s, icp.Revision())
}

func (icp *IstioControlPlane) WithRevisionIf(s string, condition bool) string {
	if !condition {
		return s
	}

	return icp.WithRevision(s)
}

func (icp *IstioControlPlane) WithNamespacedRevision(s string) string {
	return fmt.Sprintf("%s-%s", icp.WithRevision(s), icp.GetNamespace())
}

func NamespacedRevision(revision, namespace string) string {
	return fmt.Sprintf("%s.%s", revision, namespace)
}

func NamespacedNameFromRevision(revision string) types.NamespacedName {
	nn := types.NamespacedName{}
	p := strings.SplitN(revision, ".", 2)
	if len(p) == 2 {
		nn.Name = p[0]
		nn.Namespace = p[1]
	}

	return nn
}

// +kubebuilder:object:generate=false
type IstioControlPlaneWithProperties struct {
	*IstioControlPlane `json:"istioControlPlane,omitempty"`
	Properties         IstioControlPlaneProperties `json:"properties,omitempty"`
}

// Properties of the IstioControlPlane
type IstioControlPlaneProperties struct {
	Mesh                         *IstioMesh             `json:"mesh,omitempty"`
	MeshNetworks                 *v1alpha1.MeshNetworks `json:"meshNetworks,omitempty"`
	TrustedRootCACertificatePEMs []string               `json:"trustedRootCACertificatePEMs,omitempty"`
}

func (p IstioControlPlaneProperties) GetMesh() *IstioMesh {
	return p.Mesh
}

// +kubebuilder:object:root=true

// IstioControlPlaneList contains a list of IstioControlPlane
type IstioControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IstioControlPlane `json:"items"`
}

// PeerIstioControlPlane is the Schema for the clone of the istiocontrolplanes API
// +kubebuilder:object:root=true
type PeerIstioControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *IstioControlPlaneSpec  `json:"spec,omitempty"`
	Status IstioControlPlaneStatus `json:"status,omitempty"`
}

func (icp *PeerIstioControlPlane) GetStatus() IstioControlPlaneStatus {
	return icp.Status
}

func (icp *PeerIstioControlPlane) GetSpec() *IstioControlPlaneSpec {
	if icp.Spec != nil {
		return icp.Spec
	}

	return nil
}

func (r *ResourceRequirements) ConvertToK8sRR() *corev1.ResourceRequirements {
	rr := &corev1.ResourceRequirements{
		Limits:   make(corev1.ResourceList),
		Requests: make(corev1.ResourceList),
	}

	if r == nil {
		return rr
	}

	for k, v := range r.Limits {
		rr.Limits[corev1.ResourceName(k)] = v.Quantity
	}

	for k, v := range r.Requests {
		rr.Requests[corev1.ResourceName(k)] = v.Quantity
	}

	return rr
}

func InitResourceRequirementsFromK8sRR(rr *corev1.ResourceRequirements) *ResourceRequirements {
	r := &ResourceRequirements{
		Limits:   make(map[string]*Quantity),
		Requests: make(map[string]*Quantity),
	}

	if rr == nil {
		return r
	}

	for k, v := range rr.Limits {
		r.Limits[string(k)] = &Quantity{
			Quantity: v,
		}
	}

	for k, v := range rr.Requests {
		r.Requests[string(k)] = &Quantity{
			Quantity: v,
		}
	}

	return r
}

// PeerIstioControlPlaneList contains a list of PeerIstioControlPlane
// +kubebuilder:object:root=true
type PeerIstioControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PeerIstioControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IstioControlPlane{}, &IstioControlPlaneList{}, &PeerIstioControlPlane{}, &PeerIstioControlPlaneList{})
}
