// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api/v1alpha1/meshgateway.proto

package v1alpha1

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	_ "istio.io/api/mesh/v1alpha1"
	_ "istio.io/gogo-genproto/googleapis/google/api"
	v1 "k8s.io/api/core/v1"
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
	// Contains the intended version for the Mesh Gateway.
	// +kubebuilder:validation:Pattern=^1.
	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version"`
	// MeshGatewayConfiguration `json:",inline"`
	// +kubebuilder:validation:MinItems=0
	Ports                []*v1.ServicePort `protobuf:"bytes,2,rep,name=Ports,proto3" json:"ports"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
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

func (m *MeshGatewaySpec) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *MeshGatewaySpec) GetPorts() []*v1.ServicePort {
	if m != nil {
		return m.Ports
	}
	return nil
}

// <!-- go code generation tags
// +genclient
// +k8s:deepcopy-gen=true
// -->
type MeshGatewayStatus struct {
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

func init() {
	proto.RegisterType((*MeshGatewaySpec)(nil), "istio_operator.v2.api.v1alpha1.MeshGatewaySpec")
	proto.RegisterType((*MeshGatewayStatus)(nil), "istio_operator.v2.api.v1alpha1.MeshGatewayStatus")
}

func init() { proto.RegisterFile("api/v1alpha1/meshgateway.proto", fileDescriptor_fdd75a369aea761c) }

var fileDescriptor_fdd75a369aea761c = []byte{
	// 311 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xb1, 0x4e, 0xf3, 0x30,
	0x14, 0x85, 0x95, 0xfe, 0xea, 0x8f, 0x1a, 0x86, 0x8a, 0xc2, 0x50, 0x75, 0x48, 0xaa, 0x4e, 0x65,
	0xc0, 0x56, 0x8a, 0x90, 0x18, 0xa1, 0x0c, 0x4c, 0x48, 0xa8, 0xdd, 0x58, 0x2a, 0xc7, 0xbd, 0x75,
	0x2c, 0xd2, 0x5c, 0xcb, 0x76, 0x8c, 0x60, 0xe0, 0xf9, 0x18, 0x79, 0x82, 0x0a, 0x65, 0xec, 0x53,
	0x20, 0x27, 0x8d, 0xe8, 0x76, 0x74, 0x75, 0x3e, 0x9f, 0x7b, 0x8f, 0xc3, 0x88, 0x29, 0x49, 0x5d,
	0xc2, 0x72, 0x95, 0xb1, 0x84, 0x6e, 0xc1, 0x64, 0x82, 0x59, 0x78, 0x63, 0xef, 0x44, 0x69, 0xb4,
	0x38, 0x88, 0xa4, 0xb1, 0x12, 0x57, 0xa8, 0x40, 0x33, 0x8b, 0x9a, 0xb8, 0x19, 0x61, 0x4a, 0x92,
	0x96, 0x18, 0x8d, 0x3c, 0xf2, 0xf7, 0x00, 0xc7, 0x62, 0x23, 0x45, 0xc3, 0x8e, 0x2e, 0x04, 0x0a,
	0xac, 0x25, 0xf5, 0xea, 0x30, 0x8d, 0x05, 0xa2, 0xc8, 0x81, 0xfa, 0xe0, 0x8d, 0x84, 0x7c, 0xbd,
	0x4a, 0x21, 0x63, 0x4e, 0xa2, 0x3e, 0x18, 0x26, 0xaf, 0xb7, 0x86, 0x48, 0xac, 0x0d, 0x1c, 0x35,
	0x50, 0x97, 0x50, 0x01, 0x85, 0x5f, 0x00, 0xd6, 0x8d, 0x67, 0xf2, 0x19, 0xf6, 0x9f, 0xc0, 0x64,
	0x8f, 0xcd, 0xae, 0x4b, 0x05, 0x7c, 0x70, 0x19, 0x9e, 0x38, 0xd0, 0x46, 0x62, 0x31, 0x0c, 0xc6,
	0xc1, 0xb4, 0x37, 0xef, 0x57, 0xf7, 0x41, 0x67, 0xbf, 0x8b, 0xdb, 0xf1, 0xa2, 0x15, 0x83, 0xbb,
	0xb0, 0xfb, 0x8c, 0xda, 0x9a, 0x61, 0x67, 0xfc, 0x6f, 0x7a, 0x3a, 0x8b, 0x49, 0x93, 0x58, 0x5f,
	0xe6, 0x13, 0x89, 0x4b, 0xc8, 0x12, 0xb4, 0x93, 0x1c, 0xbc, 0x6f, 0xde, 0xdb, 0xef, 0xe2, 0xae,
	0xf2, 0xc4, 0xa2, 0x01, 0x27, 0xe7, 0xe1, 0xd9, 0x71, 0xbe, 0x65, 0xb6, 0x34, 0xf3, 0x87, 0xaf,
	0x2a, 0x0a, 0xbe, 0xab, 0x28, 0xf8, 0xa9, 0xa2, 0xe0, 0xe5, 0x46, 0x48, 0x9b, 0x95, 0x29, 0xe1,
	0xb8, 0xa5, 0x29, 0x2b, 0x3e, 0x98, 0xe4, 0x39, 0x96, 0x6b, 0x5a, 0x17, 0x7a, 0xd5, 0x16, 0x4a,
	0xdd, 0x8c, 0x1e, 0x7f, 0x41, 0xfa, 0xbf, 0x3e, 0xf0, 0xfa, 0x37, 0x00, 0x00, 0xff, 0xff, 0x8d,
	0xf6, 0x2c, 0x16, 0x99, 0x01, 0x00, 0x00,
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
	if len(m.Ports) > 0 {
		for iNdEx := len(m.Ports) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Ports[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMeshgateway(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Version) > 0 {
		i -= len(m.Version)
		copy(dAtA[i:], m.Version)
		i = encodeVarintMeshgateway(dAtA, i, uint64(len(m.Version)))
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
	l = len(m.Version)
	if l > 0 {
		n += 1 + l + sovMeshgateway(uint64(l))
	}
	if len(m.Ports) > 0 {
		for _, e := range m.Ports {
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
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
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
			m.Version = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ports", wireType)
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
			m.Ports = append(m.Ports, &v1.ServicePort{})
			if err := m.Ports[len(m.Ports)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
