
# structomancer

Golang struct reflection package.  Primarily aimed at package authors who want to add struct tag
functionality and/or struct serialization and deserialization to their packages.

## fast

structomancer is pretty fast.  A lot of the reflection calls it makes are only performed once per type, and are fetched from a `sync.RWMutex`-protected cache on subsequent lookups.

## what you can do

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
    // generate empty, addressable instances of a type
    //
    x := z.MakeEmpty()                           // returns a &Blah{}
    structomancer.New(Blah{}, "api").MakeEmpty() // also returns a &Blah

    //
    // validate fields based on struct tags
    //
    z.IsKnownField("inner")        // returns true
    z.IsKnownField("sessionID")    // returns false

    //
    // get field values
    //
    blah := &Blah{Name: "foobar"}
    v, err := z.GetFieldValue(blah, "name")  // returns "foobar", nil

    //
    // set field values
    //
    err := z.SetFieldValue(blah, "name", "qaax")
}
```

## custom decoding/encoding

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


## `reflect` package compatibility

If you're working with lots of `reflect.Value`s already, you probably want to avoid creating even more of them (reflection is apparently expensive because of allocations, although I forget where I read that).

To avoid unnecessary boxing/unboxing of your values, you can access all of structomancer's methods via an alternate `reflect.Value`-compatible interface:

```go
.MakeEmptyV() reflect.Value
.GetFieldValueV(v reflect.Value, fnickname string) (reflect.Value, error)
.SetFieldValueV(sv reflect.Value, fname string, value reflect.Value) error
.StructToMapV(aStruct reflect.Value) (map[string]interface{}, error)
.MapToStructV(fields map[string]interface{}) (reflect.Value, error)
```

Should be substantially faster, but I haven't profiled it yet.


# authors/contributors

- bryn bellomy (<bryn.bellomy@gmail.com>)

