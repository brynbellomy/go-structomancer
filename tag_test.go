package structomancer_test

import (
	"reflect"

	"github.com/brynbellomy/go-structomancer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tag", func() {
	type testStruct struct {
		SomeField string `xyzzy:"someField, @weezy, @flagKey=someValue"`
	}

	stype := reflect.TypeOf(testStruct{})
	field := structomancer.NewFieldSpec(stype.Field(0), "xyzzy")

	Context("when reading a struct field with a flag", func() {
		It("should recognize the existence of the flag", func() {
			Expect(field.IsFlagged("@weezy")).To(BeTrue())
		})
	})

	Context("when reading a struct field with a key/value flag", func() {
		It("should return the value of the flag when .FlagValue is called", func() {
			val, found := field.Tag().FlagValue("@flagKey")
			Expect(val).To(Equal("someValue"))
			Expect(found).To(BeTrue())
		})
	})
})
