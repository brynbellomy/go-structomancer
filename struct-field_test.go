package structomancer_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )

// var _ = Describe("FieldSpec", func() {

// 	Context("when .ValueIn is called", func() {
// 		Context("on fields containing valid values", func() {
// 			It("should return those values", func() {
// 				t := Keith{}
// 				t.Name = Name("keith richards")
// 				t.Age = Age(429)

// 				st := reflect.TypeOf(t)

// 				field := structomancer.NewFieldSpec(st.Field(0), "xyzzy")
// 				Expect(field.ValueIn(t).Interface()).To(Equal(Name("keith richards")))

// 				field = structomancer.NewFieldSpec(st.Field(1), "xyzzy")
// 				Expect(field.ValueIn(t).Interface()).To(Equal(Age(429)))
// 			})
// 		})

// 		Context("on fields containing nils", func() {
// 			It("should return nils", func() {
// 				type NilStruct struct {
// 					Foo map[string]interface{} `xyzzy:"foo"`
// 					Bar *struct{ Quux string } `xyzzy:"bar"`
// 				}

// 				t := NilStruct{}
// 				st := reflect.TypeOf(t)

// 				field := structomancer.NewFieldSpec(st.Field(0), "xyzzy")
// 				Expect(field.ValueIn(t).Interface()).To(BeNil())

// 				field = structomancer.NewFieldSpec(st.Field(1), "xyzzy")
// 				Expect(field.ValueIn(t).Interface()).To(BeNil())
// 			})
// 		})
// 	})

// })
