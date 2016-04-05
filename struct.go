package structomancer

import (
	"reflect"
	"strings"
)

type (
	structSpec struct {
		rType      reflect.Type
		rKind      reflect.Kind
		tagName    string
		fields     map[string]*FieldSpec
		fieldNames []string // cached
	}
)

func newStructSpec(t reflect.Type, tagName string) *structSpec {
	if !(IsStructType(t) || IsStructPtrType(t)) {
		panic("structomancer: unsupported type " + t.String())
	}

	var st reflect.Type
	if IsStructType(t) {
		st = t
	} else if IsStructPtrType(t) {
		st = t.Elem()
	} else {
		panic("structomancer: unsupported type " + t.String())
	}

	var fields []reflect.StructField
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		// skip fields marked with "-", just like the json package
		if tag := field.Tag.Get(tagName); strings.HasPrefix(tag, "-") {
			continue
		}
		fields = append(fields, field)
	}

	fieldMap := make(map[string]*FieldSpec, len(fields))
	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fSpec := newFieldSpec(field, tagName)
		fieldMap[fSpec.Nickname()] = fSpec
		fieldNames[i] = fSpec.Nickname()
	}

	return &structSpec{
		rType:      t,
		rKind:      t.Kind(),
		tagName:    tagName,
		fields:     fieldMap,
		fieldNames: fieldNames,
	}
}

func (s *structSpec) Type() reflect.Type {
	return s.rType
}

func (s *structSpec) Kind() reflect.Kind {
	return s.rKind
}

func (s *structSpec) TagName() string {
	return s.tagName
}

func (s *structSpec) Fields() map[string]*FieldSpec {
	return s.fields
}

// Returns a *FieldSpec object representing the given field, or nil if one was not found.
func (s *structSpec) Field(sFieldName string) *FieldSpec {
	if f, exists := s.Fields()[sFieldName]; exists {
		return f
	} else {
		return nil
	}
}

// Returns the number of exported fields recognized by Structomancer in the struct type.
func (s *structSpec) NumFields() int {
	return len(s.Fields())
}

// Returns a slice of the exported field names recognized by Structomancer in the struct type.
func (s *structSpec) FieldNames() []string {
	return s.fieldNames
}
