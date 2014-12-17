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

// Package views implements a type-safe method for accessing data stored in a
// generic container such as map[string]interface{}

package views

import (
	"fmt"
	"reflect"
	"strings"
)

type ViewError struct {
	Reason string
}

func (v ViewError) Error() string {
	return fmt.Sprintf("view error - %s", v.Reason)
}

type MutableFloat interface {
	Set(float64)
	Get() float64
	GetChecked() (float64, bool)
}

type MutableString interface {
	Set(string)
	Get() string
	GetChecked() (string, bool)
}

var mutableFloatType = reflect.TypeOf((*MutableFloat)(nil)).Elem()
var mutableStringType = reflect.TypeOf((*MutableString)(nil)).Elem()

func Fill(out interface{}, basePath interface{}, in map[string]interface{}) error {
	switch basePath.(type) {
	case string:
		return fillFromMap(out, strings.Split(basePath.(string), "."), in)
	case []string:
		return fillFromMap(out, basePath.([]string), in)
	default:
		panic(fmt.Sprintf("bad argument type to views.Fill '%s'", reflect.TypeOf(basePath)))
	}
}

func fillFromMap(out interface{}, basePath []string, in map[string]interface{}) error {
	var container map[string]interface{}
	var err error

	if len(basePath) == 0 || basePath[0] == "" {
		container = in
	} else {
		container, err = getContainer(basePath, in)
	}
	if err != nil {
		return err
	}

	outValue := reflect.ValueOf(out)
	if outValue.Kind() == reflect.Ptr {
		outValue = outValue.Elem()
	}
	outFields := getFields(outValue.Type())
	for _, field := range outFields {
		var (
			fieldContainer map[string]interface{}
			err            error
			ok             bool
			v              interface{}
		)
		if fieldContainer, err = getContainer(field.path, container); err != nil {
			return err
		}
		if v, ok = fieldContainer[field.name]; !ok {
			if field.optional {
				continue
			}
			return ViewError{fmt.Sprintf("could not find %s.%s in container", strings.Join(field.path, "."), field.name)}
		}

		vValue := reflect.ValueOf(v)
		vType := vValue.Type()
		fieldOutValue := outValue.FieldByIndex(field.index)
		fieldOutType := fieldOutValue.Type()

		switch fieldOutValue.Kind() {
		case reflect.Interface:
			switch {
			case fieldOutType.Implements(mutableFloatType):
				if _, ok := v.(float64); ok {
					fieldOutValue.Set(reflect.ValueOf(floatMapMutator{fieldContainer, field.name}))
				} else {
					return ViewError{fmt.Sprintf("cannot assign '%s' to '%s' at path '%s.%s' in struct of type %s", vType, fieldOutType, strings.Join(field.path, "."), field.name, outValue.Type())}
				}
			case fieldOutType.Implements(mutableStringType):
				if _, ok := v.(string); ok {
					fieldOutValue.Set(reflect.ValueOf(stringMapMutator{fieldContainer, field.name}))
				} else {
					return ViewError{fmt.Sprintf("cannot assign '%s' to '%s' at path '%s.%s' in struct of type %s", vType, fieldOutType, strings.Join(field.path, "."), field.name, outValue.Type())}
				}
			default:
				return ViewError{fmt.Sprintf("could not set unknown view interface '%s' to at path '%s.%s'", fieldOutType, strings.Join(field.path, "."), field.name)}
			}
			continue
		case reflect.Struct:
			// todo, recurse
			continue
		case reflect.Slice:
			// todo
			continue
		case reflect.Ptr:
			// todo
			continue

		default:
			assignable := vType.AssignableTo(fieldOutType)

			if !fieldOutValue.CanSet() {
				// This should be caught in getFields below
				panic(ViewError{fmt.Sprintf("cannot set '%s' to at path '%s.%s'", fieldOutType, strings.Join(field.path, "."), field.name)})
			} else if !assignable && field.convert && vType.ConvertibleTo(fieldOutType) {
				// convert below
			} else if !assignable {
				return ViewError{fmt.Sprintf("cannot assign or convert '%s' to '%s' at path '%s.%s' in struct of type %s", vType, fieldOutType, strings.Join(field.path, "."), field.name, outValue.Type())}
			}

			if !assignable {
				vValue = vValue.Convert(fieldOutType)
			} else if field.isPtr {
				vValue = reflect.ValueOf(&v)
			}
			fieldOutValue.Set(vValue)
		}
	}

	return nil
}

func getContainer(path []string, container map[string]interface{}) (map[string]interface{}, error) {
	if len(path) == 0 {
		return container, nil
	}
	outContainer := container
	var ok bool
	for i, key := range path {
		var mapValue interface{}
		if mapValue, ok = outContainer[key]; !ok {
			return nil, ViewError{fmt.Sprintf("no such key '%s' at index %d in path '%s'", key, i, strings.Join(path, "."))}
		}
		if outContainer, ok = mapValue.(map[string]interface{}); !ok {
			return nil, ViewError{fmt.Sprintf("for key '%s' at index %d in path '%s', expected map[string]interface{} not %s", key, i, strings.Join(path, "."), reflect.TypeOf(mapValue))}
		}
	}
	return outContainer, nil
}

// A field represents a single field found in a struct.
type field struct {
	name string
	path []string

	tag            bool
	index          []int
	typ            reflect.Type
	isPtr          bool
	convert        bool
	optional       bool
	mutatorFactory interface{} // should be castable to a specific mutator based on typ and the container type
}

type floatMapMutator struct {
	container map[string]interface{}
	key       string
}

func (m floatMapMutator) Get() float64 {
	if v, ok := m.container[m.key].(float64); ok {
		return v
	}
	return 0.0
}
func (m floatMapMutator) GetChecked() (float64, bool) {
	v, ok := m.container[m.key].(float64)
	return v, ok
}
func (m floatMapMutator) Set(value float64) {
	m.container[m.key] = value
}

type stringMapMutator struct {
	container map[string]interface{}
	key       string
}

func (m stringMapMutator) Get() string {
	if v, ok := m.container[m.key].(string); ok {
		return v
	}
	return ""
}
func (m stringMapMutator) GetChecked() (string, bool) {
	v, ok := m.container[m.key].(string)
	return v, ok
}
func (m stringMapMutator) Set(value string) {
	m.container[m.key] = value
}

// This is based off of encoding/json
func getFields(t reflect.Type) []field {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	// Types already visited and fields collected
	visited := map[reflect.Type]bool{}
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, typeField := range current {
			if visited[typeField.typ] {
				continue
			}
			visited[typeField.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < typeField.typ.NumField(); i++ {
				structField := typeField.typ.Field(i)
				if structField.PkgPath != "" { // unexported
					continue
				}

				tag := structField.Tag.Get("views")
				if tag == "-" {
					continue
				}

				name, path, opts := parseTag(tag)
				if !isValidTag(name) {
					name = ""
				}

				index := make([]int, len(typeField.index)+1)
				copy(index, typeField.index)
				index[len(typeField.index)] = i

				structFieldType := structField.Type
				isPtr := structFieldType.Kind() == reflect.Ptr
				if structFieldType.Name() == "" && isPtr {
					// Follow pointer.
					structFieldType = structFieldType.Elem()
					isPtr = true
				}

				// Record found field and index sequence.
				if name != "" || !structField.Anonymous || structFieldType.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = structField.Name
						path = []string{}
					}
					fields = append(fields, field{
						name:     name,
						path:     path,
						tag:      tagged,
						index:    index,
						typ:      structFieldType,
						convert:  opts.Contains("convert"),
						optional: opts.Contains("optional"),
						isPtr:    isPtr,
						//mutator:  makeMutator(structFieldType),
					})
					if count[typeField.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[structFieldType]++
				if nextCount[structFieldType] == 1 {
					next = append(next, field{name: structFieldType.Name(), path: path, index: index, typ: structFieldType})
				}
			}
		}
	}

	// TODO: Sort by path depth so that we don't recurse into the map structure more
	// than necessary.
	return fields
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	return true
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's views tag into its name and
// comma-separated options.
func parseTag(tag string) (string, []string, tagOptions) {
	if len(tag) == 0 {
		return "", []string{}, tagOptions("")
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		path := strings.Split(tag[:idx], ".")
		return path[len(path)-1], path[:len(path)-1], tagOptions(tag[idx+1:])
	}
	path := strings.Split(tag, ".")
	return path[len(path)-1], path[:len(path)-1], tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}
