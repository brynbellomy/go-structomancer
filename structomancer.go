package structomancer

import (
	"errors"
	"reflect"
)

type (
	Structomancer struct {
		*structSpec
		tagName string

		fieldEncoders, fieldDecoders map[string]FieldCoderFunc
	}

	FieldCoderFunc func(interface{}) (interface{}, error)
)

func New(specimen interface{}, tagName string) *Structomancer {
	return NewWithType(reflect.TypeOf(specimen), tagName)
}

func NewWithType(t reflect.Type, tagName string) *Structomancer {
	return &Structomancer{
		tagName:       tagName,
		structSpec:    structSpecForType(t, tagName),
		fieldEncoders: make(map[string]FieldCoderFunc),
		fieldDecoders: make(map[string]FieldCoderFunc),
	}
}

// Sets the function used to encode the given field to a native Go value.
func (z *Structomancer) SetFieldEncoder(fname string, encoder FieldCoderFunc) {
	z.fieldEncoders[fname] = encoder
}

// Sets the function used to decode the given field from a native Go value.
func (z *Structomancer) SetFieldDecoder(fname string, decoder FieldCoderFunc) {
	z.fieldDecoders[fname] = decoder
}

// Returns a pointer to a new, empty instance of the struct, regardless of whether the struct type
// is a struct or a pointer to a struct.  This method is appropriate for creating an instance that is
// guaranteed to be addressable (see reflect.Value.CanAddr()).
func (z *Structomancer) MakeEmpty() interface{} {
	return z.MakeEmptyV().Interface()
}

// Returns a reflect.Value containing a pointer to a new, empty instance of the struct, regardless of
// whether the struct type is a struct or a pointer to a struct.  This method is appropriate for
// creating an instance that is guaranteed to be addressable (see reflect.Value.CanAddr()).
func (z *Structomancer) MakeEmptyV() reflect.Value {
	if z.Kind() == reflect.Ptr {
		return reflect.New(z.Type().Elem())
	} else {
		return reflect.New(z.Type())
	}
}

// If a field is marked as "-", it is considered "not known".  This function returns false for those
// fields, and true for all others (including fields with no struct tag).
func (z *Structomancer) IsKnownField(fname string) bool {
	return z.Field(fname) != nil
}

// Returns the value of the struct field with the given nickname.
func (z *Structomancer) GetFieldValue(aStruct interface{}, fnickname string) (interface{}, error) {
	fv, err := z.GetFieldValueV(reflect.ValueOf(aStruct), fnickname)
	if err != nil {
		return nil, err
	}
	return fv.Interface(), nil
}

// Returns a reflect.Value containing the value of the struct field with the given nickname.
func (z *Structomancer) GetFieldValueV(v reflect.Value, fnickname string) (reflect.Value, error) {
	field := z.Field(fnickname)
	if field == nil {
		return reflect.Value{}, errors.New("structomancer.GetFieldValue: unknown field '" + fnickname + "'")
	}

	if v.Kind() == reflect.Ptr && (!v.Elem().IsValid() || v.IsNil()) {
		return reflect.Value{}, errors.New("structomancer.GetFieldValue: aStruct argument cannot be nil")
	}

	var fieldVal reflect.Value
	if IsStructPtrValue(v) {
		fieldVal = reflect.Indirect(v).FieldByName(field.Name())
	} else if IsStructValue(v) {
		fieldVal = v.FieldByName(field.Name())
	} else {
		return reflect.Value{}, errors.New("structomancer.GetFieldValue: Unsupported type '" + v.Type().String() + "'")
	}

	if encoder, exists := z.fieldEncoders[fnickname]; exists {
		fv := fieldVal.Interface()
		fv, err := encoder(fv)
		if err != nil {
			// @@TODO
			return reflect.Value{}, errors.New("structomancer.GetFieldValue: error calling user encoder: " + err.Error())
		}
		fieldVal = reflect.ValueOf(fv)

	} else {
		var subtag string
		if sub, isDefined := field.FlagValue("@tag"); isDefined {
			subtag = sub
		} else {
			subtag = z.tagName
		}

		var err error
		fieldVal, err = ToNativeValue(fieldVal, subtag)
		if err != nil {
			return reflect.Value{}, errors.New("structomancer.GetFieldValue: " + err.Error())
		}
	}

	return fieldVal, nil
}

// Sets `field` to `value` in `aStruct`, converting the value if it is of a convertible type.  If it
// is not convertible to the receiving field's type, this function returns an error.
func (z *Structomancer) SetFieldValue(aStruct interface{}, fname string, value interface{}) error {
	return z.SetFieldValueV(reflect.ValueOf(aStruct), fname, reflect.ValueOf(value))
}

// Sets `field` to `value` in the struct contained by `sv`, converting the value if it is of a
// convertible type.  If it is not convertible to the receiving field's type, this function returns
// an error.
func (z *Structomancer) SetFieldValueV(sv reflect.Value, fname string, value reflect.Value) error {
	field := z.Field(fname)
	if field == nil {
		return errors.New("structomancer.SetFieldValue: unknown field '" + fname + "'")
	}

	if sv.Kind() == reflect.Ptr && (!sv.Elem().IsValid() || sv.IsNil()) {
		return errors.New("structomancer.SetFieldValue: struct argument cannot be nil")
	}

	var fieldVal reflect.Value
	if IsStructPtrValue(sv) {
		fieldVal = reflect.Indirect(sv).FieldByName(field.Name())
	} else if IsStructValue(sv) {
		fieldVal = sv.FieldByName(field.Name())
	} else {
		return errors.New("structomancer.GetFieldValue: unsupported type '" + sv.Type().String() + "'")
	}

	if decode, ok := z.fieldDecoders[fname]; ok {
		val, err := decode(value.Interface())
		if err != nil {
			return err
		}
		value = reflect.ValueOf(val)

	} else {
		var subtag string
		if sub, isDefined := field.FlagValue("@tag"); isDefined {
			subtag = sub
		} else {
			subtag = z.tagName
		}

		var err error
		value, err = FromNativeValue(value, field.Type(), subtag)
		if err != nil {
			return err
		}
	}

	fieldVal.Set(value)
	return nil
}

// Returns a map containing the contents of `aStruct`, taking into account the field tags defined for
// the current `tagName`.
func (z *Structomancer) StructToMap(aStruct interface{}) (map[string]interface{}, error) {
	return z.StructToMapV(reflect.ValueOf(aStruct))
}

// Returns a reflect.Value containing a map containing the contents of `aStruct`, taking into account
// the field tags defined for the current `tagName`.
func (z *Structomancer) StructToMapV(aStruct reflect.Value) (map[string]interface{}, error) {
	fieldMap := make(map[string]interface{})

	for fname, field := range z.Fields() {
		// throw away fields that aren't known (or are marked with a "-" struct field tag)
		if !z.IsKnownField(fname) {
			continue
		}

		rval, err := z.GetFieldValueV(aStruct, fname)
		if err != nil {
			return nil, err
		}

		var val interface{}
		if rval.IsValid() && rval.CanInterface() {
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

		fieldMap[fname] = val
	}
	return fieldMap, nil
}

// Returns a struct created by decoding the contents of `fields`.
func (z *Structomancer) MapToStruct(fields map[string]interface{}) (interface{}, error) {
	sv, err := z.MapToStructV(fields)
	if err != nil {
		return nil, err
	}

	return sv.Interface(), nil
}

// Returns a reflect.Value containing a struct created by decoding the contents of `fields`.
func (z *Structomancer) MapToStructV(fields map[string]interface{}) (reflect.Value, error) {
	aStruct := z.MakeEmptyV()

	for fname, mapVal := range fields {
		if !z.IsKnownField(fname) {
			continue
		} else if mapVal == nil || IsZero(reflect.ValueOf(mapVal)) {
			continue
		}

		err := z.SetFieldValueV(aStruct, fname, reflect.ValueOf(mapVal))
		if err != nil {
			return reflect.Value{}, err
		}
	}

	// if the structomancer's type is a struct, not a struct pointer, dereference the pointer so we
	// return the right type
	if IsStructType(z.Type()) {
		return aStruct.Elem(), nil
	} else {
		return aStruct, nil
	}
}
