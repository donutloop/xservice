// Copyright 2018 XService, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package types

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
)

// Import represent an individual imported package.
type Import interface {
	// GetPackage returns the go import package, like reflect.Type.PkgPath()
	GetPackage() string
	// GetAlias returns an alias string to refer to the import package, or the
	// empty string to omit an import alias.
	GetAlias() string
}

// TypeReference represent a specific reference (either an interface, function, struct or global)
type TypeReference interface {
	// GetImports returns the imports required to use this type. A struct, for example,
	// collects all the imports for its fields and itself.
	GetImports() []Import
	// GetName returns the go-syntax name of the type.
	GetName() string
}

// ImportSpec implements Import to represent an imported go package
type ImportSpec struct {
	Package   string
	Alias     string
	Qualified bool
}

// getQualifier returns the fully qualified package (e.g. bytes.) for use in a qualified
// declared type
func (i *ImportSpec) getQualifier() string {
	if !i.Qualified {
		return ""
	}

	var buff bytes.Buffer
	if i.Alias != "" {
		buff.WriteString(i.Alias)
	} else {
		// the package may contain slashes, so only write the base name of the package,
		// not the full package
		buff.WriteString(path.Base(i.Package))
	}
	buff.WriteString(".")

	return buff.String()
}

// GetAlias returns the alias associated with the package
func (i *ImportSpec) GetAlias() string {
	return i.Alias
}

// GetPackage returns the package
func (i *ImportSpec) GetPackage() string {
	return i.Package
}

// UnqualifiedPrefix The prefix for type aliases that will be interpreted as unqualified
const UnqualifiedPrefix = "_unqualified"

func stringPointerFunc() *string {
	s := ""
	return &s
}

func boolPointerFunc() *bool {
	b := false
	return &b
}

func intPointerFunc() *int {
	n := 0
	return &n
}

func int8PointerFunc() *int8 {
	n := int8(0)
	return &n
}

func int16PointerFunc() *int16 {
	n := int16(0)
	return &n
}

func int32PointerFunc() *int32 {
	n := int32(0)
	return &n
}

func int64PointerFunc() *int64 {
	n := int64(0)
	return &n
}

func uintPointerFunc() *uint {
	n := uint(0)
	return &n
}

func uint8PointerFunc() *uint8 {
	n := uint8(0)
	return &n
}

func uint16PointerFunc() *uint16 {
	n := uint16(0)
	return &n
}

func uint32PointerFunc() *uint32 {
	n := uint32(0)
	return &n
}

func uint64PointerFunc() *uint64 {
	n := uint64(0)
	return &n
}

var (
	// String A TypeReference for string
	String = TypeReferenceFromInstance("")
	// String A TypeReference for string pointer
	StringPointer = TypeReferenceFromInstance(stringPointerFunc())
	// Bool A TypeReference for bool
	Bool = TypeReferenceFromInstance(false)
	// Bool A TypeReference for bool pointer
	BoolPointer = TypeReferenceFromInstance(boolPointerFunc())
	// Int A TypeReference for int
	Int = TypeReferenceFromInstance(0)
	// Int A TypeReference for int
	IntPointer = TypeReferenceFromInstance(intPointerFunc())
	// Int8 A TypeReference for int8
	Int8 = TypeReferenceFromInstance(int8(0))
	// Int8 A TypeReference for int8 pointer
	Int8Pointer = TypeReferenceFromInstance(int8PointerFunc)
	// Int16 A TypeReference for int16
	Int16 = TypeReferenceFromInstance(int16(0))
	// Int16 A TypeReference for int16 pointer
	Int16Pointer = TypeReferenceFromInstance(int16PointerFunc)
	// Int32 A TypeReference for int32
	Int32 = TypeReferenceFromInstance(int32(0))
	// Int32 A TypeReference for int32 pointer
	Int32Pointer = TypeReferenceFromInstance(int32PointerFunc)
	// Int64 A TypeReference for int64
	Int64 = TypeReferenceFromInstance(int64(0))
	// Int64 A TypeReference for int64 pointer
	Int64Pointer = TypeReferenceFromInstance(int64PointerFunc())
	// Int A TypeReference for int
	Uint = TypeReferenceFromInstance(uint(0))
	// Int A TypeReference for int
	UintPointer = TypeReferenceFromInstance(uintPointerFunc())
	// Int8 A TypeReference for int8
	Uint8 = TypeReferenceFromInstance(uint8(0))
	// Int8 A TypeReference for int8 pointer
	Uint8Pointer = TypeReferenceFromInstance(uint8PointerFunc)
	// Int16 A TypeReference for int16
	Uint16 = TypeReferenceFromInstance(uint16(0))
	// Int16 A TypeReference for int16 pointer
	Uint16Pointer = TypeReferenceFromInstance(uint16PointerFunc)
	// Int32 A TypeReference for int32
	Uint32 = TypeReferenceFromInstance(uint32(0))
	// Int32 A TypeReference for int32 pointer
	Uint32Pointer = TypeReferenceFromInstance(uint32PointerFunc)
	// Int64 A TypeReference for int64
	Uint64 = TypeReferenceFromInstance(uint64(0))
	// Int64 A TypeReference for int64 pointer
	Uint64Pointer = TypeReferenceFromInstance(uint64PointerFunc())

	// todo missing pointers
	// Uintptr A TypeReference for uintptr
	Uintptr = TypeReferenceFromInstance(uintptr(0))
	// Float32 A TypeReference for float32
	Float32 = TypeReferenceFromInstance(float32(0))
	// Float64 A TypeReference for float64
	Float64 = TypeReferenceFromInstance(float64(0))
	// Complex64 A TypeReference for complex64
	Complex64 = TypeReferenceFromInstance(complex64(0))
	// Complex128 A TypeReference for complex128
	Complex128 = TypeReferenceFromInstance(complex128(0))
	// Byte A TypeReference for byte
	Byte = TypeReferenceFromInstanceWithCustomName(uint8(0), "byte")
	// Rune A TypeReference for rune
	Rune = TypeReferenceFromInstanceWithCustomName(int32(0), "rune")
	// Error A TypeReference for error
	Error = TypeReferenceFromInstanceWithCustomName(errors.New(""), "error")
)

func UnsafePointerReference(p string) string {
	return fmt.Sprintf("*%s", p)
}

type Parameter struct {
	NameOfParameter string
	Typ             TypeReference
}

func NewParameterWithTypeReference(name string, t TypeReference) *Parameter {
	return &Parameter{
		NameOfParameter: name,
		Typ:             t,
	}
}

func NewParameterWithTypeReferenceFromInstance(name string, t interface{}) *Parameter {
	return &Parameter{
		NameOfParameter: name,
		Typ:             TypeReferenceFromInstance(t),
	}
}

func NewParameterWithUnsafeTypeReference(name, t string) *Parameter {
	return &Parameter{
		NameOfParameter: name,
		Typ:             NewUnsafeTypeReference(t),
	}
}

// TypeReferenceFromInstance creates a TypeReference from an instance of a variable
func TypeReferenceFromInstance(t interface{}) TypeReference {
	return newTypeReferenceFromInstance(t, "")
}

// NewUnsafeTypeReference creates a (Unsafe) TypeReference from an string id
func NewUnsafeTypeReference(t string) TypeReference {
	return &unsafeTypeReferenceValue{Name: t}
}

// TypeReferenceFromInstanceWithAlias creates a TypeReference from an instance of a variable
// with the given package alias
func TypeReferenceFromInstanceWithAlias(t interface{}, alias string) TypeReference {
	return newTypeReferenceFromInstance(t, alias)
}

// TypeReferenceFromInstanceWithCustomName creates a TypeReference from an instance of a variable
// with the given custom name, for use of a type alias's name rather than the underlying
// reflect type.
func TypeReferenceFromInstanceWithCustomName(t interface{}, name string) TypeReference {
	typeRef := &typeReferenceWithCustomName{
		TypeReference: newTypeReferenceFromInstance(t, ""),
		name:          name,
	}

	return typeRef
}

type typeReferenceWithCustomName struct {
	TypeReference
	name string
}

func (t *typeReferenceWithCustomName) GetName() string {
	return t.name
}

func newTypeReferenceFromInstance(t interface{}, alias string) TypeReference {
	reflectType := reflect.TypeOf(t)
	if reflectType == nil {
		panic("Invalid nil instance without associated type")
	}

	if reflectType.Kind() == reflect.Func {
		return newTypeReferenceFromFunction(t, alias)
	}

	return newTypeReferenceFromValue(t, alias)
}

type typeReferenceMap struct {
	KeyType   TypeReference
	ValueType TypeReference
	prefix    string
}

func newTypeReferenceFromMap(t interface{}, prefix string) TypeReference {
	refType := reflect.TypeOf(t)

	return &typeReferenceMap{
		KeyType:   newTypeReferenceFromInstance(reflect.New(refType.Key()).Elem().Interface(), ""),
		ValueType: newTypeReferenceFromInstance(reflect.New(refType.Elem()).Elem().Interface(), ""),
		prefix:    prefix,
	}
}

func (t *typeReferenceMap) GetImports() []Import {
	imports := make([]Import, 0)
	imports = append(imports, t.KeyType.GetImports()...)
	imports = append(imports, t.ValueType.GetImports()...)
	return imports
}

func (t *typeReferenceMap) GetName() string {
	return fmt.Sprintf("%smap[%s]%s", t.prefix, t.KeyType.GetName(), t.ValueType.GetName())
}

type typeReferenceValue struct {
	Import *ImportSpec
	Name   string
	prefix string
}

func newTypeReferenceFromValue(t interface{}, alias string) TypeReference {
	refType := reflect.TypeOf(t)
	result := new(typeReferenceValue)
	result.prefix, refType = dereferenceType("", refType)

	switch refType.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Struct:
		result.Import = &ImportSpec{
			Qualified: !strings.HasPrefix(refType.Name(), UnqualifiedPrefix),
			Package:   refType.PkgPath(),
			Alias:     alias,
		}
	case reflect.Map:
		return newTypeReferenceFromMap(reflect.New(refType).Elem().Interface(), result.prefix)
	}

	result.Name = strings.TrimPrefix(refType.Name(), UnqualifiedPrefix)

	return result
}

func dereferenceType(prefix string, refType reflect.Type) (string, reflect.Type) {
	for {
		if refType.Kind() == reflect.Ptr {
			refType = refType.Elem()
			// interfaces are already pointers, so don't need to add prefix
			if refType.Kind() != reflect.Interface {
				prefix += "*"
			}
		} else if refType.Kind() == reflect.Slice || refType.Kind() == reflect.Array {
			prefix += "[]"
			refType = refType.Elem()
		} else if refType.Kind() == reflect.Chan {
			prefix += refType.ChanDir().String() + " "
			refType = refType.Elem()
		} else {
			break
		}
	}

	return prefix, refType
}

func (t *typeReferenceValue) GetImports() []Import {
	return []Import{t.Import}
}

func (t *typeReferenceValue) GetName() string {
	var buff bytes.Buffer
	buff.WriteString(t.prefix)
	if t.Import != nil {
		buff.WriteString(t.Import.getQualifier())
	}
	buff.WriteString(t.Name)
	return buff.String()
}

type typeReferenceFunc struct {
	Import *ImportSpec
	Name   string
}

func newTypeReferenceFromFunction(t interface{}, alias string) TypeReference {
	// split up the function's name from its package path
	n := runtime.FuncForPC(reflect.ValueOf(t).Pointer()).Name()
	ndxOfLastDot := strings.LastIndex(n, ".")

	return &typeReferenceFunc{
		Import: &ImportSpec{
			Qualified: true,
			Package:   n[:ndxOfLastDot],
			Alias:     alias,
		},
		Name: n[ndxOfLastDot+1:],
	}
}

func (t *typeReferenceFunc) GetImports() []Import {
	return []Import{t.Import}
}

func (t *typeReferenceFunc) GetName() string {
	var buff bytes.Buffer
	if t.Import != nil {
		buff.WriteString(t.Import.getQualifier())
	}
	buff.WriteString(t.Name)
	return buff.String()
}

type unsafeTypeReferenceValue struct {
	Name string
}

func (t *unsafeTypeReferenceValue) GetImports() []Import {
	return nil
}

func (t *unsafeTypeReferenceValue) GetName() string {
	return t.Name
}
