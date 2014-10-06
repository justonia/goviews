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
	//"fmt"
	. "gopkg.in/check.v1"
	//"reflect"
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

var validData = []byte(`
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
	data := s.getData(validData)
	container, err := getContainer([]string{}, data)
	c.Assert(container, IsNil)
	c.Assert(err, ErrorMatches, ".*empty path.*")

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
	c.Assert(func() { parseTag("") }, PanicMatches, ".*out of range.*")
	//path, name, options := parseTag("")
}

type subStruct struct {
	Field1 *string       `views:"field1"`
	Field2 []interface{} `views:"field2"`
}

type testView struct {
	FieldValue              int64          `views:"a.b.c,convert"`
	FieldReference          *float64       `views:"a.b.c"`
	Struct                  subStruct      `views:"a.b.d"`
	Structs                 []subStruct    `views:"a.e"`
	GenericStructsValue     []interface{}  `views:"a.e"`
	GenericStructsReference *[]interface{} `views:"a.e"`
}

func (s *ViewsSuite) TestTypeField(c *C) {
	/*
	fields := getFields(reflect.TypeOf(testView{}))
	c.Assert(fields, HasLen, 6)
	abcCount := 0
	for _, f := range fields {
		fmt.Println(f)
		switch f.name {
		case "a.b.c":
			c.Assert(f.path, DeepEquals, []string{"a","b","c"})
			c.Assert(f.tag, Equals, true)
			abcCount += 1
		}
	}
	c.Assert(abcCount, Equals, 2)
	*/
}
