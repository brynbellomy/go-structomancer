package structomancer_test

import (
	"reflect"

	"github.com/listenonrepeat/backend/common/structomancer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ToNativeValue", func() {
	type Name string
	type Age uint32
	type Ok bool

	It("should deserialize typealiased values to their underlying native Go types", func() {
		n := Name("bryn")
		v, err := structomancer.ToNativeValue(reflect.ValueOf(n), "")
		if err != nil {
			Fail(err.Error())
		}
		Expect(v.Interface()).To(Equal("bryn"))

		a := Age(29)
		v, err = structomancer.ToNativeValue(reflect.ValueOf(a), "")
		if err != nil {
			Fail(err.Error())
		}
		Expect(v.Interface()).To(Equal(uint32(29)))

		ok := Ok(true)
		v, err = structomancer.ToNativeValue(reflect.ValueOf(ok), "")
		if err != nil {
			Fail(err.Error())
		}
		Expect(v.Interface()).To(Equal(true))
	})

	It("should deserialize typed slices to generic/native slices containing only native-typed values", func() {
		slice := []Ok{true, false, true}
		v, err := structomancer.ToNativeValue(reflect.ValueOf(slice), "")
		if err != nil {
			Fail(err.Error())
		}
		Expect(v.Interface()).To(Equal([]interface{}{true, false, true}))
	})
})
