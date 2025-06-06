// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.12.4
// source: protos/audio_convert.proto

package audioconverter

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AudioFormat int32

const (
	AudioFormat_AUDIO_FORMAT_UNSPECIFIED AudioFormat = 0
	AudioFormat_MP3                      AudioFormat = 1
	AudioFormat_M4A                      AudioFormat = 2
)

// Enum value maps for AudioFormat.
var (
	AudioFormat_name = map[int32]string{
		0: "AUDIO_FORMAT_UNSPECIFIED",
		1: "MP3",
		2: "M4A",
	}
	AudioFormat_value = map[string]int32{
		"AUDIO_FORMAT_UNSPECIFIED": 0,
		"MP3":                      1,
		"M4A":                      2,
	}
)

func (x AudioFormat) Enum() *AudioFormat {
	p := new(AudioFormat)
	*p = x
	return p
}

func (x AudioFormat) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AudioFormat) Descriptor() protoreflect.EnumDescriptor {
	return file_protos_audio_convert_proto_enumTypes[0].Descriptor()
}

func (AudioFormat) Type() protoreflect.EnumType {
	return &file_protos_audio_convert_proto_enumTypes[0]
}

func (x AudioFormat) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AudioFormat.Descriptor instead.
func (AudioFormat) EnumDescriptor() ([]byte, []int) {
	return file_protos_audio_convert_proto_rawDescGZIP(), []int{0}
}

type ConvertInput struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Format        AudioFormat            `protobuf:"varint,1,opt,name=format,proto3,enum=audio_convert.v1.AudioFormat" json:"format,omitempty"`
	Metadata      []string               `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty"`
	Data          []byte                 `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConvertInput) Reset() {
	*x = ConvertInput{}
	mi := &file_protos_audio_convert_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConvertInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConvertInput) ProtoMessage() {}

func (x *ConvertInput) ProtoReflect() protoreflect.Message {
	mi := &file_protos_audio_convert_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConvertInput.ProtoReflect.Descriptor instead.
func (*ConvertInput) Descriptor() ([]byte, []int) {
	return file_protos_audio_convert_proto_rawDescGZIP(), []int{0}
}

func (x *ConvertInput) GetFormat() AudioFormat {
	if x != nil {
		return x.Format
	}
	return AudioFormat_AUDIO_FORMAT_UNSPECIFIED
}

func (x *ConvertInput) GetMetadata() []string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *ConvertInput) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type ConvertReply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Data          []byte                 `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConvertReply) Reset() {
	*x = ConvertReply{}
	mi := &file_protos_audio_convert_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConvertReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConvertReply) ProtoMessage() {}

func (x *ConvertReply) ProtoReflect() protoreflect.Message {
	mi := &file_protos_audio_convert_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConvertReply.ProtoReflect.Descriptor instead.
func (*ConvertReply) Descriptor() ([]byte, []int) {
	return file_protos_audio_convert_proto_rawDescGZIP(), []int{1}
}

func (x *ConvertReply) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type StreamConvertInput struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Payload:
	//
	//	*StreamConvertInput_Metadata
	//	*StreamConvertInput_Chunk
	Payload       isStreamConvertInput_Payload `protobuf_oneof:"payload"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamConvertInput) Reset() {
	*x = StreamConvertInput{}
	mi := &file_protos_audio_convert_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamConvertInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamConvertInput) ProtoMessage() {}

func (x *StreamConvertInput) ProtoReflect() protoreflect.Message {
	mi := &file_protos_audio_convert_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamConvertInput.ProtoReflect.Descriptor instead.
func (*StreamConvertInput) Descriptor() ([]byte, []int) {
	return file_protos_audio_convert_proto_rawDescGZIP(), []int{2}
}

func (x *StreamConvertInput) GetPayload() isStreamConvertInput_Payload {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *StreamConvertInput) GetMetadata() *InitialMetadata {
	if x != nil {
		if x, ok := x.Payload.(*StreamConvertInput_Metadata); ok {
			return x.Metadata
		}
	}
	return nil
}

func (x *StreamConvertInput) GetChunk() []byte {
	if x != nil {
		if x, ok := x.Payload.(*StreamConvertInput_Chunk); ok {
			return x.Chunk
		}
	}
	return nil
}

type isStreamConvertInput_Payload interface {
	isStreamConvertInput_Payload()
}

type StreamConvertInput_Metadata struct {
	Metadata *InitialMetadata `protobuf:"bytes,1,opt,name=metadata,proto3,oneof"`
}

type StreamConvertInput_Chunk struct {
	Chunk []byte `protobuf:"bytes,2,opt,name=chunk,proto3,oneof"`
}

func (*StreamConvertInput_Metadata) isStreamConvertInput_Payload() {}

func (*StreamConvertInput_Chunk) isStreamConvertInput_Payload() {}

type InitialMetadata struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Format        AudioFormat            `protobuf:"varint,1,opt,name=format,proto3,enum=audio_convert.v1.AudioFormat" json:"format,omitempty"`
	Metadata      []string               `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InitialMetadata) Reset() {
	*x = InitialMetadata{}
	mi := &file_protos_audio_convert_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InitialMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitialMetadata) ProtoMessage() {}

func (x *InitialMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_protos_audio_convert_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitialMetadata.ProtoReflect.Descriptor instead.
func (*InitialMetadata) Descriptor() ([]byte, []int) {
	return file_protos_audio_convert_proto_rawDescGZIP(), []int{3}
}

func (x *InitialMetadata) GetFormat() AudioFormat {
	if x != nil {
		return x.Format
	}
	return AudioFormat_AUDIO_FORMAT_UNSPECIFIED
}

func (x *InitialMetadata) GetMetadata() []string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type StreamFileReply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Chunk         []byte                 `protobuf:"bytes,1,opt,name=chunk,proto3" json:"chunk,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamFileReply) Reset() {
	*x = StreamFileReply{}
	mi := &file_protos_audio_convert_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamFileReply) ProtoMessage() {}

func (x *StreamFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_protos_audio_convert_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamFileReply.ProtoReflect.Descriptor instead.
func (*StreamFileReply) Descriptor() ([]byte, []int) {
	return file_protos_audio_convert_proto_rawDescGZIP(), []int{4}
}

func (x *StreamFileReply) GetChunk() []byte {
	if x != nil {
		return x.Chunk
	}
	return nil
}

var File_protos_audio_convert_proto protoreflect.FileDescriptor

const file_protos_audio_convert_proto_rawDesc = "" +
	"\n" +
	"\x1aprotos/audio_convert.proto\x12\x10audio_convert.v1\"u\n" +
	"\fConvertInput\x125\n" +
	"\x06format\x18\x01 \x01(\x0e2\x1d.audio_convert.v1.AudioFormatR\x06format\x12\x1a\n" +
	"\bmetadata\x18\x02 \x03(\tR\bmetadata\x12\x12\n" +
	"\x04data\x18\x03 \x01(\fR\x04data\"\"\n" +
	"\fConvertReply\x12\x12\n" +
	"\x04data\x18\x01 \x01(\fR\x04data\"x\n" +
	"\x12StreamConvertInput\x12?\n" +
	"\bmetadata\x18\x01 \x01(\v2!.audio_convert.v1.InitialMetadataH\x00R\bmetadata\x12\x16\n" +
	"\x05chunk\x18\x02 \x01(\fH\x00R\x05chunkB\t\n" +
	"\apayload\"d\n" +
	"\x0fInitialMetadata\x125\n" +
	"\x06format\x18\x01 \x01(\x0e2\x1d.audio_convert.v1.AudioFormatR\x06format\x12\x1a\n" +
	"\bmetadata\x18\x02 \x03(\tR\bmetadata\"'\n" +
	"\x0fStreamFileReply\x12\x14\n" +
	"\x05chunk\x18\x01 \x01(\fR\x05chunk*=\n" +
	"\vAudioFormat\x12\x1c\n" +
	"\x18AUDIO_FORMAT_UNSPECIFIED\x10\x00\x12\a\n" +
	"\x03MP3\x10\x01\x12\a\n" +
	"\x03M4A\x10\x022\xb9\x01\n" +
	"\x0eAudioConverter\x12I\n" +
	"\aConvert\x12\x1e.audio_convert.v1.ConvertInput\x1a\x1e.audio_convert.v1.ConvertReply\x12\\\n" +
	"\rConvertStream\x12$.audio_convert.v1.StreamConvertInput\x1a!.audio_convert.v1.StreamFileReply(\x010\x01B#Z!gen/audioconverter;audioconverterb\x06proto3"

var (
	file_protos_audio_convert_proto_rawDescOnce sync.Once
	file_protos_audio_convert_proto_rawDescData []byte
)

func file_protos_audio_convert_proto_rawDescGZIP() []byte {
	file_protos_audio_convert_proto_rawDescOnce.Do(func() {
		file_protos_audio_convert_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protos_audio_convert_proto_rawDesc), len(file_protos_audio_convert_proto_rawDesc)))
	})
	return file_protos_audio_convert_proto_rawDescData
}

var file_protos_audio_convert_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protos_audio_convert_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_protos_audio_convert_proto_goTypes = []any{
	(AudioFormat)(0),           // 0: audio_convert.v1.AudioFormat
	(*ConvertInput)(nil),       // 1: audio_convert.v1.ConvertInput
	(*ConvertReply)(nil),       // 2: audio_convert.v1.ConvertReply
	(*StreamConvertInput)(nil), // 3: audio_convert.v1.StreamConvertInput
	(*InitialMetadata)(nil),    // 4: audio_convert.v1.InitialMetadata
	(*StreamFileReply)(nil),    // 5: audio_convert.v1.StreamFileReply
}
var file_protos_audio_convert_proto_depIdxs = []int32{
	0, // 0: audio_convert.v1.ConvertInput.format:type_name -> audio_convert.v1.AudioFormat
	4, // 1: audio_convert.v1.StreamConvertInput.metadata:type_name -> audio_convert.v1.InitialMetadata
	0, // 2: audio_convert.v1.InitialMetadata.format:type_name -> audio_convert.v1.AudioFormat
	1, // 3: audio_convert.v1.AudioConverter.Convert:input_type -> audio_convert.v1.ConvertInput
	3, // 4: audio_convert.v1.AudioConverter.ConvertStream:input_type -> audio_convert.v1.StreamConvertInput
	2, // 5: audio_convert.v1.AudioConverter.Convert:output_type -> audio_convert.v1.ConvertReply
	5, // 6: audio_convert.v1.AudioConverter.ConvertStream:output_type -> audio_convert.v1.StreamFileReply
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protos_audio_convert_proto_init() }
func file_protos_audio_convert_proto_init() {
	if File_protos_audio_convert_proto != nil {
		return
	}
	file_protos_audio_convert_proto_msgTypes[2].OneofWrappers = []any{
		(*StreamConvertInput_Metadata)(nil),
		(*StreamConvertInput_Chunk)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protos_audio_convert_proto_rawDesc), len(file_protos_audio_convert_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protos_audio_convert_proto_goTypes,
		DependencyIndexes: file_protos_audio_convert_proto_depIdxs,
		EnumInfos:         file_protos_audio_convert_proto_enumTypes,
		MessageInfos:      file_protos_audio_convert_proto_msgTypes,
	}.Build()
	File_protos_audio_convert_proto = out.File
	file_protos_audio_convert_proto_goTypes = nil
	file_protos_audio_convert_proto_depIdxs = nil
}
