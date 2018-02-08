package generator

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

func Generator(cfg Config) {
	fmt.Println("Starting Generator")
	defer fmt.Println("Done Generator")

	g := generator.New()

	// data, err := ioutil.ReadFile(cfg.InFile)
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	fmt.Println(string(data))

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	g.CommandLineParameters(g.Request.GetParameter())

	fmt.Println(*g.Request)
}
