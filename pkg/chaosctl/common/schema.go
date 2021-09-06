// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
)

const (
	ObjectKind  = "OBJECT"
	ListKind    = "LIST"
	NonNullKind = "NON_NULL"
	EnumKind    = "ENUM"
	ScalarKind  = "SCALAR"
)

const (
	ScalarString  = "String"
	ScalarInt     = "Int"
	ScalarTime    = "Time"
	ScalarFloat   = "Float"
	ScalarBoolean = "Boolean"
	ScalarMap     = "Map"
)

type ScalarType string

type Schema struct {
	*RawSchema
	TypeMap map[string]*Type
}

type RawSchema struct {
	MutationType     TypeName
	QueryType        TypeName
	SubscriptionType TypeName
	Types            []*RawType
}

type TypeName struct {
	Name graphql.String
}

type Type struct {
	*RawType
	FieldMap map[string]*Field
	EnumMap  map[string]*EnumValue
}

type RawType struct {
	Kind        graphql.String
	Name        graphql.String
	Description graphql.String
	EnumValues  []*EnumValue
	Fields      []*Field
}

type Field struct {
	Name        graphql.String
	Description graphql.String
	Type        TypeRef1
	Args        []*Argument
}

type Argument struct {
	Name        graphql.String
	Description graphql.String
	Type        TypeRef1
}

type EnumValue struct {
	Name        graphql.String
	Description graphql.String
}

type TypeRef1 struct {
	Kind   graphql.String
	Name   *graphql.String
	OfType *TypeRef2
}

type TypeRef2 struct {
	Kind   graphql.String
	Name   *graphql.String
	OfType *TypeRef3
}

type TypeRef3 struct {
	Kind   graphql.String
	Name   *graphql.String
	OfType *TypeRef4
}

type TypeRef4 struct {
	Kind graphql.String
	Name *graphql.String
}

type TypeRef interface {
	GetKind() graphql.String
	GetName() *graphql.String
	GetOfType() TypeRef
}

type TypeDecorator func(reflect.Type) (reflect.Type, error)

func NewSchema(s *RawSchema) *Schema {
	schema := &Schema{
		RawSchema: s,
		TypeMap:   make(map[string]*Type),
	}

	for _, typ := range schema.Types {
		schema.TypeMap[string(typ.Name)] = NewType(typ)
	}

	return schema
}

func NewType(t *RawType) *Type {
	typ := &Type{
		RawType:  t,
		FieldMap: make(map[string]*Field),
		EnumMap:  make(map[string]*EnumValue),
	}

	for _, field := range typ.Fields {
		typ.FieldMap[string(field.Name)] = field
	}

	for _, enum := range typ.EnumValues {
		typ.EnumMap[strings.ToLower(string(enum.Name))] = enum
	}

	return typ
}

func (ref *TypeRef1) GetKind() graphql.String {
	return ref.Kind
}

func (ref *TypeRef1) GetName() *graphql.String {
	return ref.Name
}

func (ref *TypeRef1) GetOfType() TypeRef {
	return ref.OfType
}

func (ref *TypeRef2) GetKind() graphql.String {
	return ref.Kind
}

func (ref *TypeRef2) GetName() *graphql.String {
	return ref.Name
}

func (ref *TypeRef2) GetOfType() TypeRef {
	return ref.OfType
}

func (ref *TypeRef3) GetKind() graphql.String {
	return ref.Kind
}

func (ref *TypeRef3) GetName() *graphql.String {
	return ref.Name
}

func (ref *TypeRef3) GetOfType() TypeRef {
	return ref.OfType
}

func (ref *TypeRef4) GetKind() graphql.String {
	return ref.Kind
}

func (ref *TypeRef4) GetName() *graphql.String {
	return ref.Name
}

func (ref *TypeRef4) GetOfType() TypeRef {
	return nil
}

func (s *Schema) MustGetType(name string) (*Type, error) {
	typ, ok := s.TypeMap[name]
	if !ok {
		return nil, fmt.Errorf("type `%s` not found", name)
	}
	return typ, nil
}

func (t *Type) mustGetField(name string) (*Field, error) {
	field, ok := t.FieldMap[name]
	if !ok {
		fields := make([]string, 0)
		for f := range t.FieldMap {
			if strings.HasPrefix(f, name) {
				fields = append(fields, f)
			}
		}
		return nil, &FieldNotFound{
			Target: name,
			Fields: fields,
		}
	}
	return field, nil
}

func (t *Type) mustGetEnum(name string) (*EnumValue, error) {
	enum, ok := t.EnumMap[name]
	if !ok {
		variants := make([]string, 0)
		for v := range t.EnumMap {
			if strings.HasPrefix(v, name) {
				variants = append(variants, v)
			}
		}
		return nil, &EnumValueNotFound{
			Target:   name,
			Variants: variants,
		}
	}
	return enum, nil
}

func (s *Schema) resolve(ref TypeRef) (*Type, error) {
	if ref == nil {
		return nil, errors.New("type ref cannot be nil")
	}

	if ref.GetKind() == NonNullKind || ref.GetKind() == ListKind {
		return s.resolve(ref.GetOfType())
	}

	if ref.GetName() == nil {
		return nil, errors.New("name of concret type ref cannot be nil")
	}

	return s.MustGetType(string(*ref.GetName()))
}

func (s *Schema) typeDecorator(ref TypeRef) (TypeDecorator, error) {
	if ref == nil {
		return nil, errors.New("type ref cannot be nil")
	}

	if ref.GetKind() == NonNullKind {
		subDecorator, err := s.typeDecorator(ref.GetOfType())
		if err != nil {
			return nil, err
		}

		return func(t reflect.Type) (reflect.Type, error) {
			newType, err := subDecorator(t)
			if err != nil {
				return t, err
			}

			switch newType.Kind() {
			case reflect.Ptr:
				return newType.Elem(), nil
			case reflect.Slice, reflect.Map:
				return newType, nil
			default:
				return nil, fmt.Errorf("non null decorators do not support type %#v", newType)
			}
		}, nil
	}

	if ref.GetKind() == ListKind {
		subDecorator, err := s.typeDecorator(ref.GetOfType())
		if err != nil {
			return nil, err
		}

		return func(t reflect.Type) (reflect.Type, error) {
			newType, err := subDecorator(t)
			if err != nil {
				return t, err
			}
			return reflect.SliceOf(newType), nil
		}, nil
	}

	return func(t reflect.Type) (reflect.Type, error) {
		return t, nil
	}, nil
}

func (t ScalarType) reflect() (reflect.Type, error) {
	switch t {
	case ScalarString:
		return reflect.PtrTo(reflect.TypeOf(graphql.String(""))), nil
	case ScalarInt:
		return reflect.PtrTo(reflect.TypeOf(graphql.Int(0))), nil
	case ScalarFloat:
		return reflect.PtrTo(reflect.TypeOf(graphql.Float(0))), nil
	case ScalarBoolean:
		return reflect.PtrTo(reflect.TypeOf(graphql.Boolean(false))), nil
	case ScalarTime:
		return reflect.PtrTo(reflect.TypeOf(time.Time{})), nil
	case ScalarMap:
		return reflect.TypeOf(map[string]interface{}{}), nil
	default:
		return nil, fmt.Errorf("unsupported scalar type: %s", t)
	}
}
