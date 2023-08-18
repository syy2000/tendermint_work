// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tendermint/types/txBlock.proto

package types

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	crypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	io "io"
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

type PoHBlock struct {
	Height        int64           `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	PoHTimestamps []*PoHTimestamp `protobuf:"bytes,2,rep,name=poHTimestamps,proto3" json:"poHTimestamps,omitempty"`
	Signature     []byte          `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	Address       []byte          `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *PoHBlock) Reset()         { *m = PoHBlock{} }
func (m *PoHBlock) String() string { return proto.CompactTextString(m) }
func (*PoHBlock) ProtoMessage()    {}
func (*PoHBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_09a99439714e6ae8, []int{0}
}
func (m *PoHBlock) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PoHBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PoHBlock.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PoHBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoHBlock.Merge(m, src)
}
func (m *PoHBlock) XXX_Size() int {
	return m.Size()
}
func (m *PoHBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_PoHBlock.DiscardUnknown(m)
}

var xxx_messageInfo_PoHBlock proto.InternalMessageInfo

func (m *PoHBlock) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *PoHBlock) GetPoHTimestamps() []*PoHTimestamp {
	if m != nil {
		return m.PoHTimestamps
	}
	return nil
}

func (m *PoHBlock) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *PoHBlock) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

type PoHBlockPart struct {
	Height  int64         `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Total   uint32        `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	Index   uint32        `protobuf:"varint,3,opt,name=index,proto3" json:"index,omitempty"`
	Bytes   []byte        `protobuf:"bytes,4,opt,name=bytes,proto3" json:"bytes,omitempty"`
	Proof   *crypto.Proof `protobuf:"bytes,5,opt,name=proof,proto3" json:"proof,omitempty"`
	Address []byte        `protobuf:"bytes,6,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *PoHBlockPart) Reset()         { *m = PoHBlockPart{} }
func (m *PoHBlockPart) String() string { return proto.CompactTextString(m) }
func (*PoHBlockPart) ProtoMessage()    {}
func (*PoHBlockPart) Descriptor() ([]byte, []int) {
	return fileDescriptor_09a99439714e6ae8, []int{1}
}
func (m *PoHBlockPart) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PoHBlockPart) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PoHBlockPart.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PoHBlockPart) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoHBlockPart.Merge(m, src)
}
func (m *PoHBlockPart) XXX_Size() int {
	return m.Size()
}
func (m *PoHBlockPart) XXX_DiscardUnknown() {
	xxx_messageInfo_PoHBlockPart.DiscardUnknown(m)
}

var xxx_messageInfo_PoHBlockPart proto.InternalMessageInfo

func (m *PoHBlockPart) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *PoHBlockPart) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *PoHBlockPart) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *PoHBlockPart) GetBytes() []byte {
	if m != nil {
		return m.Bytes
	}
	return nil
}

func (m *PoHBlockPart) GetProof() *crypto.Proof {
	if m != nil {
		return m.Proof
	}
	return nil
}

func (m *PoHBlockPart) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func init() {
	proto.RegisterType((*PoHBlock)(nil), "tendermint.types.PoHBlock")
	proto.RegisterType((*PoHBlockPart)(nil), "tendermint.types.PoHBlockPart")
}

func init() { proto.RegisterFile("tendermint/types/txBlock.proto", fileDescriptor_09a99439714e6ae8) }

var fileDescriptor_09a99439714e6ae8 = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x51, 0xc1, 0x4e, 0x02, 0x31,
	0x14, 0xa4, 0x20, 0xa8, 0x05, 0x12, 0xb3, 0x31, 0xa6, 0x21, 0xda, 0x6c, 0x38, 0xed, 0xa9, 0x9b,
	0xe0, 0xc1, 0x3b, 0xf1, 0xc0, 0x71, 0x6d, 0x3c, 0x79, 0x5b, 0xd8, 0x0a, 0x8d, 0xec, 0xb6, 0x69,
	0x1f, 0x09, 0xfc, 0x85, 0x9f, 0xe0, 0x37, 0xf8, 0x15, 0x1e, 0x39, 0x7a, 0x34, 0xf0, 0x23, 0x86,
	0x56, 0xdc, 0x85, 0xe8, 0x6d, 0x67, 0xe6, 0xbd, 0x9d, 0xe9, 0x1b, 0x4c, 0x41, 0x14, 0x99, 0x30,
	0xb9, 0x2c, 0x20, 0x86, 0x95, 0x16, 0x36, 0x86, 0xe5, 0x70, 0xae, 0x26, 0x2f, 0x4c, 0x1b, 0x05,
	0x2a, 0xb8, 0x28, 0x75, 0xe6, 0xf4, 0x5e, 0xff, 0x8f, 0x8d, 0x47, 0x99, 0x0b, 0x0b, 0x69, 0xae,
	0xfd, 0x56, 0xef, 0xa6, 0x32, 0x33, 0x31, 0x2b, 0x0d, 0x2a, 0xd6, 0x46, 0xa9, 0x67, 0x2f, 0xf7,
	0xdf, 0x10, 0x3e, 0x4b, 0xd4, 0xc8, 0xf9, 0x04, 0x57, 0xb8, 0x35, 0x13, 0x72, 0x3a, 0x03, 0x82,
	0x42, 0x14, 0x35, 0xf8, 0x0f, 0x0a, 0xee, 0x71, 0x57, 0xab, 0xd1, 0xef, 0x9f, 0x2d, 0xa9, 0x87,
	0x8d, 0xa8, 0x3d, 0xa0, 0xec, 0x38, 0x11, 0x4b, 0x2a, 0x63, 0xfc, 0x70, 0x29, 0xb8, 0xc6, 0xe7,
	0x56, 0x4e, 0x8b, 0x14, 0x16, 0x46, 0x90, 0x46, 0x88, 0xa2, 0x0e, 0x2f, 0x89, 0x80, 0xe0, 0xd3,
	0x34, 0xcb, 0x8c, 0xb0, 0x96, 0x9c, 0x38, 0x6d, 0x0f, 0xfb, 0xef, 0x08, 0x77, 0xf6, 0x11, 0x93,
	0xd4, 0xc0, 0xbf, 0x31, 0x2f, 0x71, 0x13, 0x14, 0xa4, 0x73, 0x52, 0x0f, 0x51, 0xd4, 0xe5, 0x1e,
	0xec, 0x58, 0x59, 0x64, 0x62, 0xe9, 0x2c, 0xbb, 0xdc, 0x83, 0x1d, 0x3b, 0x5e, 0x81, 0xd8, 0x9b,
	0x79, 0x10, 0x30, 0xdc, 0x74, 0xc7, 0x21, 0xcd, 0x10, 0x45, 0xed, 0x01, 0xa9, 0x3e, 0xd0, 0x1f,
	0x8f, 0x25, 0x3b, 0x9d, 0xfb, 0xb1, 0x6a, 0xe8, 0xd6, 0x41, 0xe8, 0xe1, 0xc3, 0xc7, 0x86, 0xa2,
	0xf5, 0x86, 0xa2, 0xaf, 0x0d, 0x45, 0xaf, 0x5b, 0x5a, 0x5b, 0x6f, 0x69, 0xed, 0x73, 0x4b, 0x6b,
	0x4f, 0x77, 0x53, 0x09, 0xb3, 0xc5, 0x98, 0x4d, 0x54, 0x1e, 0x57, 0xfb, 0x2b, 0x3f, 0x5d, 0x33,
	0xf1, 0x71, 0xb7, 0xe3, 0x96, 0xe3, 0x6f, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x31, 0x69, 0x95,
	0x36, 0x28, 0x02, 0x00, 0x00,
}

func (m *PoHBlock) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PoHBlock) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PoHBlock) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintTxBlock(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintTxBlock(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.PoHTimestamps) > 0 {
		for iNdEx := len(m.PoHTimestamps) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.PoHTimestamps[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTxBlock(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Height != 0 {
		i = encodeVarintTxBlock(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *PoHBlockPart) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PoHBlockPart) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PoHBlockPart) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintTxBlock(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x32
	}
	if m.Proof != nil {
		{
			size, err := m.Proof.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTxBlock(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Bytes) > 0 {
		i -= len(m.Bytes)
		copy(dAtA[i:], m.Bytes)
		i = encodeVarintTxBlock(dAtA, i, uint64(len(m.Bytes)))
		i--
		dAtA[i] = 0x22
	}
	if m.Index != 0 {
		i = encodeVarintTxBlock(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x18
	}
	if m.Total != 0 {
		i = encodeVarintTxBlock(dAtA, i, uint64(m.Total))
		i--
		dAtA[i] = 0x10
	}
	if m.Height != 0 {
		i = encodeVarintTxBlock(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintTxBlock(dAtA []byte, offset int, v uint64) int {
	offset -= sovTxBlock(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PoHBlock) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTxBlock(uint64(m.Height))
	}
	if len(m.PoHTimestamps) > 0 {
		for _, e := range m.PoHTimestamps {
			l = e.Size()
			n += 1 + l + sovTxBlock(uint64(l))
		}
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovTxBlock(uint64(l))
	}
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovTxBlock(uint64(l))
	}
	return n
}

func (m *PoHBlockPart) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTxBlock(uint64(m.Height))
	}
	if m.Total != 0 {
		n += 1 + sovTxBlock(uint64(m.Total))
	}
	if m.Index != 0 {
		n += 1 + sovTxBlock(uint64(m.Index))
	}
	l = len(m.Bytes)
	if l > 0 {
		n += 1 + l + sovTxBlock(uint64(l))
	}
	if m.Proof != nil {
		l = m.Proof.Size()
		n += 1 + l + sovTxBlock(uint64(l))
	}
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovTxBlock(uint64(l))
	}
	return n
}

func sovTxBlock(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTxBlock(x uint64) (n int) {
	return sovTxBlock(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PoHBlock) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTxBlock
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
			return fmt.Errorf("proto: PoHBlock: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PoHBlock: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoHTimestamps", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
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
				return ErrInvalidLengthTxBlock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTxBlock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PoHTimestamps = append(m.PoHTimestamps, &PoHTimestamp{})
			if err := m.PoHTimestamps[len(m.PoHTimestamps)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTxBlock
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTxBlock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTxBlock
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTxBlock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = append(m.Address[:0], dAtA[iNdEx:postIndex]...)
			if m.Address == nil {
				m.Address = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTxBlock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTxBlock
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PoHBlockPart) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTxBlock
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
			return fmt.Errorf("proto: PoHBlockPart: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PoHBlockPart: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Total", wireType)
			}
			m.Total = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Total |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Index |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bytes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTxBlock
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTxBlock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bytes = append(m.Bytes[:0], dAtA[iNdEx:postIndex]...)
			if m.Bytes == nil {
				m.Bytes = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proof", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
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
				return ErrInvalidLengthTxBlock
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTxBlock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Proof == nil {
				m.Proof = &crypto.Proof{}
			}
			if err := m.Proof.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTxBlock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTxBlock
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTxBlock
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = append(m.Address[:0], dAtA[iNdEx:postIndex]...)
			if m.Address == nil {
				m.Address = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTxBlock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTxBlock
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTxBlock(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTxBlock
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
					return 0, ErrIntOverflowTxBlock
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
					return 0, ErrIntOverflowTxBlock
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
				return 0, ErrInvalidLengthTxBlock
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTxBlock
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTxBlock
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTxBlock        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTxBlock          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTxBlock = fmt.Errorf("proto: unexpected end of group")
)
