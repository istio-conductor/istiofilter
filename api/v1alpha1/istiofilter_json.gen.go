// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api/v1alpha1/istiofilter.proto

// `IstioFilter` defines filters that apply to istio configuration.

package v1alpha1

import (
	bytes "bytes"
	fmt "fmt"
	github_com_gogo_protobuf_jsonpb "github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// MarshalJSON is a custom marshaler for IstioFilter
func (this *IstioFilter) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter
func (this *IstioFilter) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioFilter_Selector
func (this *IstioFilter_Selector) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter_Selector
func (this *IstioFilter_Selector) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioFilter_Change
func (this *IstioFilter_Change) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter_Change
func (this *IstioFilter_Change) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioFilter_StringMatch
func (this *IstioFilter_StringMatch) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter_StringMatch
func (this *IstioFilter_StringMatch) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioFilter_Match
func (this *IstioFilter_Match) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter_Match
func (this *IstioFilter_Match) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioFilter_SelectorMatch
func (this *IstioFilter_SelectorMatch) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter_SelectorMatch
func (this *IstioFilter_SelectorMatch) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioFilter_Patch
func (this *IstioFilter_Patch) MarshalJSON() ([]byte, error) {
	str, err := IstiofilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioFilter_Patch
func (this *IstioFilter_Patch) UnmarshalJSON(b []byte) error {
	return IstiofilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

var (
	IstiofilterMarshaler   = &github_com_gogo_protobuf_jsonpb.Marshaler{}
	IstiofilterUnmarshaler = &github_com_gogo_protobuf_jsonpb.Unmarshaler{AllowUnknownFields: true}
)