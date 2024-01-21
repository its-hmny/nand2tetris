package main

import (
	"fmt"
	"os"

	"its-hmny.dev/nand2tetris/pkg/asm"
)

var program = []asm.Statement{
	asm.LabelDecl{Name: "LOOP"},
	asm.AInstruction{Location: "R0"},
	asm.CInstruction{Comp: "A+1", Dest: "D"},
	asm.CInstruction{Comp: "D", Jump: "JNE"},
}

func main() {
	// Now, instantiates a code generator for the Asm language
	codegen := asm.NewCodeGenerator(program)
	// Iterates over each program instruction and spits out the textual code
	compiled, err := codegen.Generate()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'codegen' pass:\n\t %s", err)
		os.Exit(-1)
	}

	for _, comp := range compiled {
		fmt.Printf("%s\n", comp)
	}
}
