package structomancer_test

import (
	"github.com/brynbellomy/go-structomancer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tag", func() {
	type testStruct struct {
		SomeField string `xyzzy:"someField, @weezy, @flagKey=someValue"`
	}

	z := structomancer.New(&testStruct{}, "xyzzy")
	field := z.Field("someField")

	Context("when reading a struct field with a flag", func() {
		It("should recognize the existence of the flag", func() {
			Expect(field.IsFlagged("@weezy")).To(BeTrue())
		})
	})

	Context("when reading a struct field with a key/value flag", func() {
		It("should return the value of the flag when .FlagValue is called", func() {
			val, found := field.FlagValue("@flagKey")
			Expect(val).To(Equal("someValue"))
			Expect(found).To(BeTrue())
		})
	})
})
