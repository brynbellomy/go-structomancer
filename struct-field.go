package structomancer

import "reflect"

type (
	FieldSpec struct {
		name  string
		tag   tag
		index []int
		rType reflect.Type
		rKind reflect.Kind
	}

	IFieldSpec interface {
		Name() string
		Nickname() string
		Index() []int

		Type() reflect.Type
		Kind() reflect.Kind

		// Tag() Tag
		TagName() string
		IsFlagged(flag string) bool
		FlagValue(flag string) (string, bool)
	}
)

func newFieldSpec(field reflect.StructField, tagName string) *FieldSpec {
	// it's worth caching the reflect.StructField data, as calling `.Field(...)` on a reflect.Value
	// creates the reflect.StructField from scratch every time
	return &FieldSpec{
		name:  field.Name,
		rType: field.Type,
		rKind: field.Type.Kind(),
		index: field.Index,
		tag:   newTag(field, tagName),
	}
}

func (f *FieldSpec) Name() string {
	return f.name
}

func (f *FieldSpec) Index() []int {
	return f.index
}

// func (f *FieldSpec) Tag() Tag {
// 	return f.tag
// }

func (f *FieldSpec) Type() reflect.Type {
	return f.rType
}

func (f *FieldSpec) Kind() reflect.Kind {
	return f.rKind
}

func (f *FieldSpec) TagName() string {
	return f.tag.TagName()
}

func (f *FieldSpec) Nickname() string {
	return f.tag.Nickname()
}

func (f *FieldSpec) IsFlagged(flag string) bool {
	return f.tag.IsFlagged(flag)
}

func (f *FieldSpec) FlagValue(flag string) (string, bool) {
	return f.tag.FlagValue(flag)
}
