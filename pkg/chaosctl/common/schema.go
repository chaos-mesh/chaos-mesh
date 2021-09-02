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

import "github.com/hasura/go-graphql-client"

const (
	ObjectKind  = "OBJECT"
	ListKind    = "LIST"
	NonNullKind = "NON_NULL"
	EnumKind    = "ENUM"
)

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
		typ.EnumMap[string(enum.Name)] = enum
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
