package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var typeMap = map[string]bool{}

func Parse(cfg Config) (structs []Struct, err error) {
	log.Debug("Starting Parse")
	defer log.Debug("Done Parse")

	populateTypeMap()

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
			structs = append(structs, Struct{})
			// structs[structInd].Name = NewName(x.Name.String())
		case *ast.Field:
			fieldName := NewName(getFieldName(x.Names))
			fieldType := getFieldType(x.Type, 0)

			structs[structInd].Fields = append(structs[structInd].Fields, Field{
				// Name:       fieldName,
				Tag:        getTag(x.Tag, fieldName.LowerSnake),
				DataTypeIn: fieldType,
				IsRepeated: getIsRepeated(x.Type),
				IsMap:      strings.Contains(fieldType, "map["),
			})
		}
		return true
	})

	log.Debug(structs)
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

func getIsRepeated(in ast.Expr) (out bool) {
	// i := &types.Info{}
	// fmt.Println(i.TypeOf(in))
	fmt.Println(types.ExprString(in))
	// fmt.Println(types.Int)
	fmt.Println(isBuiltin(in))

	return
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
		jsonTag = fmt.Sprintf(`json:"%s"`, snake)
	}

	var dbTag string
	if !strings.Contains(out, `db:"`) {
		dbTag = fmt.Sprintf(`db:"%s"`, snake)
	}

	return strings.TrimSpace(fmt.Sprintf("%s %s %s", jsonTag, dbTag, out))
}

func getFieldType(in ast.Expr, recursionCnt int) (out string) {
	// log.Debug(reflect.TypeOf(in))
	switch y := in.(type) {
	case *ast.Ident:
		if recursionCnt == 0 {
			return fmt.Sprintf("*%s", y.Name)
		}
		return y.Name
	case *ast.MapType:
		recursionCnt++
		k := getFieldType(y.Key, recursionCnt)
		v := getFieldType(y.Value, recursionCnt)
		return fmt.Sprintf("*map[%s]%s", k, v)
	case *ast.ArrayType:
		recursionCnt++
		t := getFieldType(y.Elt, recursionCnt)
		return fmt.Sprintf("[]*%s", t)
	case *ast.StructType:
		log.Error("Anonymous struct field not supported. Exiting..")
		os.Exit(1)
	case *ast.InterfaceType:
		log.Error("Interface fields not supported. Exiting..")
		os.Exit(1)
	case *ast.StarExpr:
		log.Error("Pointer fields not supported. Exiting..")
		os.Exit(1)
	default:
		log.Errorf("Unhandled Type Encountered %s. Exiting..", y)
		os.Exit(1)
	}
	return
}
