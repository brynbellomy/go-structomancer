package structomancer

import (
	"reflect"
	"strings"
)

type (
	Tag struct {
		tagName   string // the name of the tag itself, i.e., "api" in `api:"myField,data,blah"`
		fieldName string // the first element of the comma-separated tag contents
		TagParts         // the rest of the elements of the tag string after the `FieldName`
	}

	TagParts []string
)

func NewTag(field reflect.StructField, tagName string) Tag {
	tagParts := strings.Split(field.Tag.Get(tagName), ",")

	// ignore spaces — they can be helpful for readability
	for i, part := range tagParts {
		tagParts[i] = strings.Trim(part, " \t")
	}

	// the first component of the tag string is the "serialized" (i.e., non-struct, i.e., JSON-y) name of the field
	sFieldName := tagParts[0]
	// if it isn't specified, we give it a default name (which is just its Go name)
	if sFieldName == "" {
		sFieldName = field.Name
	}

	// tagParts only contains the parts after the serialized name
	tagParts = tagParts[1:]

	return Tag{
		tagName:   tagName,
		fieldName: sFieldName,
		TagParts:  TagParts(tagParts),
	}
}

func (t Tag) TagName() string {
	return t.tagName
}

func (t Tag) FieldName() string {
	return t.fieldName
}

func (t TagParts) IsFlagged(flag string) bool {
	for _, part := range t {
		if part == flag {
			return true
		}
	}
	return false
}

func (t TagParts) FlagValue(flag string) (string, bool) {
	for _, part := range t {
		if strings.HasPrefix(part, flag+"=") {
			return part[len(flag)+1:], true
		}
	}
	return "", false
}
