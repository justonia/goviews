goviews
=======

Type-safe access into generic containers in golang.

TODO
====

There is a lot left to implement, including substruct filling, slices, handling
of nil values via pointers in the struct, and cleanup of the code. 

Motivation
==========

When dealing with schema-less JSON data, usually deserialized into a ``map[string]interface{}``, it became
a pain to gain access to nested data. It also led to ugly code full of type checks, obtuse helper functions,
and endless error checking. This library is meant to be an attempt to ease this pain.

Example
=======
    package main

    import "github.com/justonia/goviews"

    func example() {
        someJson := []byte(`
        {
            "a": {
                "b": {
                    "c": 2000.12354,
                    "d": {
                        "field1": "foobar"
                    },
                    "E": 432.1
                }
            }
        }`)

        data := make(map[string]interface{})
        json.Unmarshal(someJson, &data)
        
        type B struct {
            C      int64              `views:"c,convert"`
            Field1 string             `views:"d.field1"`
            E      views.MutableFloat
        }

        b := B{}
        if err := views.Fill(&b, "a.b", data); err != nil {
            panic(err)
        }

        // b.C == 2000
        // b.Field1 == "foobar"
        // b.E.Get() == 432.1

        b.E.Set(100.34)
        // data["a]["b"]["E"] == 100.34
    }
