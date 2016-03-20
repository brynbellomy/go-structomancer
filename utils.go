package structomancer

import (
	"errors"
	"fmt"
	"reflect"
)

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Func, reflect.Map, reflect.Slice, reflect.Chan:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && IsZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && IsZero(v.Field(i))
		}
		return z
	default:
		// Compare other types directly:
		z := reflect.Zero(v.Type())
		return v.Interface() == z.Interface()
	}
}

func IsStruct(x interface{}) bool {
	return IsStructValue(reflect.ValueOf(x))
}

func IsStructPtr(x interface{}) bool {
	return IsStructPtrValue(reflect.ValueOf(x))
}

func IsStructType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func IsStructValue(v reflect.Value) bool {
	return v.Kind() == reflect.Struct
}

func IsStructPtrType(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func IsStructPtrValue(v reflect.Value) bool {
	if !v.IsValid() {
		panic("Cannot call IsStructPtrValue on a nil value")
	}
	return v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct
}

func EnsureStructType(t reflect.Type) reflect.Type {
	if IsStructType(t) {
		return t
	} else if IsStructPtrType(t) {
		return t.Elem()
	} else {
		panic("structomancer: unsupported type " + t.String())
	}
}

func EnsureStructValue(v reflect.Value) reflect.Value {
	if IsStructValue(v) {
		return v
	} else if IsStructPtrValue(v) {
		return v.Elem()
	} else {
		if v.IsValid() {
			panic("structomancer: unsupported type " + v.Type().String())
		} else {
			panic("structomancer: unsupported type: nil")
		}
	}
}

func EnsureStructOrStructPointerType(t reflect.Type) reflect.Type {
	if IsStructType(t) || IsStructPtrType(t) {
		return t
	} else {
		panic("structomancer: unsupported type " + t.String())
	}
}

func EnsureStructOrStructPointerValue(v reflect.Value) reflect.Value {
	if IsStructValue(v) || IsStructPtrValue(v) {
		return v
	} else {
		if v.IsValid() {
			panic("structomancer: unsupported type " + v.Type().String())
		} else {
			panic("structomancer: unsupported type: nil")
		}
	}
}

func ConvertSet(dest reflect.Value, src reflect.Value, destType reflect.Type) error {
	if !src.IsValid() {
		dest.Set(src)

	} else if IsZero(src) {
		dest.Set(src)

	} else if src.Type() == destType {
		dest.Set(src)

	} else if src.Type().ConvertibleTo(destType) {
		dest.Set(src.Convert(destType))

	} else {
		return errors.New("structomancer.ConvertSet: Cannot convert " + src.Type().String() + " to " + destType.String())
	}

	return nil
}

var stringType = reflect.TypeOf("")

func ToNativeValue(v reflect.Value, subtag string) (nv reflect.Value, err error) {
	switch v.Kind() {
	case reflect.Invalid:
		return reflect.ValueOf(nil), nil
	case reflect.Bool:
		return reflect.ValueOf(v.Bool()), nil
	case reflect.Int:
		return reflect.ValueOf(int(v.Int())), nil
	case reflect.Int8:
		return reflect.ValueOf(int8(v.Int())), nil
	case reflect.Int16:
		return reflect.ValueOf(int16(v.Int())), nil
	case reflect.Int32:
		return reflect.ValueOf(int32(v.Int())), nil
	case reflect.Int64:
		return reflect.ValueOf(v.Int()), nil
	case reflect.Uint:
		return reflect.ValueOf(uint(v.Uint())), nil
	case reflect.Uint8:
		return reflect.ValueOf(uint8(v.Uint())), nil
	case reflect.Uint16:
		return reflect.ValueOf(uint16(v.Uint())), nil
	case reflect.Uint32:
		return reflect.ValueOf(uint32(v.Uint())), nil
	case reflect.Uint64:
		return reflect.ValueOf(uint64(v.Uint())), nil
	case reflect.Float32:
		return reflect.ValueOf(float32(v.Float())), nil
	case reflect.Float64:
		return reflect.ValueOf(v.Float()), nil
	case reflect.Complex64:
		return reflect.ValueOf(complex64(v.Complex())), nil
	case reflect.Complex128:
		return reflect.ValueOf(v.Complex()), nil
	case reflect.String:
		return reflect.ValueOf(v.String()), nil

	case reflect.Slice, reflect.Array:
		dest := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			src := reflect.ValueOf(v.Index(i).Interface())

			nval, err := ToNativeValue(src, subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			dest[i] = nval.Interface()
		}
		return reflect.ValueOf(dest), nil

	case reflect.Map:
		dest := make(map[string]interface{}, v.Len())
		ks := v.MapKeys()
		for i := 0; i < len(ks); i++ {
			// keys must be convertible to strings or this function will return an error
			nvKey, err := ToNativeValue(ks[i], subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			if nvKey.Type() != stringType {
				return reflect.Value{}, errors.New("structomancer.ToNativeValue: cannot convert " + ks[i].Type().String() + " to string (for map key)")
			}

			strKey := nvKey.Interface().(string)

			velem := v.MapIndex(ks[i])

			if velem.Kind() == reflect.Interface {
				// unwrap interfaces to expose the inner type
				velem = reflect.ValueOf(velem.Interface())
			}

			// convert to native value
			nval, err := ToNativeValue(velem, subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			dest[strKey] = nval.Interface()
		}
		return reflect.ValueOf(dest), nil

	case reflect.Struct:
		z := NewWithType(v.Type(), subtag)
		m, err := z.StructToMapV(v)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(m), nil

	case reflect.Ptr:
		if !v.IsValid() || v.IsNil() {
			return v, nil
		}

		// we simply collapse pointers when converting to native values
		innerVal, err := ToNativeValue(v.Elem(), subtag)
		if err != nil {
			return reflect.Value{}, err
		}
		return innerVal, nil

	case reflect.Interface:
		// @@TODO?
		return v, nil

	case reflect.Func,
		reflect.Chan,
		reflect.Uintptr,
		reflect.UnsafePointer:
		return v, nil

	default:
		panic("Unknown kind")
	}
}

func FromNativeValue(nv reflect.Value, destType reflect.Type, subtag string) (v reflect.Value, err error) {
	switch destType.Kind() {
	case reflect.Invalid:
		panic("structomancer.FromNativeValue: bad destType parameter, destType.Kind() = reflect.Invalid")

	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String:

		if nv.Kind() == reflect.Interface {
			nv = reflect.ValueOf(nv.Interface())
		}

		if nv.Type() == destType {
			return nv, nil
		} else if nv.Type().ConvertibleTo(destType) {
			return nv.Convert(destType), nil
		} else {
			return reflect.Value{}, errors.New("structomancer.FromNativeValue: cannot convert " + nv.Type().String() + " to " + destType.String())
		}

	case reflect.Slice:
		slice := reflect.MakeSlice(destType, nv.Len(), nv.Cap())

		for i := 0; i < nv.Len(); i++ {
			velem := nv.Index(i)

			if velem.Kind() == reflect.Interface {
				// this strips any existing `interface{}` wrapper so we can see the real type
				velem = reflect.ValueOf(velem.Interface())
			}

			velem, err := FromNativeValue(velem, destType.Elem(), subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			slice.Index(i).Set(velem)
		}

		return slice, nil

	case reflect.Array:
		n := nv.Len()
		arrayType := reflect.ArrayOf(n, destType.Elem())
		array := reflect.New(arrayType).Elem()

		for i := 0; i < n; i++ {
			velem := nv.Index(i)

			if velem.Kind() == reflect.Interface {
				// this strips any existing `interface{}` wrapper so we can see the real type
				velem = reflect.ValueOf(velem.Interface())
			}

			velem, err = FromNativeValue(velem, destType.Elem(), subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			array.Index(i).Set(velem)
		}

		return array, nil

	case reflect.Struct:
		z := NewWithType(destType, subtag)

		if nv.Kind() != reflect.Map {
			return reflect.Value{}, errors.New("structomancer.FromNativeValue: cannot convert " + nv.Type().String() + " to " + destType.String())
		}

		if m, ok := nv.Interface().(map[string]interface{}); ok {
			val, err := z.MapToStructV(m)
			if err != nil {
				return reflect.Value{}, err
			}
			return val, nil

		} else {
			aStruct := z.MakeEmpty()
			mapKeys := nv.MapKeys()
			aStructVal := reflect.ValueOf(aStruct)

			for i := 0; i < len(mapKeys); i++ {
				mapKey := mapKeys[i]
				if mapKey.Kind() == reflect.Interface {
					// this strips any existing `interface{}` wrapper so we can see the real type
					mapKey = reflect.ValueOf(mapKey.Interface())
				}

				if !mapKey.Type().ConvertibleTo(stringType) {
					return reflect.Value{}, fmt.Errorf("structomancer.FromNativeValue: cannot deserialize struct from map with non-string keys (type = %v)", mapKey.Type().String())
				}

				fname := mapKey.Convert(stringType).Interface().(string)
				if !z.IsKnownField(fname) {
					continue
				}

				mapVal := nv.MapIndex(mapKey)
				if !mapVal.IsValid() || IsZero(mapVal) {
					continue
				}

				if mapVal.Kind() == reflect.Interface {
					// this strips any existing `interface{}` wrapper so we can see the real type
					mapVal = reflect.ValueOf(mapVal.Interface())
				}

				err := z.SetFieldValueV(aStructVal, fname, mapVal)
				if err != nil {
					return reflect.Value{}, err
				}

			}
			return aStructVal.Elem(), nil
		}

	case reflect.Map:
		dest := reflect.MakeMap(destType)
		mapKeys := nv.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			velem := nv.MapIndex(mapKeys[i])

			if velem.Kind() == reflect.Interface {
				// this strips any existing `interface{}` wrapper so we can see the real type
				velem = reflect.ValueOf(velem.Interface())
			}

			cnvKey, err := FromNativeValue(mapKeys[i], destType.Key(), subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			velem, err = FromNativeValue(velem, destType.Elem(), subtag)
			if err != nil {
				return reflect.Value{}, err
			}

			dest.SetMapIndex(cnvKey, velem)
		}

		return dest, nil

	case reflect.Ptr:
		if !nv.IsValid() || nv.IsNil() {
			return nv, nil
		}
		innerVal, err := FromNativeValue(nv, destType.Elem(), subtag)
		if err != nil {
			return reflect.Value{}, err
		}
		ptrval := reflect.New(destType.Elem())
		ptrval.Elem().Set(innerVal)
		return ptrval, nil

	case reflect.Interface:
		// @@TODO?
		return nv, nil

	case reflect.Func,
		reflect.Chan,
		reflect.Uintptr,
		reflect.UnsafePointer:
		return nv, nil

	default:
		panic("Unknown kind")
	}
}
