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
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type Query struct {
	Name      string
	Argument  string
	ArgValue  interface{} // inner immutable
	Type      *Type
	Decorator TypeDecorator // inner immutable
	Fields    map[string]*Query
}

type Component string

type Variables []*Variable

type Variable struct {
	Name  string
	Value interface{}
}

var (
	enumMap = map[string]reflect.Type{
		"Component": reflect.TypeOf(Component("")),
	}
)

func NewQuery(name string, typ *Type, decorator TypeDecorator) *Query {
	return &Query{
		Name:      name,
		Type:      typ,
		Decorator: decorator,
		Fields:    make(map[string]*Query),
	}
}

func (q *Query) SetArgument(argument, value string, typ *Type) (err error) {
	switch typ.Kind {
	case EnumKind:
		err = q.setEnumValue(value, typ)
	case ScalarKind:
		err = q.setScalarValue(value, typ)
	default:
		err = fmt.Errorf("type %#v is not supported as arguments", typ)
	}

	if err != nil {
		return
	}

	q.Argument = argument
	return nil
}

func (q *Query) setEnumValue(value string, typ *Type) error {
	reflectType, ok := enumMap[string(typ.Name)]
	if !ok {
		return fmt.Errorf("unsupported enum type: %s", typ.Name)
	}

	variant, ok := typ.EnumMap[value]
	if !ok {
		return fmt.Errorf("enum variant not found: %s", value)
	}

	enumValue := reflect.New(reflectType)
	enumValue.Elem().SetString(string(variant.Name))
	q.ArgValue = enumValue.Elem().Interface()
	return nil
}

func (q *Query) setScalarValue(value string, typ *Type) error {
	switch typ.Name {
	case ScalarString:
		q.ArgValue = graphql.String(value)
	case ScalarBoolean:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		q.ArgValue = graphql.Boolean(v)
	case ScalarInt:
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		q.ArgValue = graphql.Int(v)
	case ScalarFloat:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		q.ArgValue = graphql.Float(v)
	default:
		return fmt.Errorf("unsupported argument type: %#v", typ)
	}
	return nil
}

func (q Query) Clone() *Query {
	newFields := make(map[string]*Query)
	for name, field := range q.Fields {
		newFields[name] = field.Clone()
	}
	q.Fields = newFields
	return &q
}

func (q *Query) Merge(other *Query) (*Query, error) {
	if q == nil {
		return other.Merge(q)
	}

	if other == nil {
		return q.Clone(), nil
	}

	if q.Name != other.Name || q.Argument != other.Argument || !reflect.DeepEqual(q.ArgValue, other.ArgValue) {
		return nil, fmt.Errorf("query %#v cannot merge %#v", q, other)
	}

	newQuery := q.Clone()
	for name, field := range other.Fields {
		newField, err := newQuery.Fields[name].Merge(field)
		if err != nil {
			return nil, err
		}
		newQuery.Fields[name] = newField
	}
	return newQuery, nil
}

func (q *Query) String() string {
	segment := q.Name
	if q.ArgValue != nil {
		switch v := q.ArgValue.(type) {
		case graphql.String:
			segment = strings.Join([]string{segment, string(v)}, "/")
		case graphql.Boolean:
			segment = strings.Join([]string{segment, strconv.FormatBool(bool(v))}, "/")
		case graphql.Int:
			segment = strings.Join([]string{segment, fmt.Sprintf("%d", v)}, "/")
		case graphql.Float:
			segment = strings.Join([]string{segment, fmt.Sprintf("%f", v)}, "/")
		default:
			value := reflect.ValueOf(q.ArgValue)
			if value.Type().Kind() == reflect.String {
				arg := value.Convert(reflect.TypeOf("")).Interface().(string)
				segment = strings.Join([]string{segment, strings.ToLower(arg)}, "/")
			}
		}
	}

	if len(q.Fields) == 0 {
		return segment
	}

	fields := make([]string, 0, len(q.Fields))
	for _, f := range q.Fields {
		fields = append(fields, f.String())
	}

	fieldStr := strings.Join(fields, ",")
	return strings.Join([]string{segment, fieldStr}, "/")
}

func (s *Schema) ParseQuery(query []string, super *Type) (*Query, error) {
	if len(query) == 0 {
		return nil, errors.New("query cannot be empty")
	}

	if len(query) == 1 {
		return nil, fmt.Errorf("query %s has single segment", query)
	}

	subQuery := query[1:]

	field, err := super.mustGetField(query[0])
	if err != nil {
		return nil, err
	}

	typ, err := s.resolve(&field.Type)
	if err != nil {
		return nil, err
	}

	decorator, err := s.typeDecorator(&field.Type)
	if err != nil {
		return nil, err
	}

	newQuery := NewQuery(query[0], typ, decorator)
	if len(field.Args) != 0 {
		argument := field.Args[0]
		argType, err := s.resolve(&argument.Type)
		if err != nil {
			return nil, err
		}
		err = newQuery.SetArgument(string(argument.Name), query[1], argType)
		if err != nil {
			return nil, err
		}
		subQuery = query[2:]
	}

	if len(subQuery) == 1 {
		fields, err := s.parseLeaves(subQuery[0], typ)
		if err != nil {
			return nil, err
		}

		for _, field := range fields {
			newQuery.Fields[field.Name] = field
		}
	}

	if len(subQuery) > 1 {
		field, err := s.ParseQuery(subQuery, typ)
		if err != nil {
			return nil, err
		}
		newQuery.Fields[field.Name] = field
	}

	return newQuery, nil
}

func (s *Schema) parseLeaves(leaves string, super *Type) ([]*Query, error) {
	fields := strings.Split(leaves, ",")
	queries := make([]*Query, 0, len(fields))
	for _, f := range fields {
		field, err := super.mustGetField(f)
		if err != nil {
			return nil, err
		}

		if len(field.Args) != 0 {
			// TODO: support default args
			return nil, fmt.Errorf("leaf %s has argument", f)
		}

		typ, err := s.resolve(&field.Type)
		if err != nil {
			return nil, err
		}

		if typ.Kind != ScalarKind {
			// TODO: support object kind
			return nil, fmt.Errorf("type %s is not a scalar kind", typ.Name)
		}

		decorator, err := s.typeDecorator(&field.Type)
		if err != nil {
			return nil, err
		}

		queries = append(queries, NewQuery(f, typ, decorator))
	}
	return queries, nil
}
