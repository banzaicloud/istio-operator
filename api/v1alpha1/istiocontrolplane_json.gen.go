// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api/v1alpha1/istiocontrolplane.proto

package v1alpha1

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	github_com_gogo_protobuf_jsonpb "github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
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
	IstiocontrolplaneMarshaler   = &github_com_gogo_protobuf_jsonpb.Marshaler{Int64Uint64asIntegers: true}
	IstiocontrolplaneUnmarshaler = &github_com_gogo_protobuf_jsonpb.Unmarshaler{AllowUnknownFields: true}
)
