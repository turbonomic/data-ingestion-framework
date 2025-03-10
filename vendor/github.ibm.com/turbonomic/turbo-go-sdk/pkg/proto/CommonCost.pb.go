// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v5.29.3
// source: CommonCost.proto

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

type PaymentOption int32

const (
	// The user must pay the entire price of this instance upfront. There is no recurring
	// cost.
	// (e.g. $10000.00 upfront for the year)
	PaymentOption_ALL_UPFRONT PaymentOption = 1
	// The user must pay some part of the instance price upfront, and the rest over time.
	// (e.g. $1000.00 upfront, and $0.5 per instance-hour afterwards).
	PaymentOption_PARTIAL_UPFRONT PaymentOption = 2
	// The entire price of the instance is recurring
	// (e.g. $0.7 per instance-hour)
	PaymentOption_NO_UPFRONT PaymentOption = 3
)

// Enum value maps for PaymentOption.
var (
	PaymentOption_name = map[int32]string{
		1: "ALL_UPFRONT",
		2: "PARTIAL_UPFRONT",
		3: "NO_UPFRONT",
	}
	PaymentOption_value = map[string]int32{
		"ALL_UPFRONT":     1,
		"PARTIAL_UPFRONT": 2,
		"NO_UPFRONT":      3,
	}
)

func (x PaymentOption) Enum() *PaymentOption {
	p := new(PaymentOption)
	*p = x
	return p
}

func (x PaymentOption) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PaymentOption) Descriptor() protoreflect.EnumDescriptor {
	return file_CommonCost_proto_enumTypes[0].Descriptor()
}

func (PaymentOption) Type() protoreflect.EnumType {
	return &file_CommonCost_proto_enumTypes[0]
}

func (x PaymentOption) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *PaymentOption) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = PaymentOption(num)
	return nil
}

// Deprecated: Use PaymentOption.Descriptor instead.
func (PaymentOption) EnumDescriptor() ([]byte, []int) {
	return file_CommonCost_proto_rawDescGZIP(), []int{0}
}

type PriceModel int32

const (
	PriceModel_ON_DEMAND PriceModel = 1
	// For GCP, this can be used to represent the debit.
	PriceModel_CREDIT PriceModel = 2
	// Ex: Reserved license fee, reserved compute cost for AWS + Azure (cost = 0)
	PriceModel_RESERVED            PriceModel = 3
	PriceModel_SPOT                PriceModel = 4
	PriceModel_FREE_TIER           PriceModel = 5
	PriceModel_COMMITMENT_COVERED  PriceModel = 6
	PriceModel_UNKNOWN_PRICE_MODEL PriceModel = 2047
)

// Enum value maps for PriceModel.
var (
	PriceModel_name = map[int32]string{
		1:    "ON_DEMAND",
		2:    "CREDIT",
		3:    "RESERVED",
		4:    "SPOT",
		5:    "FREE_TIER",
		6:    "COMMITMENT_COVERED",
		2047: "UNKNOWN_PRICE_MODEL",
	}
	PriceModel_value = map[string]int32{
		"ON_DEMAND":           1,
		"CREDIT":              2,
		"RESERVED":            3,
		"SPOT":                4,
		"FREE_TIER":           5,
		"COMMITMENT_COVERED":  6,
		"UNKNOWN_PRICE_MODEL": 2047,
	}
)

func (x PriceModel) Enum() *PriceModel {
	p := new(PriceModel)
	*p = x
	return p
}

func (x PriceModel) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PriceModel) Descriptor() protoreflect.EnumDescriptor {
	return file_CommonCost_proto_enumTypes[1].Descriptor()
}

func (PriceModel) Type() protoreflect.EnumType {
	return &file_CommonCost_proto_enumTypes[1]
}

func (x PriceModel) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *PriceModel) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = PriceModel(num)
	return nil
}

// Deprecated: Use PriceModel.Descriptor instead.
func (PriceModel) EnumDescriptor() ([]byte, []int) {
	return file_CommonCost_proto_rawDescGZIP(), []int{1}
}

// An amount of money, expressed in some currency.
type CurrencyAmount struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The currency in which the amount is expressed.
	// This is the ISO 4217 numeric code.
	// The default (840) is the USD currency code.
	//
	// We use the ISO 4217 standard so that in the future it would be easier to integrate
	// with JSR 354: Money and Currency API.
	Currency *int32 `protobuf:"varint,1,opt,name=currency,def=840" json:"currency,omitempty"`
	// The value, in the currency.
	// This should be non-negative, with 0 representing "free".
	Amount        *float64 `protobuf:"fixed64,2,opt,name=amount" json:"amount,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

// Default values for CurrencyAmount fields.
const (
	Default_CurrencyAmount_Currency = int32(840)
)

func (x *CurrencyAmount) Reset() {
	*x = CurrencyAmount{}
	mi := &file_CommonCost_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CurrencyAmount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CurrencyAmount) ProtoMessage() {}

func (x *CurrencyAmount) ProtoReflect() protoreflect.Message {
	mi := &file_CommonCost_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CurrencyAmount.ProtoReflect.Descriptor instead.
func (*CurrencyAmount) Descriptor() ([]byte, []int) {
	return file_CommonCost_proto_rawDescGZIP(), []int{0}
}

func (x *CurrencyAmount) GetCurrency() int32 {
	if x != nil && x.Currency != nil {
		return *x.Currency
	}
	return Default_CurrencyAmount_Currency
}

func (x *CurrencyAmount) GetAmount() float64 {
	if x != nil && x.Amount != nil {
		return *x.Amount
	}
	return 0
}

var File_CommonCost_proto protoreflect.FileDescriptor

var file_CommonCost_proto_rawDesc = string([]byte{
	0x0a, 0x10, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x43, 0x6f, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x64, 0x74, 0x6f, 0x22, 0x49,
	0x0a, 0x0e, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x1f, 0x0a, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x3a, 0x03, 0x38, 0x34, 0x30, 0x52, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63,
	0x79, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x01, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2a, 0x45, 0x0a, 0x0d, 0x50, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0f, 0x0a, 0x0b, 0x41, 0x4c,
	0x4c, 0x5f, 0x55, 0x50, 0x46, 0x52, 0x4f, 0x4e, 0x54, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x50,
	0x41, 0x52, 0x54, 0x49, 0x41, 0x4c, 0x5f, 0x55, 0x50, 0x46, 0x52, 0x4f, 0x4e, 0x54, 0x10, 0x02,
	0x12, 0x0e, 0x0a, 0x0a, 0x4e, 0x4f, 0x5f, 0x55, 0x50, 0x46, 0x52, 0x4f, 0x4e, 0x54, 0x10, 0x03,
	0x2a, 0x80, 0x01, 0x0a, 0x0a, 0x50, 0x72, 0x69, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12,
	0x0d, 0x0a, 0x09, 0x4f, 0x4e, 0x5f, 0x44, 0x45, 0x4d, 0x41, 0x4e, 0x44, 0x10, 0x01, 0x12, 0x0a,
	0x0a, 0x06, 0x43, 0x52, 0x45, 0x44, 0x49, 0x54, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x52, 0x45,
	0x53, 0x45, 0x52, 0x56, 0x45, 0x44, 0x10, 0x03, 0x12, 0x08, 0x0a, 0x04, 0x53, 0x50, 0x4f, 0x54,
	0x10, 0x04, 0x12, 0x0d, 0x0a, 0x09, 0x46, 0x52, 0x45, 0x45, 0x5f, 0x54, 0x49, 0x45, 0x52, 0x10,
	0x05, 0x12, 0x16, 0x0a, 0x12, 0x43, 0x4f, 0x4d, 0x4d, 0x49, 0x54, 0x4d, 0x45, 0x4e, 0x54, 0x5f,
	0x43, 0x4f, 0x56, 0x45, 0x52, 0x45, 0x44, 0x10, 0x06, 0x12, 0x18, 0x0a, 0x13, 0x55, 0x4e, 0x4b,
	0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x50, 0x52, 0x49, 0x43, 0x45, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x4c,
	0x10, 0xff, 0x0f, 0x42, 0x53, 0x0a, 0x1f, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x6d, 0x74, 0x75, 0x72,
	0x62, 0x6f, 0x2e, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x73, 0x64, 0x6b, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x69,
	0x62, 0x6d, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x6e, 0x6f, 0x6d, 0x69,
	0x63, 0x2f, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x2d, 0x67, 0x6f, 0x2d, 0x73, 0x64, 0x6b, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
})

var (
	file_CommonCost_proto_rawDescOnce sync.Once
	file_CommonCost_proto_rawDescData []byte
)

func file_CommonCost_proto_rawDescGZIP() []byte {
	file_CommonCost_proto_rawDescOnce.Do(func() {
		file_CommonCost_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_CommonCost_proto_rawDesc), len(file_CommonCost_proto_rawDesc)))
	})
	return file_CommonCost_proto_rawDescData
}

var file_CommonCost_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_CommonCost_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CommonCost_proto_goTypes = []any{
	(PaymentOption)(0),     // 0: common_dto.PaymentOption
	(PriceModel)(0),        // 1: common_dto.PriceModel
	(*CurrencyAmount)(nil), // 2: common_dto.CurrencyAmount
}
var file_CommonCost_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_CommonCost_proto_init() }
func file_CommonCost_proto_init() {
	if File_CommonCost_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_CommonCost_proto_rawDesc), len(file_CommonCost_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CommonCost_proto_goTypes,
		DependencyIndexes: file_CommonCost_proto_depIdxs,
		EnumInfos:         file_CommonCost_proto_enumTypes,
		MessageInfos:      file_CommonCost_proto_msgTypes,
	}.Build()
	File_CommonCost_proto = out.File
	file_CommonCost_proto_goTypes = nil
	file_CommonCost_proto_depIdxs = nil
}
