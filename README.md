
# structomancer

Golang struct reflection package.  Primarily aimed at package authors who want to add struct tag
functionality and/or serialization and deserialization to their packages.

```go
import "github.com/brynbellomy/go-structomancer"

type Blah struct {
    Name      string      `api:"name"                 db:"name"`
    Token     int         `api:"token"                db:"-"`
    SessionID string      `api:"-"                    db:"-"`
    Inner     InnerStruct `api:"inner, @tag=weezy"    db:"inner_struct"`
}

type InnerStruct struct {
    Quux string `weezy:"quux"`
}

func main() {
    z := structomancer.New(&Blah{}, "api") // provide a specimen struct and the struct tag

    //
    // deserialize a map to a struct
    //
    s, err := z.MapToStruct(map[string]interface{}{
        "name": "xyzzy",
        "token": 123,
        "inner": map[string]interface{}{
            "quux": "asdf",
        },
    })

    //
    // serialize a struct to a map
    //
    m, err := z.StructToMap(&Blah{
        Name: "xyzzy",
        Token: 123,
        Inner: InnerStruct{Quux: "asdf"},
    })

    //
    // generate empty instances of a type
    //
    x := z.MakeEmpty() // returns a &Blah{}

    //
    // examine individual fields
    //
    z.IsKnownField("inner") // returns true
    z.IsKnownField("sessionID") // returns false

    blah := &Blah{Name: "foobar"}
    z.GetFieldValue(blah, "name")  // returns a reflect.Value containing "foobar"

    z.SetField(blah, "name", "quux")

    // you can .ConvertSetField if the type is a type alias
    type PersonName string
    z.ConvertSetField(blah, "name", PersonName("quux"))


}
```

You might find that you need to set up custom serializer/deserializer functions for individual fields (for example, fields with interface types, which cannot be automatically deserialized by structomancer).

Doing so is easy:

```go
type IFooer interface{ Foo() }

type Blah struct {
    Fooers []IFooer `api:"fooers"`
}

func main() {
    z := structomancer.New(&Blah{}, "api")

    z.SetFieldEncoder("fooers", func(x interface{}) (interface{}, error) {
        fooers := x.([]IFooer)
        // ... convert from []IFooer to []interface{} ...
        return fs, nil
    })

    z.SetFieldDecoder("fooers", func(x interface{}) (interface{}, error) {
        fs := x.([]interface{})
        // ... convert from []interface{} to []IFooer ...
        return fooers, nil
    })

    // now, your (de)serializers will run when you call the following methods:
    z.StructToMap(...)
    z.MapToStruct(...)
}
```