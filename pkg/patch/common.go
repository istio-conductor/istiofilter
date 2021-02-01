package patch

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

var (
	ErrUnknownKind          = errors.New("unknown kind")
	ErrUnknownContext       = errors.New("unknown context")
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrUnknownSpec          = errors.New("unknown spec")
)

func matchLabels(labels, match map[string]string) bool {
	for k, v := range match {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func merge(dst proto.Message, src *types.Struct) {
	srcJS, _ := (&jsonpb.Marshaler{}).MarshalToString(src)
	v := reflect.New(reflect.TypeOf(dst).Elem())
	srcMsg := v.Interface().(proto.Message)
	_ = (&jsonpb.Unmarshaler{AllowUnknownFields: true}).Unmarshal(bytes.NewBufferString(srcJS), srcMsg)
	proto.Merge(dst, srcMsg)
}
