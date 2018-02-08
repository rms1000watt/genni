package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func Generator(cfg Config) {
	fmt.Println("Starting Generator")
	defer fmt.Println("Done Generator")

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, "/Users/rsmith/go/src/github.com/rms1000watt/genni/examples/people.go", nil, parser.AllErrors)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ast.Inspect(astFile, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			fmt.Printf("STRUCT Name: %s\n", x.Name)
		case *ast.Field:
			tagName := getTagName(x.Tag)
			fmt.Printf("FIELD Name: %s Type: %s Tag: %s\n", x.Names, x.Type, tagName)
		}
		return true
	})

}

func getTagName(l *ast.BasicLit) (out string) {
	if l != nil {
		out = l.Value
	}
	return
}
