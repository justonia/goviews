// The MIT License (MIT)
//
// Copyright (c) 2014 Justin Larrabee
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package views

import (
	"encoding/json"
	"fmt"
	. "gopkg.in/check.v1"
	"reflect"
	"strings"
	"testing"
)

func TestViews(t *testing.T) { TestingT(t) }

type ViewsSuite struct {
}

var _ = Suite(&ViewsSuite{})

func (s *ViewsSuite) SetUpTest(c *C) {
}

func (s *ViewsSuite) TearDownTest(c *C) {
}

func (s *ViewsSuite) getData(bytes []byte) map[string]interface{} {
	out := make(map[string]interface{})
	var err error
	err = json.Unmarshal(bytes, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func (s *ViewsSuite) TestGetContainer(c *C) {
	validData := []byte(`
	{
		"a": {
			"b": {
				"c": 2000,
				"d": {
					"field1": "foobar",
					"field2": [
						"bar",
						10,
						20
					]
				},
				"e": [{
					"field1": "asdf"
				},{
					"field2": ["1", "2"]
				}],
				"f": [

				]
			}
		}
	}`)

	data := s.getData(validData)
	container, err := getContainer([]string{}, data)
	c.Assert(container, DeepEquals, data)
	c.Assert(err, IsNil)

	container, err = getContainer([]string{"b", "c"}, data)
	c.Assert(container, IsNil)
	c.Assert(err, ErrorMatches, ".*no such key 'b' at index 0.*")

	container, err = getContainer([]string{"a", "b", "c", "d"}, data)
	c.Assert(container, IsNil)
	c.Assert(err, ErrorMatches, ".*for key 'c' at index 2 .* expected map\\[string\\]interface.*")

	container, err = getContainer([]string{"a", "b"}, data)
	c.Assert(container, DeepEquals, data["a"].(map[string]interface{})["b"])
	c.Assert(err, IsNil)

	container, err = getContainer([]string{"a", "b", "d"}, data)
	c.Assert(container["field1"], FitsTypeOf, "")
	c.Assert(container["field1"].(string), Equals, "foobar")
	c.Assert(err, IsNil)
}

func (s *ViewsSuite) TestParseTag(c *C) {
	name, path, options := parseTag("")
	c.Assert(name, Equals, "")
	c.Assert(path, HasLen, 0)
	c.Assert(options, Equals, tagOptions(""))

	name, path, options = parseTag("foo")
	c.Assert(name, Equals, "foo")
	c.Assert(path, HasLen, 0)
	c.Assert(options, Equals, tagOptions(""))

	name, path, options = parseTag("foo.bar.baz")
	c.Assert(name, Equals, "baz")
	c.Assert(path, HasLen, 2)
	c.Assert(path, DeepEquals, []string{"foo", "bar"})
	c.Assert(options, Equals, tagOptions(""))

	name, path, options = parseTag("foo.bar.baz,convert")
	c.Assert(name, Equals, "baz")
	c.Assert(path, HasLen, 2)
	c.Assert(path, DeepEquals, []string{"foo", "bar"})
	c.Assert(options, Equals, tagOptions("convert"))
	c.Assert(options.Contains("convert"), Equals, true)
	c.Assert(options.Contains("blah"), Equals, false)

	name, path, options = parseTag("foo.bar.baz,convert,omit")
	c.Assert(options, Equals, tagOptions("convert,omit"))
	c.Assert(options.Contains("convert"), Equals, true)
	c.Assert(options.Contains("omit"), Equals, true)
}

type subStruct struct {
	Field1 *string       `views:"field1"`
	Field2 []interface{} `views:"field2"`
}

type StringTypedef string

type testView struct {
	FieldFloatConvertValue   int64        `views:"a.b.c,convert"`
	FieldFloatValue          float64      `views:"a.b.c"`
	FieldFloatMutable        MutableFloat `views:"a.b.c"`
	FieldStringValue         string
	FieldStringTypedef       StringTypedef `views:"FieldStringValue,convert"`
	FieldStringValueIgnore   string        `views:"-"`
	FieldStringValueOptional string        `views:",optional"`
	FieldStringMutableValue  MutableString `views:"FieldStringValue"`
	FieldBoolValue           bool          `views:"a.b.f"`
	//FieldBoolMutable        MutableBool `views:"a.b.c"`
	Struct                  subStruct      `views:"a.b.d"`
	Structs                 []subStruct    `views:"a.e"`
	GenericStructsValue     []interface{}  `views:"a.e"`
	GenericStructsReference *[]interface{} `views:"a.e"`
}

func (s *ViewsSuite) TestTypeField(c *C) {
	fields := getFields(reflect.TypeOf(testView{}))
	//c.Assert(fields, HasLen, 8)
	abcCount := 0
	aeCount := 0
	for _, f := range fields {
		switch strings.Join(append(f.path, f.name), ".") {
		case "a.b.c":
			c.Assert(f.path, DeepEquals, []string{"a", "b"})
			c.Assert(f.tag, Equals, true)
			switch f.typ.Kind() {
			case reflect.Int64:
				c.Assert(f.convert, Equals, true)
			case reflect.Float64:
				c.Assert(f.convert, Equals, false)
			case reflect.Interface:
				c.Assert(f.typ.Implements(reflect.TypeOf((*MutableFloat)(nil)).Elem()), Equals, true)
			case reflect.Bool:
			default:
				panic(fmt.Sprintf("unknown typ field '%s' for a.b.c", f.typ))
			}
			abcCount += 1

		case "a.e":
			aeCount += 1

		case "FieldStringValue":
		}
	}
	c.Assert(aeCount, Equals, 3)
	c.Assert(abcCount, Equals, 3)

	fields = getFields(reflect.TypeOf(subStruct{}))
	c.Assert(fields, HasLen, 2)
	for _, f := range fields {
		switch f.name {
		case "field1":
			c.Assert(f.typ, Equals, reflect.TypeOf(""))
		case "field2":
			c.Assert(f.typ, Equals, reflect.TypeOf([]interface{}{}))
		default:
			panic(fmt.Sprintf("unknown subStruct field %s", f.name))
		}
	}
}

func (s *ViewsSuite) TestFillFromMap(c *C) {
	validData := []byte(`
	{
		"a": {
			"b": {
				"c": 2000,
				"d": {
					"field1": "foobar",
					"field2": [
						"bar",
						10,
						20
					]
				},
				"f": true
			},
			"e": [{
				"field1": "asdf"
			},{
				"field2": ["1", "2"]
			}],
			"g": {
				"f": false,
				"a": { "e": [], "b": { "c": 5000, "d":[], "f":true  } } ,
				"FieldStringValue": "barbaz"
			}
		},
		"FieldStringValue": "foobar",
		"FieldStringValueOptional": "foobar2"
	}`)
	data := s.getData(validData)
	out := testView{}
	err := fillFromMap(&out, []string{}, data)
	c.Assert(err, IsNil)
	c.Assert(out.FieldFloatConvertValue, Equals, int64(2000))
	c.Assert(out.FieldFloatValue, Equals, float64(2000))
	c.Assert(out.FieldFloatMutable.Get(), Equals, float64(2000))
	c.Assert(out.FieldStringValue, Equals, "foobar")
	c.Assert(out.FieldStringTypedef, Equals, StringTypedef("foobar"))
	c.Assert(out.FieldStringValueOptional, Equals, "foobar2")
	c.Assert(out.FieldStringValueIgnore, Equals, "")
	c.Assert(out.FieldBoolValue, Equals, true)

	// try changing a value inside data
	out.FieldFloatMutable.Set(float64(4000))
	c.Assert(out.FieldFloatMutable.Get(), Equals, float64(4000))
	out = testView{}
	err = fillFromMap(&out, []string{}, data)
	c.Assert(err, IsNil)
	c.Assert(out.FieldFloatMutable.Get(), Equals, float64(4000))

	data = s.getData(validData)
	out = testView{}
	err = fillFromMap(&out, []string{"a", "g"}, data)
	c.Assert(err, IsNil)
	c.Assert(out.FieldStringValue, Equals, "barbaz")
	c.Assert(out.FieldFloatConvertValue, Equals, int64(5000))
	c.Assert(out.FieldFloatValue, Equals, float64(5000))
	c.Assert(out.FieldStringValueOptional, Equals, "")
	c.Assert(out.FieldStringValueIgnore, Equals, "")
}

func (s *ViewsSuite) TestFillFromMapBad(c *C) {
	type UnknownInterface interface {
		Bar()
	}
	type foo struct {
		Barr UnknownInterface
	}
	validData := []byte(`
	{
		"Barr": 5
	}`)
	data := s.getData(validData)
	out := foo{}
	err := fillFromMap(&out, []string{}, data)
	c.Assert(err, ErrorMatches, ".*could not set unknown view interface 'views.UnknownInterface'.*")

	type foo2 struct {
		NotAFloat MutableFloat
	}
	validData = []byte(`
	{
		"NotAFloat": "1234"
	}`)
	data = s.getData(validData)
	out2 := foo2{}
	err = fillFromMap(&out2, []string{}, data)
	c.Assert(err, ErrorMatches, ".*cannot assign 'string' to 'views.MutableFloat' at path '.NotAFloat' in struct of type views.foo2.*")

	type foo3 struct {
		NotAString MutableString
	}
	validData = []byte(`
	{
		"NotAString": 1234
	}`)
	data = s.getData(validData)
	out3 := foo3{}
	err = fillFromMap(&out3, []string{}, data)
	c.Assert(err, ErrorMatches, ".*cannot assign 'float64' to 'views.MutableString' at path '.NotAString' in struct of type views.foo3.*")

	type foo4 struct {
		FieldBoolConvertValue bool `views:",convert"`
	}
	validData = []byte(`
	{
		"FieldBoolConvertValue": 1234
	}`)
	data = s.getData(validData)
	out4 := foo4{}
	err = fillFromMap(&out4, []string{}, data)
	c.Assert(err, ErrorMatches, ".*cannot assign or convert 'float64' to 'bool'.*")

}

func (s *ViewsSuite) TestExample(c *C) {
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
        E      MutableFloat
    }

    b := B{}
	if err := Fill(&b, "a.b", data); err != nil {
		panic(err)
	}

    c.Assert(b.C, Equals, int64(2000))
    c.Assert(b.Field1, Equals, "foobar")
    c.Assert(b.E.Get(), Equals, 432.1)

    b.E.Set(100.34)
	c.Assert(b.E.Get(), Equals, 100.34)
	container, err := getContainer([]string{"a","b"}, data)
	c.Assert(err, IsNil)
	f, ok := container["E"].(float64)
	c.Assert(ok, Equals, true)
	c.Assert(f, Equals, 100.34)
	
    // data["a]["b"]["E"] == 100.34
}
