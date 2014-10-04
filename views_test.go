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

package views_test

import (
	. "gopkg.in/check.v1"
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

type subStruct struct {
	Field1 *string       `views:"field1"`
	Field2 []interface{} `views:"field2"`
}

type testView struct {
	FieldValue              int64          `views:"a.b.c,convert"`
	FieldReference          *int64         `views:"a.b.c"`
	Struct                  subStruct      `views:"a.b.d"`
	Structs                 []subStruct    `views:"a.b"`
	GenericStructsValue     []interface{}  `views:"a.f"`
	GenericStructsReference *[]interface{} `views:"a.f"`
}
