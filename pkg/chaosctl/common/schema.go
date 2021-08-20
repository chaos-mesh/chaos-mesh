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
	MutationType     TypeName
	QueryType        TypeName
	SubscriptionType TypeName
	Types            []*Type
}

type TypeName struct {
	Name graphql.String
}

type Type struct {
	Kind        graphql.String
	Name        graphql.String
	Description graphql.String
	EnumValues  []*EnumValue
	Fields      []*Field
}

type Field struct {
	Name        graphql.String
	Description graphql.String
	Type        FieldType
}

type EnumValue struct {
	Name        graphql.String
	Description graphql.String
}

type FieldType struct {
	Kind graphql.String
	Name *graphql.String
	// OfType *FieldType
}
