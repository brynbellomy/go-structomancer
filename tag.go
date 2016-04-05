package structomancer

import (
	"reflect"
	"strings"
)

type (
	tag struct {
		tagName  string // the name of the tag itself, i.e., "api" in `api:"myField,data,blah"`
		nickname string // the first element of the comma-separated tag contents
		tagParts        // the rest of the elements of the tag string after the `nickname`
	}

	tagParts []string
)

func newTag(field reflect.StructField, tagName string) tag {
	parts := strings.Split(field.Tag.Get(tagName), ",")

	// ignore spaces — they can be helpful for readability
	for i, part := range parts {
		parts[i] = strings.Trim(part, " \t")
	}

	// the first component of the tag string is the "serialized" (i.e., non-struct, i.e., JSON-y) name of the field
	nickname := parts[0]
	// if it isn't specified, we give it a default name (which is just its Go name)
	if nickname == "" {
		nickname = field.Name
	}

	// parts only contains the parts after the serialized name
	parts = parts[1:]

	return tag{
		tagName:  tagName,
		nickname: nickname,
		tagParts: tagParts(parts),
	}
}

func (t tag) TagName() string {
	return t.tagName
}

func (t tag) Nickname() string {
	return t.nickname
}

func (t tagParts) IsFlagged(flag string) bool {
	for _, part := range t {
		if part == flag {
			return true
		}
	}
	return false
}

func (t tagParts) FlagValue(flag string) (string, bool) {
	for _, part := range t {
		if strings.HasPrefix(part, flag+"=") {
			return part[len(flag)+1:], true
		}
	}
	return "", false
}
