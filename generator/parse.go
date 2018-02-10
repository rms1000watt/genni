package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

func Parse(cfg Config) (structs []Struct, err error) {
	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, cfg.InFile, nil, parser.AllErrors)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	currentStructName := ""
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			fmt.Printf("STRUCT Name: %s\n", x.Name)
			currentStructName = x.Name.String()

		case *ast.Field:
			tag := getTag(x.Tag)
			fieldName := getFieldName(x.Names)

			fmt.Println(reflect.TypeOf(x.Type))
			switch y := x.Type.(type) {
			case *ast.MapType:
				fmt.Println("MapType", y.Key, y.Value)
			case *ast.ArrayType:
				fmt.Println("ArrayType")
			default:
				fmt.Println(y)
			}

			fmt.Printf("FIELD Name: %s Type: %s Tag: %s\n", fieldName, x.Type, tag)
		}
		return true
	})
	return
}

func getFieldName(in []*ast.Ident) (out string) {
	if len(in) == 1 {
		out = in[0].Name
	}
	return
}

func getTag(in *ast.BasicLit) (out string) {
	if in != nil {
		out = in.Value
	}
	return
}
