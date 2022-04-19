// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package autoconvert

import (
	"container/list"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/common"
)

var errTypeNotSupported = errors.New("type is not supported")
var errTypeNotFound = errors.New("type not found")

func generateConvert(log logr.Logger, version string, hub string) error {
	fileSet := token.NewFileSet()

	apiDirectory := "api" + "/" + version
	sources, err := ioutil.ReadDir(apiDirectory)
	if err != nil {
		return errors.Wrapf(err, "read directory", apiDirectory)
	}

	definitionMap := map[string]*ast.GenDecl{}
	types := newUniqueList()
	for _, file := range sources {
		// skip files which are not golang source code
		if !strings.HasSuffix(file.Name(), "go") {
			continue
		}

		filePath := apiDirectory + "/" + file.Name()
		fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
		if err != nil {
			return errors.Wrapf(err, "parse file %s", filePath)
		}

		// add all type definition to the definitionMap
		ast.Inspect(fileAst, func(n ast.Node) bool {
			node, ok := n.(*ast.GenDecl)
			if !ok || node.Tok != token.TYPE {
				return true
			}

			typeName := node.Specs[0].(*ast.TypeSpec).Name.Name
			definitionMap[typeName] = node

			return false
		})

		// read the comment map to decide which types need to be converted
		cmap := ast.NewCommentMap(fileSet, fileAst, fileAst.Comments)
		for node, commentGroups := range cmap {
			node, ok := node.(*ast.GenDecl)
			if !ok || node.Tok != token.TYPE {
				continue
			}

			typeName := node.Specs[0].(*ast.TypeSpec).Name.Name
			for _, commentGroup := range commentGroups {
				isObjectRoot := false

				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+kubebuilder:object:root=true") {
						isObjectRoot = true
					}
				}

				if isObjectRoot {
					types.push(typeName)
				}
			}
		}
	}

	convertFilePath := apiDirectory + "/" + "zz_generated.convert.chaosmesh.go"
	convertFile, err := os.Create(convertFilePath)
	if err != nil {
		return errors.Wrapf(err, "create file %s", convertFilePath)
	}
	defer convertFile.Close()

	impl := convertImpl{
		definitionMap: definitionMap,
		version:       version,
		hub:           hub,
		types:         types,
		log:           log,
	}

	return impl.generate(convertFile)
}

type convertImpl struct {
	definitionMap map[string]*ast.GenDecl

	version, hub string

	types *uniqueList

	log logr.Logger
}

func (c *convertImpl) generate(convertFile *os.File) error {
	// mark the original added types.
	// the signature of these types will be different from other `Convert` function.
	// because the `controller-runtime` requires them to convert to the `conversion.Hub`
	hubTypes := make(map[string]struct{})

	err := c.types.forEach(func(typ string) error {
		hubTypes[typ] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}

	convertFile.WriteString(common.Boilerplate + "\n")
	convertFile.WriteString("package " + c.version + "\n\n")
	convertFile.WriteString(`
import (
	"github.com/chaos-mesh/chaos-mesh/api/` + c.hub + `"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)
	`)

	c.types.forEach(func(typ string) error {
		from, to, err := c.generateTypeConvert(typ)
		if err != nil {
			return err
		}

		if _, ok := hubTypes[typ]; ok {
			convertFile.WriteString(`
			func (in *` + typ + `) ConvertTo(dstRaw conversion.Hub) error {
				dst := dstRaw.(*` + c.hub + `.` + typ + `)
			
			` + to + `
				return nil
			}
			
			func (in *` + typ + `) ConvertFrom(srcRaw conversion.Hub) error {
				src := srcRaw.(*` + c.hub + `.` + typ + `)
			
			` + from + `
				return nil
			}`)
		} else {
			convertFile.WriteString(`
			func (in *` + typ + `) ConvertTo(dst *` + c.hub + `.` + typ + `) error {
			` + to + `
				return nil
			}
			
			func (in *` + typ + `) ConvertFrom(src *` + c.hub + `.` + typ + `) error {
			` + from + `
				return nil
			}`)
		}

		return nil
	})
	return nil
}

var builtInTypes = map[string]struct{}{
	"bool":    {},
	"string":  {},
	"byte":    {},
	"int":     {},
	"int8":    {},
	"int16":   {},
	"int32":   {},
	"int64":   {},
	"uint":    {},
	"uint8":   {},
	"uint16":  {},
	"uint32":  {},
	"uint64":  {},
	"float":   {},
	"float32": {},
	"float64": {},
}

func (c *convertImpl) printExprToNewVersion(expr ast.Expr) (string, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		if _, ok := builtInTypes[t.Name]; ok {
			return t.Name, nil
		}
		return c.hub + "." + t.Name, nil
	case *ast.SelectorExpr:
		return types.ExprString(expr), nil
	case *ast.StarExpr:
		newTyp, err := c.printExprToNewVersion(t.X)
		if err != nil {
			return "", err
		}
		return "*" + newTyp, nil
	case *ast.ArrayType:
		newTyp, err := c.printExprToNewVersion(t.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + newTyp, nil
	case *ast.MapType:
		newKeyTyp, err := c.printExprToNewVersion(t.Key)
		if err != nil {
			return "", err
		}
		newValueTyp, err := c.printExprToNewVersion(t.Key)
		if err != nil {
			return "", err
		}
		return "map[" + newKeyTyp + "]" + newValueTyp, nil
	default:
		return "", errors.Wrapf(errTypeNotSupported, "type %T", t)
	}
}

// generateTypeConvert generates the convert function for the type
func (c *convertImpl) generateTypeConvert(typ string) (string, string, error) {
	from := ""
	to := ""

	typeDeclare, ok := c.definitionMap[typ]
	if !ok {
		return "", "", errors.WithStack(errTypeNotFound)
	}

	typeSpec := typeDeclare.Specs[0].(*ast.TypeSpec)
	derefedTypeSpec := interface{}(nil)

	// if the type spec is a pointer, deref it
	//
	// because for any type (other than slice, map), we will generate the convert
	// function whose receiver is the pointer type, so we can deref the pointer here for the convenience
	pointerTypeSpec, ok := typeSpec.Type.(*ast.StarExpr)
	if ok {
		derefedTypeSpec = pointerTypeSpec.X
		switch derefedTypeSpec.(type) {
		case *ast.ArrayType:
			return "", "", errors.Wrapf(errTypeNotSupported, "type %T", derefedTypeSpec)
		case *ast.MapType:
			return "", "", errors.Wrapf(errTypeNotSupported, "type %T", derefedTypeSpec)
		}
	} else {
		derefedTypeSpec = typeSpec.Type
	}

	switch typeSpecDef := derefedTypeSpec.(type) {
	case *ast.StructType:
		// if this type is a definition for struct
		// we will generate the convert function for every field
		for _, field := range typeSpecDef.Fields.List {
			fieldName, err := c.getFieldName(field)
			if err != nil {
				return "", "", err
			}

			if c.canDirectConvert(field.Type) {
				to += "dst." + fieldName + " = in." + fieldName + "\n"
				from += "in." + fieldName + " = src." + fieldName + "\n"
			} else {
				switch fieldTyp := field.Type.(type) {
				case *ast.ArrayType:
					arrayFrom, arrayTo := c.generateArrayConvert("dst."+fieldName, "src."+fieldName, "in."+fieldName, fieldTyp, false)
					newFieldType, err := c.printExprToNewVersion(field.Type)
					if err != nil {
						return "", "", err
					}

					to += `
					dst.` + fieldName + ` = make(` + newFieldType + `, len(in.` + fieldName + `))
					` + arrayTo + `
					`
					from += `
					in.` + fieldName + ` = make(` + types.ExprString(fieldTyp) + `, len(src.` + fieldName + `))
					` + arrayFrom + `
					`
				case *ast.MapType:
					mapFrom, mapTo, err := c.generateMapConvert("dst."+fieldName, "src."+fieldName, "in."+fieldName, fieldTyp, false)
					if err != nil {
						return "", "", err
					}

					newFieldType, err := c.printExprToNewVersion(field.Type)
					if err != nil {
						return "", "", err
					}

					to += `
					dst.` + fieldName + ` = make(` + newFieldType + `)
					` + mapTo + `
					`
					from += `
					in.` + fieldName + ` = make(` + types.ExprString(fieldTyp) + `)
					` + mapFrom + `
					`
				case *ast.Ident:
					to += `
						in.` + fieldName + `.ConvertTo(&dst.` + fieldName + `)
						`
					from += `
						in.` + fieldName + `.ConvertFrom(&src.` + fieldName + `)
						`
				case *ast.StarExpr:
					// TODO: support the pointer of slice and map as the field type
					if _, ok := fieldTyp.X.(*ast.StructType); !ok {
						return "", "", errors.Wrapf(errTypeNotSupported, "type %T", fieldTyp.X)
					}

					typeName := types.ExprString(fieldTyp.X)
					typeNameInNewVersion, err := c.printExprToNewVersion(fieldTyp.X)
					if err != nil {
						return "", "", err
					}

					to += `
					if in.` + fieldName + ` == nil {
						dst.` + fieldName + ` = nil
					} else {
						if dst.` + fieldName + ` == nil {
							dst.` + fieldName + ` = new(` + typeNameInNewVersion + `)
						}
						in.` + fieldName + `.ConvertTo(dst.` + fieldName + `)
					}
					`

					from += `
					if in.` + fieldName + ` == nil {
						in.` + fieldName + ` = new(` + typeName + `) 
					}
					in.` + fieldName + `.ConvertFrom(src.` + fieldName + `)
					`
				}

			}
		}
	case *ast.Ident:
		// this type definition is an type alias, then it can be simply converted.
		to += "*dst = " + c.hub + "." + typ + "(*in)\n"
		from += "*in = " + typ + "(*src)\n"
	case *ast.ArrayType:
		// this type definition is an array
		to += "*dst = make(" + c.hub + "." + typ + ",len(*in))\n"
		from += "*in = make(" + typ + ",len(*src))\n"

		arrayFrom, arrayTo := c.generateArrayConvert("dst", "src", "in", typeSpecDef, true)
		to += arrayTo
		from += arrayFrom
	case *ast.MapType:
		// this type definition is a map
		to += "*dst = make(" + typ + ")\n"
		from += "*in = make(" + typ + ")\n"

		mapFrom, mapTo, err := c.generateMapConvert("dst", "src", "in", typeSpecDef, true)
		if err != nil {
			return "", "", err
		}

		to += mapTo
		from += mapFrom
	default:
		return "", "", errors.Wrapf(errTypeNotSupported, "type %T", typeSpecDef)
	}

	return from, to, nil
}

func (c *convertImpl) needRef(typ ast.Expr) bool {
	switch typ.(type) {
	case *ast.ArrayType, *ast.StarExpr:
		return false
	default:
		return true
	}
}

// generateMapConvert will generate the conversion between map
//
// all of dst, src and in should have been initialized.
func (c *convertImpl) generateMapConvert(dst, src, in string, typ *ast.MapType, ptr bool) (string, string, error) {
	from := ""
	to := ""

	if ptr {
		in = `(*` + in + `)`
		src = `(*` + src + `)`
		dst = `(*` + dst + `)`
	}

	if c.canDirectConvert(typ.Value) {
		from += `
		for key,val := range ` + src + ` {
			` + in + `[key] = val
		}
		`

		to += `
		for key,val := range ` + in + ` {
			` + dst + `[key] = val
		}
		`
	} else {
		switch valueTyp := typ.Value.(type) {
		case *ast.StarExpr:
			// we need to initialize the target value and we will have to handle
			// different situations for different type of value
			return "", "", errors.Wrapf(errTypeNotSupported, "type %T", typ.Value)
		case *ast.SliceExpr:
			return "", "", errors.Wrapf(errTypeNotSupported, "type %T", typ.Value)
		case *ast.Ident:
			// this ident cannot be directly convert, while the value is not a pointer

			// assume this value is a struct
			// TODO: support other type of values, they should be initialize to empty
			from += `
			for key,val := range ` + src + ` {
				tmpValue := new(` + valueTyp.Name + `)
				tmpValue.ConvertFrom(&val)
				` + in + `[key] = *tmpValue
			}
			`

			to += `
			for key,val := range ` + in + ` {
				tmpValue := new(` + c.hub + "." + valueTyp.Name + `)
				val.ConvertTo(tmpValue)
				` + dst + `[key] = *tmpValue
			}
			`
		}
	}

	return from, to, nil
}

// generateArrayConvert will generate the conversion between array
//
// all of dst, src and in should have been initialized.
func (c *convertImpl) generateArrayConvert(dst, src, in string, typ *ast.ArrayType, ptr bool) (string, string) {
	from := ""
	to := ""

	if ptr {
		in = `(*` + in + `)`
		src = `(*` + src + `)`
		dst = `(*` + dst + `)`
	}

	if c.canDirectConvert(typ.Elt) {
		from += `
		for i := range ` + src + ` {
			` + in + `[i] = ` + src + `[i]
		}
		`

		to += `
		for i := range ` + in + ` {
			` + dst + `[i] = ` + in + `[i]
		}
		`
	} else {
		ref := ""
		if c.needRef(typ.Elt) {
			ref = "&"
		}

		from += `
		for i := range ` + src + ` {
			` + in + `[i].ConvertFrom(` + ref + src + `[i])
		}
		`

		to += `
		for i := range ` + in + ` {
			` + in + `[i].ConvertTo(` + ref + dst + `[i])
		}
		`
	}

	return from, to
}

// getFieldName returns the name of a field in the struct
//
// for example, the following struct:
//
// ```go
// type S struct {
// 		A string
// 		common.B
//      C
// }
// ```
//
// getFieldName for these three fields would return
// `A`, `B`, `C`
// Then if you have a variable `s` with type `S`, then you can
// refer to the field with `s.A`, `s.B` and `s.C`
func (c *convertImpl) getFieldName(field *ast.Field) (string, error) {
	fieldName := ""
	if field.Names != nil {
		fieldName = field.Names[0].Name
	} else {
		fieldType := interface{}(nil)
		derefFieldType, ok := field.Type.(*ast.StarExpr)
		if ok {
			fieldType = derefFieldType.X
		} else {
			fieldType = field.Type
		}

		// or it is an embedded struct
		switch fieldType := fieldType.(type) {
		case *ast.Ident:
			fieldName = fieldType.Name
		case *ast.SelectorExpr:
			// only third party type is supported
			// for example, the time.Time, or kubernetes metadata
			fieldName = fieldType.Sel.Name
		default:
			return "", errors.Errorf("unknown embedded struct type: %T", fieldType)
		}
	}

	return fieldName, nil
}

// canDirectConvert returns whether the type can be directly converted.
//
// But as this package only works with AST, without type related information, it
// cannot be 100 percent sure. This function guaranteed that:
//
// 1. If this type is a built in type, it will return true
// 2. If this type is a type from other package, it will return true, as we assume
//    the dependency of different versions of API will be the same.
// 3. If this type is a pointer, it will be dereferenced and see whether the type
//    it references can be directly converted.
// 4. If this type is a slice, it will return true if the element type can be
//    converted directly.
// 5. For other senerio, we assume this type cannot be directly converted, and it will be
//    pushed into the convertTypes. An `ConvertTo` and `ConvertFrom` will be implemented
//    for this type.
func (c *convertImpl) canDirectConvert(typ ast.Expr) bool {
	directConvert := false

	realType := interface{}(nil)

	derefedType, ok := typ.(*ast.StarExpr)
	if ok {
		realType = derefedType.X
	} else {
		realType = typ
	}

	switch fieldType := realType.(type) {
	case *ast.Ident:
		if _, ok := builtInTypes[fieldType.Name]; ok {
			directConvert = true
		} else {
			c.types.push(fieldType.Name)
		}
	case *ast.SelectorExpr:
		// assume the selectorExpr always refers to a third party type
		directConvert = true
	case *ast.ArrayType:
		return c.canDirectConvert(fieldType.Elt)
	default:
		// do nothing for other possibilities
	}

	return directConvert
}

type uniqueList struct {
	convertTypes *list.List
	exist        map[string]struct{}
}

func newUniqueList() *uniqueList {
	return &uniqueList{
		convertTypes: list.New(),
		exist:        make(map[string]struct{}),
	}
}

func (c *uniqueList) push(typs ...string) {
	for _, typ := range typs {
		if _, ok := c.exist[typ]; !ok {
			c.convertTypes.PushBack(typ)
			c.exist[typ] = struct{}{}
		}
	}
}

func (c *uniqueList) forEach(f func(string) error) error {
	for e := c.convertTypes.Front(); e != nil; e = e.Next() {
		err := f(e.Value.(string))
		if err != nil {
			return err
		}
	}

	return nil
}
