package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
)

func Parse(cfg Config) (structs []Struct, err error) {
	log.Debug("Starting Parse")
	defer log.Debug("Done Parse")

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, cfg.InFile, nil, parser.AllErrors)
	if err != nil {
		log.Debug("Error:", err)
		return
	}

	structInd := -1
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			log.Debugf("Struct Name: %s", x.Name)
			structInd++
			structs[structInd].Name = NewName(x.Name.String())
		case *ast.Field:
			field := Field{
				Name: NewName(getFieldName(x.Names)),
				// DataTypeName
			}

			fieldTag := getTag(x.Tag)
			fieldName := getFieldName(x.Names)
			fieldType := getFieldType(x.Type)

			// fmt.Printf("FIELD Name: %s Type: %s Tag: %s\n", fieldName, x.Type, tag)
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

func getFieldType(in ast.Expr) (out string) {
	log.Debug(reflect.TypeOf(in))
	switch y := in.(type) {
	case *ast.MapType:
		log.Debug("MapType: ", y.Key, y.Value)
	case *ast.ArrayType:
		log.Debug("ArrayType: ", y.Elt)
	case *ast.StructType:
		log.Debug("StructType: ", x.Names)
	case *ast.Ident:
		log.Debug("Ident: ", y.Name)
	case *ast.InterfaceType:
		log.Errorf("Interfaces %s not supported. Exiting..", x.Names)
		os.Exit(1)
	case *ast.StarExpr:
		log.Errorf("Pointers %s not supported. Exiting..", x.Names)
		os.Exit(1)
	default:
		log.Error("Unhandled Type Encountered: ", y)
		os.Exit(1)
	}
	return
}
