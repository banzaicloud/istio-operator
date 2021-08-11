// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api/v1alpha1/meshgateway.proto

package v1alpha1

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	io "io"
	_ "istio.io/gogo-genproto/googleapis/google/api"
	_ "k8s.io/api/core/v1"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type GatewayType int32

const (
	GatewayType_unspecified GatewayType = 0
	GatewayType_ingress     GatewayType = 1
	GatewayType_egress      GatewayType = 2
)

var GatewayType_name = map[int32]string{
	0: "unspecified",
	1: "ingress",
	2: "egress",
}

var GatewayType_value = map[string]int32{
	"unspecified": 0,
	"ingress":     1,
	"egress":      2,
}

func (x GatewayType) String() string {
	return proto.EnumName(GatewayType_name, int32(x))
}

func (GatewayType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fdd75a369aea761c, []int{0}
}

type ConfigState int32

const (
	ConfigState_Unspecified     ConfigState = 0
	ConfigState_Created         ConfigState = 1
	ConfigState_ReconcileFailed ConfigState = 2
	ConfigState_Reconciling     ConfigState = 3
	ConfigState_Available       ConfigState = 4
	ConfigState_Unmanaged       ConfigState = 5
)

var ConfigState_name = map[int32]string{
	0: "Unspecified",
	1: "Created",
	2: "ReconcileFailed",
	3: "Reconciling",
	4: "Available",
	5: "Unmanaged",
}

var ConfigState_value = map[string]int32{
	"Unspecified":     0,
	"Created":         1,
	"ReconcileFailed": 2,
	"Reconciling":     3,
	"Available":       4,
	"Unmanaged":       5,
}

func (x ConfigState) String() string {
	return proto.EnumName(ConfigState_name, int32(x))
}

func (ConfigState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fdd75a369aea761c, []int{1}
}

// MeshGateway defines an Istio ingress or egress gateway
//
// <!-- crd generation tags
// +cue-gen:MeshGateway:groupName:servicemesh.cisco.com
// +cue-gen:MeshGateway:version:v1alpha1
// +cue-gen:MeshGateway:storageVersion
// +cue-gen:MeshGateway:annotations:helm.sh/resource-policy=keep
// +cue-gen:MeshGateway:subresource:status
// +cue-gen:MeshGateway:scope:Namespaced
// +cue-gen:MeshGateway:resource:shortNames=mgw,meshgw
// +cue-gen:MeshGateway:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp
// representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations.
// Clients may not set this value. It is represented in RFC3339 form and is in UTC.
// Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
// +cue-gen:MeshGateway:preserveUnknownFields:false
// -->
//
// <!-- go code generation tags
// +genclient
// +k8s:deepcopy-gen=true
// -->
type MeshGatewaySpec struct {
	// Deployment spec
	Deployment *BaseKubernetesResourceConfig `protobuf:"bytes,1,opt,name=deployment,proto3" json:"deployment,omitempty"`
	// Service spec
	Service *Service `protobuf:"bytes,2,opt,name=service,proto3" json:"service,omitempty"`
	// Whether to run the gateway in a privileged container
	RunAsRoot *bool `protobuf:"bytes,3,opt,name=runAsRoot,proto3,wktptr" json:"runAsRoot,omitempty"`
	// Type of gateway, either ingress or egress
	// +kubebuilder:validation:Enum=ingress;egress
	Type GatewayType `protobuf:"varint,4,opt,name=type,proto3,enum=istio_operator.v2.api.v1alpha1.GatewayType" json:"type,omitempty"`
	// Istio CR to which this gateway belongs to
	IstioControlPlane *NamespacedName `protobuf:"bytes,5,opt,name=istioControlPlane,proto3" json:"istioControlPlane,omitempty"`
	// K8s resource overlay patches
	K8SResourceOverlays  []*K8SResourceOverlayPatch `protobuf:"bytes,6,rep,name=k8sResourceOverlays,proto3" json:"k8sResourceOverlays,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                   `json:"-"`
	XXX_unrecognized     []byte                     `json:"-"`
	XXX_sizecache        int32                      `json:"-"`
}

func (m *MeshGatewaySpec) Reset()         { *m = MeshGatewaySpec{} }
func (m *MeshGatewaySpec) String() string { return proto.CompactTextString(m) }
func (*MeshGatewaySpec) ProtoMessage()    {}
func (*MeshGatewaySpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_fdd75a369aea761c, []int{0}
}
func (m *MeshGatewaySpec) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MeshGatewaySpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MeshGatewaySpec.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MeshGatewaySpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MeshGatewaySpec.Merge(m, src)
}
func (m *MeshGatewaySpec) XXX_Size() int {
	return m.Size()
}
func (m *MeshGatewaySpec) XXX_DiscardUnknown() {
	xxx_messageInfo_MeshGatewaySpec.DiscardUnknown(m)
}

var xxx_messageInfo_MeshGatewaySpec proto.InternalMessageInfo

func (m *MeshGatewaySpec) GetDeployment() *BaseKubernetesResourceConfig {
	if m != nil {
		return m.Deployment
	}
	return nil
}

func (m *MeshGatewaySpec) GetService() *Service {
	if m != nil {
		return m.Service
	}
	return nil
}

func (m *MeshGatewaySpec) GetRunAsRoot() *bool {
	if m != nil {
		return m.RunAsRoot
	}
	return nil
}

func (m *MeshGatewaySpec) GetType() GatewayType {
	if m != nil {
		return m.Type
	}
	return GatewayType_unspecified
}

func (m *MeshGatewaySpec) GetIstioControlPlane() *NamespacedName {
	if m != nil {
		return m.IstioControlPlane
	}
	return nil
}

func (m *MeshGatewaySpec) GetK8SResourceOverlays() []*K8SResourceOverlayPatch {
	if m != nil {
		return m.K8SResourceOverlays
	}
	return nil
}

// <!-- go code generation tags
// +genclient
// +k8s:deepcopy-gen=true
// -->
type MeshGatewayStatus struct {
	// Reconciliation status of the mesh gateway
	Status ConfigState `protobuf:"varint,1,opt,name=Status,proto3,enum=istio_operator.v2.api.v1alpha1.ConfigState" json:"Status,omitempty"`
	// Current address for the gateway
	GatewayAddress []string `protobuf:"bytes,2,rep,name=GatewayAddress,proto3" json:"GatewayAddress,omitempty"`
	// Reconciliation error message if any
	ErrorMessage         string   `protobuf:"bytes,3,opt,name=ErrorMessage,proto3" json:"ErrorMessage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MeshGatewayStatus) Reset()         { *m = MeshGatewayStatus{} }
func (m *MeshGatewayStatus) String() string { return proto.CompactTextString(m) }
func (*MeshGatewayStatus) ProtoMessage()    {}
func (*MeshGatewayStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_fdd75a369aea761c, []int{1}
}
func (m *MeshGatewayStatus) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MeshGatewayStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MeshGatewayStatus.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MeshGatewayStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MeshGatewayStatus.Merge(m, src)
}
func (m *MeshGatewayStatus) XXX_Size() int {
	return m.Size()
}
func (m *MeshGatewayStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_MeshGatewayStatus.DiscardUnknown(m)
}

var xxx_messageInfo_MeshGatewayStatus proto.InternalMessageInfo

func (m *MeshGatewayStatus) GetStatus() ConfigState {
	if m != nil {
		return m.Status
	}
	return ConfigState_Unspecified
}

func (m *MeshGatewayStatus) GetGatewayAddress() []string {
	if m != nil {
		return m.GatewayAddress
	}
	return nil
}

func (m *MeshGatewayStatus) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func init() {
	proto.RegisterEnum("istio_operator.v2.api.v1alpha1.GatewayType", GatewayType_name, GatewayType_value)
	proto.RegisterEnum("istio_operator.v2.api.v1alpha1.ConfigState", ConfigState_name, ConfigState_value)
	proto.RegisterType((*MeshGatewaySpec)(nil), "istio_operator.v2.api.v1alpha1.MeshGatewaySpec")
	proto.RegisterType((*MeshGatewayStatus)(nil), "istio_operator.v2.api.v1alpha1.MeshGatewayStatus")
}

func init() { proto.RegisterFile("api/v1alpha1/meshgateway.proto", fileDescriptor_fdd75a369aea761c) }

var fileDescriptor_fdd75a369aea761c = []byte{
	// 617 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xcf, 0x6e, 0xd3, 0x4e,
	0x10, 0xc7, 0x7f, 0x4e, 0xd2, 0x54, 0xd9, 0xfc, 0x68, 0xdd, 0x2d, 0x07, 0xd3, 0x43, 0x1a, 0xe5,
	0x00, 0x51, 0x11, 0xb6, 0x12, 0x84, 0xda, 0x03, 0x42, 0x4a, 0xa2, 0xc2, 0xa1, 0x2a, 0x54, 0x2e,
	0xe5, 0x80, 0x90, 0xaa, 0xb5, 0x3d, 0x71, 0x56, 0x5d, 0xef, 0x58, 0xbb, 0xb6, 0xab, 0xf0, 0x04,
	0x3c, 0x06, 0x07, 0x1e, 0x86, 0x23, 0x6f, 0x00, 0xca, 0x93, 0x20, 0xff, 0x09, 0xa4, 0x2d, 0x22,
	0xdc, 0xc6, 0xe3, 0xf9, 0x7c, 0x77, 0xf6, 0x3b, 0x63, 0x93, 0x0e, 0x8b, 0xb9, 0x93, 0x0d, 0x98,
	0x88, 0x67, 0x6c, 0xe0, 0x44, 0xa0, 0x67, 0x21, 0x4b, 0xe0, 0x9a, 0xcd, 0xed, 0x58, 0x61, 0x82,
	0xb4, 0xc3, 0x75, 0xc2, 0xf1, 0x12, 0x63, 0x50, 0x2c, 0x41, 0x65, 0x67, 0x43, 0x9b, 0xc5, 0xdc,
	0x5e, 0x12, 0x7b, 0x9d, 0x10, 0x31, 0x14, 0xe0, 0x14, 0xd5, 0x5e, 0x3a, 0x75, 0xae, 0x15, 0x8b,
	0x63, 0x50, 0xba, 0xe4, 0xf7, 0x1e, 0xdc, 0xd0, 0xf7, 0x31, 0x8a, 0x50, 0x56, 0xaf, 0xee, 0x87,
	0x18, 0x62, 0x11, 0x3a, 0x79, 0x54, 0x65, 0xf7, 0x2b, 0xc1, 0x9c, 0x9b, 0x72, 0x10, 0xc1, 0xa5,
	0x07, 0x33, 0x96, 0x71, 0x54, 0x55, 0x41, 0xef, 0xea, 0x48, 0xdb, 0x1c, 0x8b, 0x02, 0x1f, 0x15,
	0x38, 0xd9, 0xc0, 0x09, 0x41, 0xe6, 0xfd, 0x41, 0x50, 0xd6, 0xf4, 0x3e, 0x35, 0xc8, 0xf6, 0x29,
	0xe8, 0xd9, 0xab, 0xf2, 0x2e, 0xe7, 0x31, 0xf8, 0xf4, 0x03, 0x21, 0x01, 0xc4, 0x02, 0xe7, 0x11,
	0xc8, 0xc4, 0x32, 0xba, 0x46, 0xbf, 0x3d, 0x7c, 0x6e, 0xff, 0xfd, 0x7a, 0xf6, 0x98, 0x69, 0x38,
	0x49, 0x3d, 0x50, 0x12, 0x12, 0xd0, 0x2e, 0x68, 0x4c, 0x95, 0x0f, 0x13, 0x94, 0x53, 0x1e, 0xba,
	0x2b, 0x7a, 0x74, 0x44, 0x36, 0x35, 0xa8, 0x8c, 0xfb, 0x60, 0xd5, 0x0a, 0xe9, 0x47, 0xeb, 0xa4,
	0xcf, 0xcb, 0x72, 0x77, 0xc9, 0xd1, 0x17, 0xa4, 0xa5, 0x52, 0x39, 0xd2, 0x2e, 0x62, 0x62, 0xd5,
	0x0b, 0x91, 0x3d, 0xbb, 0x74, 0xc3, 0x5e, 0xda, 0x6b, 0x8f, 0x11, 0xc5, 0x3b, 0x26, 0x52, 0x18,
	0x37, 0x3e, 0x7f, 0xdf, 0x37, 0xdc, 0xdf, 0x08, 0x3d, 0x26, 0x8d, 0x64, 0x1e, 0x83, 0xd5, 0xe8,
	0x1a, 0xfd, 0xad, 0xe1, 0xe3, 0x75, 0xe7, 0x57, 0xde, 0xbc, 0x9d, 0xc7, 0x30, 0x6e, 0x2c, 0x46,
	0x46, 0xcd, 0x2d, 0x70, 0xea, 0x91, 0x9d, 0x82, 0x9c, 0xa0, 0x4c, 0x14, 0x8a, 0x33, 0xc1, 0x24,
	0x58, 0x1b, 0x45, 0x3b, 0xf6, 0x3a, 0xcd, 0xd7, 0x2c, 0x02, 0x1d, 0x33, 0x1f, 0x82, 0x3c, 0xaa,
	0x64, 0xef, 0xca, 0x51, 0x4e, 0x76, 0xaf, 0x8e, 0x7e, 0xd9, 0xf9, 0x26, 0x03, 0x25, 0xd8, 0x5c,
	0x5b, 0xcd, 0x6e, 0xbd, 0xdf, 0x1e, 0x1e, 0xae, 0x3b, 0xe5, 0xe4, 0x0e, 0x7a, 0xc6, 0x12, 0x7f,
	0xe6, 0xfe, 0x49, 0xb3, 0xf7, 0xc5, 0x20, 0x3b, 0xab, 0xab, 0x90, 0xb0, 0x24, 0xd5, 0x74, 0x42,
	0x9a, 0x65, 0x54, 0x2c, 0xc2, 0x3f, 0xb8, 0x55, 0x8e, 0x3c, 0x67, 0xc0, 0xad, 0x50, 0xfa, 0x90,
	0x6c, 0x55, 0xaa, 0xa3, 0x20, 0x50, 0xa0, 0xb5, 0x55, 0xeb, 0xd6, 0xfb, 0x2d, 0xf7, 0x56, 0x96,
	0xf6, 0xc8, 0xff, 0xc7, 0x4a, 0xa1, 0x3a, 0x05, 0xad, 0x59, 0x08, 0xc5, 0x6c, 0x5b, 0xee, 0x8d,
	0xdc, 0xc1, 0x21, 0x69, 0xaf, 0x0c, 0x84, 0x6e, 0x93, 0x76, 0x2a, 0x75, 0x0c, 0x3e, 0x9f, 0x72,
	0x08, 0xcc, 0xff, 0x68, 0x9b, 0x6c, 0x72, 0x19, 0xe6, 0x72, 0xa6, 0x41, 0x09, 0x69, 0x42, 0x19,
	0xd7, 0x0e, 0x90, 0xb4, 0x57, 0x7a, 0xcb, 0xc1, 0x8b, 0xdb, 0xe0, 0x44, 0x41, 0xfe, 0x6d, 0x98,
	0x06, 0xdd, 0x25, 0xdb, 0x2e, 0xf8, 0x28, 0x7d, 0x2e, 0xe0, 0x25, 0xe3, 0x02, 0x02, 0xb3, 0x96,
	0x23, 0xcb, 0x24, 0x97, 0xa1, 0x59, 0xa7, 0xf7, 0x48, 0x6b, 0x94, 0x31, 0x2e, 0x98, 0x27, 0xc0,
	0x6c, 0xe4, 0x8f, 0x17, 0x32, 0x62, 0x92, 0x85, 0x10, 0x98, 0x1b, 0xe3, 0xc9, 0xd7, 0x45, 0xc7,
	0xf8, 0xb6, 0xe8, 0x18, 0x3f, 0x16, 0x1d, 0xe3, 0xfd, 0xb3, 0x90, 0x27, 0xb3, 0xd4, 0xb3, 0x7d,
	0x8c, 0x1c, 0x8f, 0xc9, 0x8f, 0x8c, 0xfb, 0x02, 0xd3, 0xc0, 0x29, 0xec, 0x7c, 0xb2, 0xb4, 0xd3,
	0xc9, 0x86, 0xce, 0xea, 0x8f, 0xc0, 0x6b, 0x16, 0x0b, 0xfd, 0xf4, 0x67, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x78, 0xfd, 0xf8, 0xc7, 0x7f, 0x04, 0x00, 0x00,
}

func (m *MeshGatewaySpec) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MeshGatewaySpec) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MeshGatewaySpec) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.K8SResourceOverlays) > 0 {
		for iNdEx := len(m.K8SResourceOverlays) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.K8SResourceOverlays[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMeshgateway(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if m.IstioControlPlane != nil {
		{
			size, err := m.IstioControlPlane.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintMeshgateway(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if m.Type != 0 {
		i = encodeVarintMeshgateway(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x20
	}
	if m.RunAsRoot != nil {
		n2, err2 := github_com_gogo_protobuf_types.StdBoolMarshalTo(*m.RunAsRoot, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdBool(*m.RunAsRoot):])
		if err2 != nil {
			return 0, err2
		}
		i -= n2
		i = encodeVarintMeshgateway(dAtA, i, uint64(n2))
		i--
		dAtA[i] = 0x1a
	}
	if m.Service != nil {
		{
			size, err := m.Service.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintMeshgateway(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.Deployment != nil {
		{
			size, err := m.Deployment.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintMeshgateway(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MeshGatewayStatus) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MeshGatewayStatus) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MeshGatewayStatus) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.ErrorMessage) > 0 {
		i -= len(m.ErrorMessage)
		copy(dAtA[i:], m.ErrorMessage)
		i = encodeVarintMeshgateway(dAtA, i, uint64(len(m.ErrorMessage)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.GatewayAddress) > 0 {
		for iNdEx := len(m.GatewayAddress) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.GatewayAddress[iNdEx])
			copy(dAtA[i:], m.GatewayAddress[iNdEx])
			i = encodeVarintMeshgateway(dAtA, i, uint64(len(m.GatewayAddress[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Status != 0 {
		i = encodeVarintMeshgateway(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintMeshgateway(dAtA []byte, offset int, v uint64) int {
	offset -= sovMeshgateway(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MeshGatewaySpec) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Deployment != nil {
		l = m.Deployment.Size()
		n += 1 + l + sovMeshgateway(uint64(l))
	}
	if m.Service != nil {
		l = m.Service.Size()
		n += 1 + l + sovMeshgateway(uint64(l))
	}
	if m.RunAsRoot != nil {
		l = github_com_gogo_protobuf_types.SizeOfStdBool(*m.RunAsRoot)
		n += 1 + l + sovMeshgateway(uint64(l))
	}
	if m.Type != 0 {
		n += 1 + sovMeshgateway(uint64(m.Type))
	}
	if m.IstioControlPlane != nil {
		l = m.IstioControlPlane.Size()
		n += 1 + l + sovMeshgateway(uint64(l))
	}
	if len(m.K8SResourceOverlays) > 0 {
		for _, e := range m.K8SResourceOverlays {
			l = e.Size()
			n += 1 + l + sovMeshgateway(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *MeshGatewayStatus) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Status != 0 {
		n += 1 + sovMeshgateway(uint64(m.Status))
	}
	if len(m.GatewayAddress) > 0 {
		for _, s := range m.GatewayAddress {
			l = len(s)
			n += 1 + l + sovMeshgateway(uint64(l))
		}
	}
	l = len(m.ErrorMessage)
	if l > 0 {
		n += 1 + l + sovMeshgateway(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovMeshgateway(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMeshgateway(x uint64) (n int) {
	return sovMeshgateway(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MeshGatewaySpec) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeshgateway
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MeshGatewaySpec: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MeshGatewaySpec: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Deployment", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Deployment == nil {
				m.Deployment = &BaseKubernetesResourceConfig{}
			}
			if err := m.Deployment.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Service", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Service == nil {
				m.Service = &Service{}
			}
			if err := m.Service.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RunAsRoot", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.RunAsRoot == nil {
				m.RunAsRoot = new(bool)
			}
			if err := github_com_gogo_protobuf_types.StdBoolUnmarshal(m.RunAsRoot, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= GatewayType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IstioControlPlane", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.IstioControlPlane == nil {
				m.IstioControlPlane = &NamespacedName{}
			}
			if err := m.IstioControlPlane.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field K8SResourceOverlays", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.K8SResourceOverlays = append(m.K8SResourceOverlays, &K8SResourceOverlayPatch{})
			if err := m.K8SResourceOverlays[len(m.K8SResourceOverlays)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMeshgateway(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MeshGatewayStatus) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeshgateway
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MeshGatewayStatus: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MeshGatewayStatus: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= ConfigState(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GatewayAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GatewayAddress = append(m.GatewayAddress, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ErrorMessage", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMeshgateway
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ErrorMessage = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMeshgateway(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMeshgateway
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMeshgateway(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMeshgateway
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMeshgateway
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthMeshgateway
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMeshgateway
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMeshgateway
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMeshgateway        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMeshgateway          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMeshgateway = fmt.Errorf("proto: unexpected end of group")
)