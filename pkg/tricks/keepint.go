package tricks

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"istio.io/istio/pkg/config"
)

// Key is the array index or map key.
// maybe int or string.
type Key interface{}

type stateContext struct {
	state []Key
	paths [][]Key
}

func (ctx *stateContext) End() {
	var copied []Key
	for _, p := range ctx.state {
		copied = append(copied, p)
	}
	ctx.paths = append(ctx.paths, copied)
}

func (ctx *stateContext) Enter(curr Key) {
	ctx.state = append(ctx.state, curr)
}

func (ctx *stateContext) Quit() {
	ctx.state = ctx.state[:len(ctx.state)-1]
}

func FindInt64Field(spec config.Spec) [][]Key {
	value := reflect.ValueOf(spec)
	ctx := &stateContext{}
	collect(nil, value, ctx)
	return ctx.paths
}

func RestoreIntField(m map[string]interface{}, fields [][]Key) (err error) {
	for _, field := range fields {
		var v interface{}
		v = m
	F:
		for i, key := range field {
			switch cvt := v.(type) {
			case []interface{}:
				k, ok := key.(int)
				if !ok {
					return errors.New("keys type wrong")
				}
				if k >= len(cvt) {
					return errors.New("index out of range")
				}
				v = cvt[k]
				if i == len(field)-1 {
					switch str := v.(type) {
					case string:
						cvt[k], err = strconv.Atoi(str)
						if err != nil {
							return err
						}
					case []interface{}:
						for _, v := range str {
							s, ok := v.(string)
							if !ok {
								continue
							}
							str[i], err = strconv.Atoi(s)
							if err != nil {
								continue
							}
						}
					}
				}
			case map[string]interface{}:
				k, ok := key.(string)
				if !ok {
					return errors.New("keys type wrong")
				}
				v, ok = cvt[k]
				if !ok {
					// this field not marshal into json
					break F
				}
				if i == len(field)-1 {
					switch str := v.(type) {
					case string:
						cvt[k], err = strconv.Atoi(str)
						if err != nil {
							return err
						}
					case []interface{}:
						for _, v := range str {
							s, ok := v.(string)
							if !ok {
								continue
							}
							str[i], err = strconv.Atoi(s)
							if err != nil {
								continue
							}
						}
					}
				}
			default:
				break
			}
		}

	}
	return nil
}

func collect(curr Key, v reflect.Value, ctx *stateContext) {
	if v.Kind() == reflect.Ptr {
		collect(curr, v.Elem(), ctx)
		return
	}
	if curr != nil {
		ctx.Enter(curr)
		defer ctx.Quit()
	}

	switch v.Kind() {
	case reflect.Uint64, reflect.Int64:
		ctx.End()
		return
	case reflect.Struct:
		structType := v.Type()
		num := structType.NumField()
		for i := 0; i < num; i++ {
			value := v.Field(i)
			valueField := structType.Field(i)
			if strings.HasPrefix(valueField.Name, "XXX_") {
				continue
			}
			switch value.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
				if value.IsNil() {
					continue
				}
			}
			if value.Type().Implements(wktType) {
				continue
			}

			if valueField.Tag.Get("protobuf_oneof") != "" {
				// value is an interface containing &T{real_value}.
				sv := value.Elem().Elem() // interface -> *T -> T
				value = sv.Field(0)
				prop := jsonProperties(sv.Type().Field(0), false)
				collect(prop.JSONName, value, ctx)
				continue
			}
			//this is not a protobuf field
			if valueField.Tag.Get("protobuf") == "" {
				continue
			}
			prop := jsonProperties(valueField, false)
			collect(prop.JSONName, v.Field(i), ctx)
		}
	case reflect.Slice:
		if v.Len() > 0 {
			switch v.Index(0).Kind() {
			case reflect.Int64, reflect.Uint64:
				ctx.End()
				return
			}
		}
		for i := 0; i < v.Len(); i++ {
			collect(i, v.Index(i), ctx)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if iter.Key().Kind() != reflect.String { //weird!
				continue
			}
			collect(iter.Key().String(), iter.Value(), ctx)
		}
	default:
		return
	}
}

// jsonProperties returns parsed proto.Properties for the field and corrects JSONName attribute.
func jsonProperties(f reflect.StructField, origName bool) *proto.Properties {
	var prop proto.Properties
	prop.Init(f.Type, f.Name, f.Tag.Get("protobuf"), &f)
	if origName || prop.JSONName == "" {
		prop.JSONName = prop.OrigName
	}
	return &prop
}

func ToJSON(spec config.Spec) (data []byte, err error) {
	field := FindInt64Field(spec)
	m, err := config.ToMap(spec)
	if err != nil {
		return nil, err
	}
	err = RestoreIntField(m, field)
	if err != nil {
		return nil, err
	}
	data, _ = json.Marshal(m)
	return data, err
}

type isWkt interface {
	XXX_WellKnownType() string
}

var (
	wktType = reflect.TypeOf((*isWkt)(nil)).Elem()
)
