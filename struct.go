package structomancer

import "reflect"

type (
	StructSpec struct {
		rType      reflect.Type
		rKind      reflect.Kind
		tagName    string
		fields     map[string]*FieldSpec
		fieldNames []string // cached
	}
)

func NewStructSpec(t reflect.Type, tagName string) *StructSpec {
	t = EnsureStructOrStructPointerType(t)
	st := EnsureStructType(t)

	var fields []reflect.StructField
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		// skip fields marked with "-", just like the json package
		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}
		fields = append(fields, field)
	}

	fieldMap := make(map[string]*FieldSpec, len(fields))
	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fSpec := NewFieldSpec(field, tagName)
		fieldMap[fSpec.SerializedName()] = fSpec
		fieldNames[i] = fSpec.SerializedName()
	}

	return &StructSpec{
		rType:      t,
		rKind:      t.Kind(),
		tagName:    tagName,
		fields:     fieldMap,
		fieldNames: fieldNames,
	}
}

func (s *StructSpec) Type() reflect.Type {
	return s.rType
}

func (s *StructSpec) Kind() reflect.Kind {
	return s.rKind
}

func (s *StructSpec) TagName() string {
	return s.tagName
}

func (s *StructSpec) Fields() map[string]*FieldSpec {
	return s.fields
}

// Returns a *FieldSpec object representing the given field, or nil if one was not found.
func (s *StructSpec) Field(sFieldName string) *FieldSpec {
	if f, exists := s.Fields()[sFieldName]; exists {
		return f
	} else {
		return nil
	}
}

// Returns the number of exported fields recognized by Structomancer in the struct type.
func (s *StructSpec) NumFields() int {
	return len(s.Fields())
}

// Returns a slice of the exported field names recognized by Structomancer in the struct type.
func (s *StructSpec) FieldNames() []string {
	return s.fieldNames
}
