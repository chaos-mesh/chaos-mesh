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
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/common"
)

func generateConvert(version string, hub string) error {
	fileSet := token.NewFileSet()

	apiDirectory := "api" + "/" + version
	sources, err := ioutil.ReadDir(apiDirectory)
	if err != nil {
		return err
	}

	structMap := map[string]*ast.GenDecl{}
	mainConvertTypes := list.New()
	for _, file := range sources {
		// skip files which are not golang source code
		if !strings.HasSuffix(file.Name(), "go") {
			continue
		}

		filePath := apiDirectory + "/" + file.Name()
		fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		ast.Inspect(fileAst, func(n ast.Node) bool {
			node, ok := n.(*ast.GenDecl)
			if !ok || node.Tok != token.TYPE {
				return true
			}

			typeName := node.Specs[0].(*ast.TypeSpec).Name.Name
			structMap[typeName] = node

			return false
		})

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
					mainConvertTypes.PushBack(typeName)
				}
			}
		}
	}

	convertFilePath := apiDirectory + "/" + "zz_generated.convert.chaosmesh.go"
	convertFile, err := os.Create(convertFilePath)
	if err != nil {
		return err
	}
	defer convertFile.Close()

	impl := convertImpl{
		structMap:    structMap,
		convertTypes: mainConvertTypes,
	}

	return impl.generate(convertFile)
}

type convertImpl struct {
	structMap    map[string]*ast.GenDecl
	convertTypes *list.List

	generatedTypes map[string]struct{}
}

func (c *convertImpl) generate(convertFile *os.File) error {
	c.generatedTypes = make(map[string]struct{})
	hubTypes := make(map[string]struct{})
	for ele := c.convertTypes.Front(); ele != nil; ele = ele.Next() {
		hubTypes[ele.Value.(string)] = struct{}{}
		c.generatedTypes[ele.Value.(string)] = struct{}{}
	}

	convertFile.WriteString(common.Boilerplate + "\n")
	convertFile.WriteString("package " + version + "\n\n")
	convertFile.WriteString(`
import (
	"github.com/chaos-mesh/chaos-mesh/api/` + hub + `"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)
	`)

	for ele := c.convertTypes.Front(); ele != nil; ele = ele.Next() {
		typ := ele.Value.(string)

		from, to := c.generateTypeConvert(typ)

		if _, ok := hubTypes[typ]; ok {
			convertFile.WriteString(`
			func (in *` + typ + `) ConvertTo(dstRaw conversion.Hub) error {
				dst := dstRaw.(*` + hub + `.` + typ + `)
			
			` + to + `
				return nil
			}
			
			func (in *` + typ + `) ConvertFrom(srcRaw conversion.Hub) error {
				src := srcRaw.(*` + hub + `.` + typ + `)
			
			` + from + `
				return nil
			}`)
		} else {
			convertFile.WriteString(`
			func (in *` + typ + `) ConvertTo(dst *` + hub + `.` + typ + `) error {
			` + to + `
				return nil
			}
			
			func (in *` + typ + `) ConvertFrom(src *` + hub + `.` + typ + `) error {
			` + from + `
				return nil
			}`)
		}

	}
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

func (c *convertImpl) printExprToNewVersion(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if _, ok := builtInTypes[t.Name]; ok {
			return t.Name
		}
		return hub + "." + t.Name
	case *ast.SelectorExpr:
		return types.ExprString(expr)
	case *ast.StarExpr:
		return "*" + c.printExprToNewVersion(t.X)
	case *ast.ArrayType:
		return "[]" + c.printExprToNewVersion(t.Elt)
	case *ast.MapType:
		return "map[" + c.printExprToNewVersion(t.Key) + "]" + c.printExprToNewVersion(t.Value)
	default:
		log.Fatal("not supported yet")
	}
	return ""
}

func (c *convertImpl) generateTypeConvert(typ string) (string, string) {
	from := ""
	to := ""

	typDeclare, ok := c.structMap[typ]
	if !ok {
		log.Fatal("type not found: ", typ)
	}

	subType := typDeclare.Specs[0]
	subTypeSpec := subType.(*ast.TypeSpec)
	subTypeSpecDef := interface{}(nil)

	derefSubTypeSpecDef, ok := subTypeSpec.Type.(*ast.StarExpr)
	if ok {
		subTypeSpecDef = derefSubTypeSpecDef.X
	} else {
		subTypeSpecDef = subTypeSpec.Type
	}

	switch subTypeSpecDef := subTypeSpecDef.(type) {
	case *ast.StructType:
		for _, field := range subTypeSpecDef.Fields.List {
			fieldName := c.getFieldName(field)
			if len(fieldName) == 0 {
				log.Fatal("field name is empty")
			}

			if c.canDirectConvert(field.Type) {
				to += "dst." + fieldName + " = in." + fieldName + "\n"
				from += "in." + fieldName + " = src." + fieldName + "\n"
			} else {
				switch fieldTyp := field.Type.(type) {
				case *ast.ArrayType:
					// TODO: support the pointer of slice as the field type

					arrayFrom, arrayTo := c.generateArrayConvert("dst."+fieldName, "src."+fieldName, "in."+fieldName, fieldTyp, false)
					to += `
					dst.` + fieldName + ` = make(` + c.printExprToNewVersion(field.Type) + `, len(in.` + fieldName + `))
					` + arrayTo + `
					`
					from += `
					in.` + fieldName + ` = make(` + types.ExprString(fieldTyp) + `, len(src.` + fieldName + `))
					` + arrayFrom + `
					`
				case *ast.MapType:
					// TODO: support the pointer of map as the field type

					mapFrom, mapTo := c.generateMapConvert("dst."+fieldName, "src."+fieldName, "in."+fieldName, fieldTyp)
					to += `
					dst.` + fieldName + ` = make(` + c.printExprToNewVersion(field.Type) + `)
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
					to += `
					if in.` + fieldName + ` == nil {
						dst.` + fieldName + ` = nil
					} else {
						if dst.` + fieldName + ` == nil {
							dst.` + fieldName + ` = new(` + c.printExprToNewVersion(fieldTyp.X) + `)
						}
						in.` + fieldName + `.ConvertTo(dst.` + fieldName + `)
					}
					`

					from += `
					if in.` + fieldName + ` == nil {
						in.` + fieldName + ` = new(` + types.ExprString(fieldTyp.X) + `) 
					}
					in.` + fieldName + `.ConvertFrom(src.` + fieldName + `)
					`
				}

			}
		}
	case *ast.Ident:
		to += "*dst = " + hub + "." + typ + "(*in)\n"
		from += "*in = " + typ + "(*src)\n"
	case *ast.ArrayType:
		to += "*dst = make(" + hub + "." + typ + ",len(*in))\n"
		from += "*in = make(" + typ + ",len(*src))\n"

		arrayFrom, arrayTo := c.generateArrayConvert("dst", "src", "in", subTypeSpecDef, true)
		to += arrayTo
		from += arrayFrom
	case *ast.MapType:
		to += "*dst = make(" + typ + ")\n"
		from += "*in = make(" + typ + ")\n"

		mapFrom, mapTo := c.generateMapConvert("dst", "src", "in", subTypeSpecDef)
		to += mapTo
		from += mapFrom
	default:
		log.Fatal("unknown type ", typ, reflect.TypeOf(subTypeSpecDef))
	}

	return from, to
}

func (c *convertImpl) needRef(typ ast.Expr) bool {
	switch typ.(type) {
	case *ast.ArrayType, *ast.StarExpr:
		return false
	default:
		return true
	}
}

func (c *convertImpl) generateMapConvert(dst, src, in string, typ *ast.MapType) (string, string) {
	from := ""
	to := ""

	keyIdent, ok := typ.Key.(*ast.Ident)
	if !ok {
		log.Fatal("unsupported map key type", reflect.TypeOf(typ.Key))
	}

	if keyIdent.Name != "string" {
		log.Fatal("unsupported map key", keyIdent.Name)
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
			from += `
			for key,val := range ` + src + ` {
				// TODO: convert the value
			}
			`

			to += `
			for key,val := range ` + in + ` {
				// TODO: convert the value
			}
			`

			// we need to initialize the target value and we will have to handle
			// different situations for different type of value
			log.Fatal("not supported yet")
		case *ast.SliceExpr:
			log.Fatal("not supported yet")
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
				tmpValue := new(` + hub + "." + valueTyp.Name + `)
				val.ConvertTo(tmpValue)
				` + dst + `[key] = *tmpValue
			}
			`
		}
	}

	return from, to
}

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

func (c *convertImpl) getFieldName(field *ast.Field) string {
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
			// only thrid party type is supported
			// for example, the time.Time, or kubernetes metadata
			fieldName = fieldType.Sel.Name
		default:
			log.Fatal("unknown embedded struct type", reflect.TypeOf(fieldType))
		}
	}

	return fieldName
}

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
			if _, ok := c.generatedTypes[fieldType.Name]; !ok {
				c.convertTypes.PushBack(fieldType.Name)
				c.generatedTypes[fieldType.Name] = struct{}{}
			}
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
