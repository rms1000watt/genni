package generator

import (
	"fmt"
)

func Generator(cfg Config) {
	fmt.Println("Starting Generator")
	defer fmt.Println("Done Generator")

	structs, err := Parse(cfg)
	if err != nil {

	}

	if err := Write(structs, cfg); err != nil {

	}
}
