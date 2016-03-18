package structomancer_test

import (
	"reflect"

	"github.com/brynbellomy/go-structomancer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const tagName = "xyzzy"

type (
	Keith struct {
		Name                Name                   `xyzzy:"name"`
		Age                 Age                    `xyzzy:"age"`
		InnerStruct         InnerStruct            `xyzzy:"inner, @tag=weezy"`
		StructSlice         []InnerStruct          `xyzzy:"structSlice, @tag=weezy"`
		InterfaceSlice      []stringer             `xyzzy:"interfaceSlice"`
		MapOfStructs        map[string]InnerStruct `xyzzy:"mapOfStructs, @tag=weezy"`
		MapOfBools          map[string]bool        `xyzzy:"mapOfBools"`
		NestedInlineStructs []struct {
			Qaax string `weezy:"qaax"`
		} `xyzzy:"nested, @tag=weezy"`
	}

	InnerStruct struct {
		Foo string `weezy:"foo"`
		Bar []B    `weezy:"bar"`
	}

	Name string
	Age  uint64
	B    uint64

	stringer interface {
		String() string
	}
)

func (i InnerStruct) String() string {
	return i.Foo
}

var _ = Describe("Structomancer", func() {

	Context("when .GetFieldValue is called", func() {
		It("should serialize a struct field value to the appropriate native type", func() {
			z := structomancer.New(&Keith{}, "xyzzy")
			x := &Keith{Name: Name("keith")}

			fv, err := z.GetFieldValue(x, "name")
			if err != nil {
				Fail(err.Error())
			}

			Expect(fv.Interface()).To(Equal("keith"))
		})

		Context("with a nil struct argument", func() {
			It("should return an error", func() {
				z := structomancer.New(&Keith{}, "xyzzy")

				var aStruct *Keith
				type IBlah interface{}
				var anInterface IBlah = aStruct

				_, err := z.GetFieldValue(aStruct, "name")
				Expect(err).To(Not(BeNil()))

				_, err = z.GetFieldValue(anInterface, "name")
				Expect(err).To(Not(BeNil()))
			})
		})
	})

	Context("when .SetFieldValue is called", func() {
		It("should deserialize a native value to the type of the given struct field", func() {
			z := structomancer.New(&Keith{}, "xyzzy")
			x := &Keith{}

			err := z.SetFieldValue(x, "name", reflect.ValueOf("keith"))
			if err != nil {
				Fail(err.Error())
			}

			Expect(x.Name).To(Equal(Name("keith")))
		})
	})

	Context("when .MakeEmpty is called", func() {
		It("should summon a pointer to a zeroed struct of the correct type", func() {
			specimens := []interface{}{
				Keith{},
				&Keith{},
			}

			for _, spec := range specimens {
				z := structomancer.New(spec, tagName)
				x := z.MakeEmpty()

				Expect(reflect.ValueOf(x).CanAddr()).To(BeFalse())
				Expect(reflect.ValueOf(x).CanSet()).To(BeFalse())

				Expect(reflect.ValueOf(x).Elem().CanAddr()).To(BeTrue())
				Expect(reflect.ValueOf(x).Elem().CanSet()).To(BeTrue())
			}
		})
	})

	Context("when .MapToStruct is called", func() {
		var (
			theMap = map[string]interface{}{
				"name": "keith richards",
				"age":  429,
				"inner": map[string]interface{}{
					"foo": "xyzzy",
					"bar": []interface{}{5, 6, 7},
				},
				"structSlice": []interface{}{
					map[string]interface{}{
						"foo": "xyzzy",
						"bar": []interface{}{5, 6, 7},
					},
				},
				"interfaceSlice": []interface{}{
					map[string]interface{}{
						"foo": "xyzzy",
						"bar": []interface{}{5, 6, 7},
					},
				},
				"mapOfStructs": map[string]interface{}{
					"some-key": map[string]interface{}{"foo": "xyzzy", "bar": []interface{}{5, 6, 7}},
				},
				"mapOfBools": map[string]interface{}{"one": true, "two": false},
				"nested": []interface{}{
					map[string]interface{}{"qaax": "one"},
					map[string]interface{}{"qaax": "two"},
				},
			}

			specimens = []interface{}{
				Keith{},
				&Keith{},
			}

			expected = []interface{}{
				Keith{
					Name("keith richards"),
					Age(429),
					InnerStruct{"xyzzy", []B{5, 6, 7}},
					[]InnerStruct{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					[]stringer{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					map[string]InnerStruct{
						"some-key": {"xyzzy", []B{5, 6, 7}},
					},
					map[string]bool{"one": true, "two": false},
					[]struct {
						Qaax string `weezy:"qaax"`
					}{
						{"one"}, {"two"},
					},
				},
				&Keith{
					Name("keith richards"),
					Age(429),
					InnerStruct{"xyzzy", []B{5, 6, 7}},
					[]InnerStruct{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					[]stringer{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					map[string]InnerStruct{
						"some-key": {"xyzzy", []B{5, 6, 7}},
					},
					map[string]bool{"one": true, "two": false},
					[]struct {
						Qaax string `weezy:"qaax"`
					}{
						{"one"}, {"two"},
					},
				},
			}

			decodeInterfaceSlice = func(val interface{}) (interface{}, error) {
				slice := val.([]interface{})
				decoded := make([]stringer, len(slice))
				innerZ := structomancer.New(InnerStruct{}, "weezy")

				for i, x := range slice {
					m := x.(map[string]interface{})
					s, err := innerZ.MapToStruct(m)
					if err != nil {
						return nil, err
					}
					decoded[i] = s.(InnerStruct)
				}
				return decoded, nil
			}
		)

		It("should deserialize the contents of the map into the appropriate fields in the returned struct", func() {
			for i, s := range specimens {
				z := structomancer.New(s, tagName)
				z.SetFieldDecoder("interfaceSlice", decodeInterfaceSlice)

				k, err := z.MapToStruct(theMap)
				if err != nil {
					Fail(err.Error())
				}
				Expect(k).To(Equal(expected[i]))
			}
		})
	})

	Context("when .StructToMap is called", func() {
		var (
			specimens = []interface{}{
				Keith{
					Name("keith richards"),
					Age(429),
					InnerStruct{"xyzzy", []B{5, 6, 7}},
					[]InnerStruct{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					[]stringer{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					map[string]InnerStruct{"some-key": {"xyzzy", []B{5, 6, 7}}},
					map[string]bool{"one": true, "two": false},
					[]struct {
						Qaax string `weezy:"qaax"`
					}{
						{"one"}, {"two"},
					},
				},
				&Keith{
					Name("keith richards"),
					Age(429),
					InnerStruct{"xyzzy", []B{5, 6, 7}},
					[]InnerStruct{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					[]stringer{InnerStruct{"xyzzy", []B{5, 6, 7}}},
					map[string]InnerStruct{"some-key": {"xyzzy", []B{5, 6, 7}}},
					map[string]bool{"one": true, "two": false},
					[]struct {
						Qaax string `weezy:"qaax"`
					}{
						{"one"}, {"two"},
					},
				},
			}

			expected = map[string]interface{}{
				"name": "keith richards",
				"age":  uint64(429),
				"inner": map[string]interface{}{
					"foo": "xyzzy",
					"bar": []interface{}{uint64(5), uint64(6), uint64(7)},
				},
				"structSlice": []interface{}{
					map[string]interface{}{"foo": "xyzzy", "bar": []interface{}{5, 6, 7}},
				},
				"interfaceSlice": []interface{}{
					map[string]interface{}{"foo": "xyzzy", "bar": []interface{}{5, 6, 7}},
				},
				"mapOfStructs": map[string]interface{}{
					"some-key": map[string]interface{}{"foo": "xyzzy", "bar": []interface{}{5, 6, 7}},
				},
				"mapOfBools": map[string]interface{}{"one": true, "two": false},
				"nested": []interface{}{
					map[string]interface{}{"qaax": "one"},
					map[string]interface{}{"qaax": "two"},
				},
			}

			encodeInterfaceSlice = func(val interface{}) (interface{}, error) {
				slice := val.([]stringer)
				encoded := make([]interface{}, len(slice))
				for i, x := range slice {
					innerZ := structomancer.New(x, "weezy")
					m, err := innerZ.StructToMap(x)
					if err != nil {
						return nil, err
					}
					encoded[i] = m
				}
				return encoded, nil
			}
		)

		It("should serialize the contents of the struct into the appropriate keys in the returned map", func() {
			for _, spec := range specimens {
				z := structomancer.New(spec, "xyzzy")
				z.SetFieldEncoder("interfaceSlice", encodeInterfaceSlice)

				k, err := z.StructToMap(spec)
				if err != nil {
					Fail(err.Error())
				}

				// Expect(k).To(Equal(expected))

				Expect(k["name"]).To(Equal(expected["name"]))
				Expect(k["age"]).To(Equal(expected["age"]))
				Expect(k["inner"]).To(Equal(expected["inner"]))

				Expect(k["structSlice"]).To(HaveLen(len(expected["structSlice"].([]interface{}))))
				for i, x := range k["structSlice"].([]interface{}) {
					innerK := x.(map[string]interface{})
					innerExpected := (expected["structSlice"].([]interface{}))[i].(map[string]interface{})
					Expect(innerK["foo"]).To(Equal(innerExpected["foo"]))
				}

				Expect(k["interfaceSlice"]).To(HaveLen(len(expected["interfaceSlice"].([]interface{}))))
				for i, x := range k["interfaceSlice"].([]interface{}) {
					innerK := x.(map[string]interface{})
					innerExpected := (expected["interfaceSlice"].([]interface{}))[i].(map[string]interface{})
					Expect(innerK["foo"]).To(Equal(innerExpected["foo"]))
				}
			}
		})
	})
})
