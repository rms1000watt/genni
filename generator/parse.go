package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

var typeMap = map[string]bool{}

func Parse(cfg Config) (genni Genni, err error) {
	log.Debug("Starting Parse")
	defer log.Debug("Done Parse")

	populateTypeMap()

	// Parse structs
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, cfg.InFile, nil, parser.AllErrors)
	if err != nil {
		log.Debug("Failed parsing file:", err)
		return
	}

	structInd := -1
	genni = NewGenni()
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			log.Debugf("Struct Name: %s", x.Name)
			structInd++
			genni.Structs = append(genni.Structs, Struct{})
			genni.Structs[structInd].Name = NewName(x.Name.String() + "P")
		case *ast.Field:
			fieldName := NewName(getFieldName(x.Names))

			fmt.Println("type:", types.ExprString(x.Type))
			fmt.Println("ast type:", reflect.TypeOf(x.Type))
			fmt.Println("repeated builtin:", getIsRepeatedBuiltin(x.Type))
			fmt.Println("repeated struct:", getIsRepeatedStruct(x.Type))
			fmt.Println("struct:", getIsStruct(x.Type))
			fmt.Println("dataType:", getDataType(x.Type))
			fmt.Println("map:", getIsMap(x.Type))
			fmt.Println()

			genni.Structs[structInd].Fields = append(genni.Structs[structInd].Fields, Field{
				Name:              fieldName,
				Tag:               getTag(x.Tag, fieldName.LowerSnake),
				DataType:          getDataType(x.Type),
				DataTypeIn:        getDataTypeIn(x.Type, 0, false),
				IsRepeatedBuiltin: getIsRepeatedBuiltin(x.Type),
				IsRepeatedStruct:  getIsRepeatedStruct(x.Type),
				IsMap:             getIsMap(x.Type),
				IsStruct:          getIsStruct(x.Type),
			})
		}
		return true
	})

	// Parse package
	fset = token.NewFileSet()
	astFile, err = parser.ParseFile(fset, cfg.InFile, nil, parser.PackageClauseOnly)
	if err != nil {
		log.Debug("Failed parsing file:", err)
		return
	}

	ast.Inspect(astFile, func(n ast.Node) bool {
		if pkg, ok := n.(*ast.Ident); ok {
			genni.Package = pkg.Name
		}
		return true
	})

	return
}

func populateTypeMap() {
	for _, typ := range types.Typ {
		typeMap[typ.Name()] = true
	}
}

func isBuiltin(in ast.Expr) (out bool) {
	return typeMap[types.ExprString(in)]
}

func getIsMap(in ast.Expr) (out bool) {
	if _, ok := in.(*ast.MapType); ok {
		return true
	}
	return
}

func getIsStruct(in ast.Expr) (out bool) {
	return !isBuiltin(in)
}

func getIsRepeatedBuiltin(in ast.Expr) (out bool) {
	arrayType, ok := in.(*ast.ArrayType)
	if !ok {
		return
	}

	return isBuiltin(arrayType.Elt)
}

func getIsRepeatedStruct(in ast.Expr) (out bool) {
	arrayType, ok := in.(*ast.ArrayType)
	if !ok {
		return
	}

	return !isBuiltin(arrayType.Elt)
}

func getFieldName(in []*ast.Ident) (out string) {
	if len(in) == 1 {
		out = in[0].Name
	}
	return
}

func getTag(in *ast.BasicLit, snake string) (out string) {
	if in != nil {
		out = in.Value
	}

	var jsonTag string
	if !strings.Contains(out, `json:"`) {
		jsonTag = fmt.Sprintf(`json:"%s,omitempty"`, snake)
	}

	var dbTag string
	if !strings.Contains(out, `db:"`) {
		dbTag = fmt.Sprintf(`db:"%s"`, snake)
	}

	return strings.Trim(strings.TrimSpace(fmt.Sprintf("%s %s %s", jsonTag, dbTag, out)), "`")
}

func getDataType(in ast.Expr) (out string) {
	isValidType(in)

	switch y := in.(type) {
	case *ast.Ident:
		return y.Name
	case *ast.MapType:
		k := getDataType(y.Key)
		v := getDataType(y.Value)
		return fmt.Sprintf("map[%s]%s", k, v)
	case *ast.ArrayType:
		t := getDataType(y.Elt)
		return fmt.Sprintf("[]%s", t)
	default:
		log.Errorf("Unhandled Type Encountered %s. Exiting..", y)
		os.Exit(1)
	}
	return
}

func getDataTypeIn(in ast.Expr, recursionCnt int, isMapType bool) (out string) {
	isValidType(in)

	switch y := in.(type) {
	case *ast.Ident:
		out = y.Name
		if recursionCnt == 0 {
			out = "*" + out
		}
		if !isMapType && !isBuiltin(in) {
			out += "P"
		}
		return
	case *ast.MapType:
		recursionCnt++
		k := getDataTypeIn(y.Key, recursionCnt, true)
		v := getDataTypeIn(y.Value, recursionCnt, true)
		return fmt.Sprintf("*map[%s]%s", k, v)
	case *ast.ArrayType:
		recursionCnt++
		t := getDataTypeIn(y.Elt, recursionCnt, false)
		return fmt.Sprintf("[]*%s", t)
	default:
		log.Errorf("Unhandled Type Encountered %s. Exiting..", y)
		os.Exit(1)
	}
	return
}

func isValidType(in ast.Expr) {
	switch in.(type) {
	case *ast.StructType:
		log.Error("Anonymous struct field not supported. Exiting..")
		os.Exit(1)
	case *ast.InterfaceType:
		log.Error("Interface fields not supported. Exiting..")
		os.Exit(1)
	case *ast.StarExpr:
		log.Error("Pointer fields not supported. Exiting..")
		os.Exit(1)
	}
	return
}
