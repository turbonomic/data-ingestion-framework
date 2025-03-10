// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v5.29.3
// source: IdentityMetadata.proto

package proto

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

// EntityIdentityMetadata supplies meta information describing the properties used to
// identify an entity of a specific entity type.
type EntityIdentityMetadata struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The EntityType this metadata is for
	EntityType *EntityDTO_EntityType `protobuf:"varint,1,req,name=entityType,enum=common_dto.EntityDTO_EntityType" json:"entityType,omitempty"`
	// The version of the identifying properties for this entity type
	Version *int32 `protobuf:"varint,2,opt,name=version,def=0" json:"version,omitempty"`
	// The non-volatile identifying properties to be used
	// for this entity type. non-volatile identifying properties are the set of properties
	// necessary to identify an entity that will not change over the lifetime of the entity.
	// For example, "ID" will be a non-volatile identifying property for most entity types.
	NonVolatileProperties []*EntityIdentityMetadata_PropertyMetadata `protobuf:"bytes,3,rep,name=nonVolatileProperties" json:"nonVolatileProperties,omitempty"`
	// The volatile identifying properties to be used for this entity type.
	// Volatile identifying properties are the set of properties necessary to identify
	// an entity that may change over the lifetime of the entity. For example, for a VM,
	// the "PM_UUID" may be identifying, but moving the VM will cause the value of this property
	// to change.
	VolatileProperties []*EntityIdentityMetadata_PropertyMetadata `protobuf:"bytes,4,rep,name=volatileProperties" json:"volatileProperties,omitempty"`
	// The heuristic properties to be used for this entity type. Heuristic properties
	// are used to fuzzy match an entity's identity when an exact match using the
	// identifying non-volatile and volatile properties fails.
	HeuristicProperties []*EntityIdentityMetadata_PropertyMetadata `protobuf:"bytes,5,rep,name=heuristicProperties" json:"heuristicProperties,omitempty"`
	// The heuristic threshold is used by the identity service when matching heuristic properties
	// to determine what percentage of heuristic properties must match in order to consider
	// two objects to be the same. A heuristicThreshold of 50 would mean that at least 1/2 of
	// the heuristic properties must match for two entities to be considered to be the same.
	// This must be a value between 0 and 100.
	HeuristicThreshold *int32 `protobuf:"varint,6,opt,name=heuristicThreshold,def=75" json:"heuristicThreshold,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

// Default values for EntityIdentityMetadata fields.
const (
	Default_EntityIdentityMetadata_Version            = int32(0)
	Default_EntityIdentityMetadata_HeuristicThreshold = int32(75)
)

func (x *EntityIdentityMetadata) Reset() {
	*x = EntityIdentityMetadata{}
	mi := &file_IdentityMetadata_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EntityIdentityMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntityIdentityMetadata) ProtoMessage() {}

func (x *EntityIdentityMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_IdentityMetadata_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntityIdentityMetadata.ProtoReflect.Descriptor instead.
func (*EntityIdentityMetadata) Descriptor() ([]byte, []int) {
	return file_IdentityMetadata_proto_rawDescGZIP(), []int{0}
}

func (x *EntityIdentityMetadata) GetEntityType() EntityDTO_EntityType {
	if x != nil && x.EntityType != nil {
		return *x.EntityType
	}
	return EntityDTO_SWITCH
}

func (x *EntityIdentityMetadata) GetVersion() int32 {
	if x != nil && x.Version != nil {
		return *x.Version
	}
	return Default_EntityIdentityMetadata_Version
}

func (x *EntityIdentityMetadata) GetNonVolatileProperties() []*EntityIdentityMetadata_PropertyMetadata {
	if x != nil {
		return x.NonVolatileProperties
	}
	return nil
}

func (x *EntityIdentityMetadata) GetVolatileProperties() []*EntityIdentityMetadata_PropertyMetadata {
	if x != nil {
		return x.VolatileProperties
	}
	return nil
}

func (x *EntityIdentityMetadata) GetHeuristicProperties() []*EntityIdentityMetadata_PropertyMetadata {
	if x != nil {
		return x.HeuristicProperties
	}
	return nil
}

func (x *EntityIdentityMetadata) GetHeuristicThreshold() int32 {
	if x != nil && x.HeuristicThreshold != nil {
		return *x.HeuristicThreshold
	}
	return Default_EntityIdentityMetadata_HeuristicThreshold
}

type EntityIdentityMetadata_PropertyMetadata struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the property.
	Name          *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EntityIdentityMetadata_PropertyMetadata) Reset() {
	*x = EntityIdentityMetadata_PropertyMetadata{}
	mi := &file_IdentityMetadata_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EntityIdentityMetadata_PropertyMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntityIdentityMetadata_PropertyMetadata) ProtoMessage() {}

func (x *EntityIdentityMetadata_PropertyMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_IdentityMetadata_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntityIdentityMetadata_PropertyMetadata.ProtoReflect.Descriptor instead.
func (*EntityIdentityMetadata_PropertyMetadata) Descriptor() ([]byte, []int) {
	return file_IdentityMetadata_proto_rawDescGZIP(), []int{0, 0}
}

func (x *EntityIdentityMetadata_PropertyMetadata) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

var File_IdentityMetadata_proto protoreflect.FileDescriptor

var file_IdentityMetadata_proto_rawDesc = string([]byte{
	0x0a, 0x16, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x5f, 0x64, 0x74, 0x6f, 0x1a, 0x0f, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x44, 0x54, 0x4f, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8a, 0x04, 0x0a, 0x16, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x40, 0x0a, 0x0a, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x02, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x64, 0x74,
	0x6f, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x44, 0x54, 0x4f, 0x2e, 0x45, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x1b, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x3a, 0x01, 0x30, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x69, 0x0a, 0x15, 0x6e, 0x6f, 0x6e, 0x56, 0x6f, 0x6c, 0x61, 0x74, 0x69, 0x6c, 0x65, 0x50, 0x72,
	0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x33,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x64, 0x74, 0x6f, 0x2e, 0x45, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x52, 0x15, 0x6e, 0x6f, 0x6e, 0x56, 0x6f, 0x6c, 0x61, 0x74, 0x69, 0x6c, 0x65,
	0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12, 0x63, 0x0a, 0x12, 0x76, 0x6f,
	0x6c, 0x61, 0x74, 0x69, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f,
	0x64, 0x74, 0x6f, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65,
	0x72, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x12, 0x76, 0x6f, 0x6c,
	0x61, 0x74, 0x69, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12,
	0x65, 0x0a, 0x13, 0x68, 0x65, 0x75, 0x72, 0x69, 0x73, 0x74, 0x69, 0x63, 0x50, 0x72, 0x6f, 0x70,
	0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x64, 0x74, 0x6f, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x52, 0x13, 0x68, 0x65, 0x75, 0x72, 0x69, 0x73, 0x74, 0x69, 0x63, 0x50, 0x72, 0x6f, 0x70,
	0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12, 0x32, 0x0a, 0x12, 0x68, 0x65, 0x75, 0x72, 0x69, 0x73,
	0x74, 0x69, 0x63, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x05, 0x3a, 0x02, 0x37, 0x35, 0x52, 0x12, 0x68, 0x65, 0x75, 0x72, 0x69, 0x73, 0x74, 0x69,
	0x63, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x1a, 0x26, 0x0a, 0x10, 0x50, 0x72,
	0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x42, 0x53, 0x0a, 0x1f, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x6d, 0x74, 0x75, 0x72, 0x62,
	0x6f, 0x2e, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x73, 0x64, 0x6b, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x69, 0x62,
	0x6d, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x6e, 0x6f, 0x6d, 0x69, 0x63,
	0x2f, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x2d, 0x67, 0x6f, 0x2d, 0x73, 0x64, 0x6b, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
})

var (
	file_IdentityMetadata_proto_rawDescOnce sync.Once
	file_IdentityMetadata_proto_rawDescData []byte
)

func file_IdentityMetadata_proto_rawDescGZIP() []byte {
	file_IdentityMetadata_proto_rawDescOnce.Do(func() {
		file_IdentityMetadata_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_IdentityMetadata_proto_rawDesc), len(file_IdentityMetadata_proto_rawDesc)))
	})
	return file_IdentityMetadata_proto_rawDescData
}

var file_IdentityMetadata_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_IdentityMetadata_proto_goTypes = []any{
	(*EntityIdentityMetadata)(nil),                  // 0: common_dto.EntityIdentityMetadata
	(*EntityIdentityMetadata_PropertyMetadata)(nil), // 1: common_dto.EntityIdentityMetadata.PropertyMetadata
	(EntityDTO_EntityType)(0),                       // 2: common_dto.EntityDTO.EntityType
}
var file_IdentityMetadata_proto_depIdxs = []int32{
	2, // 0: common_dto.EntityIdentityMetadata.entityType:type_name -> common_dto.EntityDTO.EntityType
	1, // 1: common_dto.EntityIdentityMetadata.nonVolatileProperties:type_name -> common_dto.EntityIdentityMetadata.PropertyMetadata
	1, // 2: common_dto.EntityIdentityMetadata.volatileProperties:type_name -> common_dto.EntityIdentityMetadata.PropertyMetadata
	1, // 3: common_dto.EntityIdentityMetadata.heuristicProperties:type_name -> common_dto.EntityIdentityMetadata.PropertyMetadata
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_IdentityMetadata_proto_init() }
func file_IdentityMetadata_proto_init() {
	if File_IdentityMetadata_proto != nil {
		return
	}
	file_CommonDTO_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_IdentityMetadata_proto_rawDesc), len(file_IdentityMetadata_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_IdentityMetadata_proto_goTypes,
		DependencyIndexes: file_IdentityMetadata_proto_depIdxs,
		MessageInfos:      file_IdentityMetadata_proto_msgTypes,
	}.Build()
	File_IdentityMetadata_proto = out.File
	file_IdentityMetadata_proto_goTypes = nil
	file_IdentityMetadata_proto_depIdxs = nil
}
