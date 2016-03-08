package structomancer

import (
	"errors"
	"reflect"
)

type (
	Structomancer struct {
		*StructSpec
		tagName string

		fieldEncoders, fieldDecoders map[string]FieldCoderFunc
	}

	FieldCoderFunc func(interface{}) (interface{}, error)
)

func New(specimen interface{}, tagName string) *Structomancer {
	return &Structomancer{
		tagName:       tagName,
		StructSpec:    StructSpecForType(specimen, tagName),
		fieldEncoders: make(map[string]FieldCoderFunc),
		fieldDecoders: make(map[string]FieldCoderFunc),
	}
}

func (z *Structomancer) SetFieldEncoder(fname string, encoder FieldCoderFunc) {
	z.fieldEncoders[fname] = encoder
}

func (z *Structomancer) SetFieldDecoder(fname string, decoder FieldCoderFunc) {
	z.fieldDecoders[fname] = decoder
}

func (z *Structomancer) IsKnownField(fname string) bool {
	return z.Field(fname) != nil
}

func (z *Structomancer) GetFieldValue(aStruct interface{}, fnickname string) reflect.Value {
	if !z.IsKnownField(fnickname) {
		panic("Unknown field " + fnickname)
	}

	fname := z.Field(fnickname).Name()
	v := reflect.ValueOf(aStruct)
	if v.Kind() == reflect.Ptr && !v.Elem().IsValid() {
		panic("Cannot call .GetField(nil, ...)")
	}

	v = EnsureStructOrStructPointerValue(v)

	if IsStructPtrValue(v) {
		return reflect.Indirect(v).FieldByName(fname)
	} else if IsStructValue(v) {
		return v.FieldByName(fname)
	} else {
		panic("structomancer.GetField: Unsupported type")
	}
}

func (z *Structomancer) SetField(aStruct interface{}, fname string, value interface{}) {
	z.GetFieldValue(aStruct, fname).Set(reflect.ValueOf(value))
}

// Sets `field` to `value` in `aStruct`, converting the value if it is of a convertible type.  If it
// is not convertible to the receiving field's type, this function returns an error.
func (z *Structomancer) ConvertSetField(aStruct interface{}, fname string, value interface{}) error {
	field := z.Field(fname)

	var v reflect.Value
	if value == nil {
		v = reflect.Zero(field.Type())

	} else {
		v = reflect.ValueOf(value)

		if v.Type() == field.Type() {
			// no-op
		} else {
			if v.Type().ConvertibleTo(field.Type()) {
				v = v.Convert(field.Type())
			} else {
				return errors.New("structomancer.ConvertSetField: Cannot convert " + v.Type().String() + " to " + field.Type().String())
			}
		}
	}

	z.GetFieldValue(aStruct, fname).Set(v)
	return nil
}

// Returns a pointer to a new, empty instance of the struct, regardless of whether the struct type
// is a struct or a pointer to a struct.  This method is appropriate for creating an instance that is
// guaranteed to be addressable (see reflect.Value.CanAddr()).
func (z *Structomancer) MakeEmpty() interface{} {
	if z.Kind() == reflect.Ptr {
		return reflect.New(z.Type().Elem()).Interface()
	} else {
		return reflect.New(z.Type()).Interface()
	}
}

func (z *Structomancer) StructToMap(aStruct interface{}) (map[string]interface{}, error) {
	fieldMap := make(map[string]interface{})

	for fname, field := range z.Fields() {
		// throw away fields that aren't known (or are marked with a "-" struct field tag)
		if !z.IsKnownField(fname) {
			continue
		}

		rval := z.GetFieldValue(aStruct, field.SerializedName())
		var val interface{}
		if rval.IsValid() {
			val = rval.Interface()
		}

		// handle zero values / nil:
		//  - if the field is a pointer or interface, val == nil
		//  - otherwise, the field is set to an empty instance of its type
		if val == nil || IsZero(rval) {
			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				val = nil
			} else {
				val = reflect.New(field.Type()).Elem().Interface()
			}
		}

		if val != nil {
			var subtag string
			if sub, isDefined := field.FlagValue("@tag"); isDefined {
				subtag = sub
			} else {
				subtag = z.tagName
			}

			var nv reflect.Value
			if encode, ok := z.fieldEncoders[fname]; ok {
				val, err := encode(val)
				if err != nil {
					return nil, err
				}
				nv = reflect.ValueOf(val)

			} else {
				var err error
				nv, err = ToNativeValue(rval, subtag)
				if err != nil {
					return nil, err
				}
			}

			val = nv.Interface()
		}
		fieldMap[fname] = val
	}
	return fieldMap, nil
}

// Returns a struct created by deserializing the contents of `fields`.
func (z *Structomancer) MapToStruct(fields map[string]interface{}) (interface{}, error) {
	aStruct := z.MakeEmpty()

	for fname, mapVal := range fields {
		if !z.IsKnownField(fname) {
			continue
		} else if mapVal == nil || IsZero(reflect.ValueOf(mapVal)) {
			continue
		}

		field := z.Field(fname)

		var subtag string
		if tag, isDefined := field.FlagValue("@tag"); isDefined {
			subtag = tag
		} else {
			subtag = z.tagName
		}

		var v reflect.Value
		if decode, ok := z.fieldDecoders[fname]; ok {
			mapVal, err := decode(mapVal)
			if err != nil {
				return nil, err
			}
			v = reflect.ValueOf(mapVal)

		} else {
			var err error
			v, err = FromNativeValue(reflect.ValueOf(mapVal), field.Type(), subtag)
			if err != nil {
				return nil, err
			}
		}

		if IsZero(v) {
			continue
		}

		z.GetFieldValue(aStruct, field.SerializedName()).Set(v)
	}

	// if the structomancer's type is a struct, not a struct pointer, dereference the pointer so we
	// return the right type
	if IsStructType(z.Type()) {
		return reflect.ValueOf(aStruct).Elem().Interface(), nil
	} else {
		return aStruct, nil
	}
}
