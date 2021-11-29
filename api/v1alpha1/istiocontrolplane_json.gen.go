// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api/v1alpha1/istiocontrolplane.proto

package v1alpha1

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/waynz0r/protobuf/gogoproto"
	github_com_waynz0r_protobuf_jsonpb "github.com/waynz0r/protobuf/jsonpb"
	proto "github.com/waynz0r/protobuf/proto"
	_ "github.com/waynz0r/protobuf/types"
	_ "istio.io/api/mesh/v1alpha1"
	_ "istio.io/gogo-genproto/googleapis/google/api"
	_ "k8s.io/api/core/v1"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// MarshalJSON is a custom marshaler for IstioControlPlaneSpec
func (this *IstioControlPlaneSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioControlPlaneSpec
func (this *IstioControlPlaneSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for SidecarInjectorConfiguration
func (this *SidecarInjectorConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for SidecarInjectorConfiguration
func (this *SidecarInjectorConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MeshExpansionConfiguration
func (this *MeshExpansionConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MeshExpansionConfiguration
func (this *MeshExpansionConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MeshExpansionConfiguration_Istiod
func (this *MeshExpansionConfiguration_Istiod) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MeshExpansionConfiguration_Istiod
func (this *MeshExpansionConfiguration_Istiod) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MeshExpansionConfiguration_Webhook
func (this *MeshExpansionConfiguration_Webhook) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MeshExpansionConfiguration_Webhook
func (this *MeshExpansionConfiguration_Webhook) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MeshExpansionConfiguration_ClusterServices
func (this *MeshExpansionConfiguration_ClusterServices) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MeshExpansionConfiguration_ClusterServices
func (this *MeshExpansionConfiguration_ClusterServices) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MeshExpansionConfiguration_IstioMeshGatewayConfiguration
func (this *MeshExpansionConfiguration_IstioMeshGatewayConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MeshExpansionConfiguration_IstioMeshGatewayConfiguration
func (this *MeshExpansionConfiguration_IstioMeshGatewayConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for LoggingConfiguration
func (this *LoggingConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for LoggingConfiguration
func (this *LoggingConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for SDSConfiguration
func (this *SDSConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for SDSConfiguration
func (this *SDSConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ProxyConfiguration
func (this *ProxyConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ProxyConfiguration
func (this *ProxyConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ProxyInitConfiguration
func (this *ProxyInitConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ProxyInitConfiguration
func (this *ProxyInitConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIConfiguration
func (this *CNIConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIConfiguration
func (this *CNIConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIConfiguration_RepairConfiguration
func (this *CNIConfiguration_RepairConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIConfiguration_RepairConfiguration
func (this *CNIConfiguration_RepairConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIConfiguration_TaintConfiguration
func (this *CNIConfiguration_TaintConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIConfiguration_TaintConfiguration
func (this *CNIConfiguration_TaintConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIConfiguration_ResourceQuotas
func (this *CNIConfiguration_ResourceQuotas) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIConfiguration_ResourceQuotas
func (this *CNIConfiguration_ResourceQuotas) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstiodConfiguration
func (this *IstiodConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstiodConfiguration
func (this *IstiodConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ExternalIstiodConfiguration
func (this *ExternalIstiodConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ExternalIstiodConfiguration
func (this *ExternalIstiodConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for SPIFFEConfiguration
func (this *SPIFFEConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for SPIFFEConfiguration
func (this *SPIFFEConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for OperatorEndpointsConfiguration
func (this *OperatorEndpointsConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for OperatorEndpointsConfiguration
func (this *OperatorEndpointsConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TelemetryV2Configuration
func (this *TelemetryV2Configuration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TelemetryV2Configuration
func (this *TelemetryV2Configuration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ProxyWasmConfiguration
func (this *ProxyWasmConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ProxyWasmConfiguration
func (this *ProxyWasmConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for PDBConfiguration
func (this *PDBConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for PDBConfiguration
func (this *PDBConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for HTTPProxyEnvsConfiguration
func (this *HTTPProxyEnvsConfiguration) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for HTTPProxyEnvsConfiguration
func (this *HTTPProxyEnvsConfiguration) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioControlPlaneStatus
func (this *IstioControlPlaneStatus) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioControlPlaneStatus
func (this *IstioControlPlaneStatus) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for StatusChecksums
func (this *StatusChecksums) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for StatusChecksums
func (this *StatusChecksums) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

var (
	IstiocontrolplaneMarshaler   = &github_com_waynz0r_protobuf_jsonpb.Marshaler{Int64Uint64asIntegers: true}
	IstiocontrolplaneUnmarshaler = &github_com_waynz0r_protobuf_jsonpb.Unmarshaler{AllowUnknownFields: true}
)
