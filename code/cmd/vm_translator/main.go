package main

import (
	"bytes"
	"fmt"
	"os"

	"its-hmny.dev/nand2tetris/pkg/asm"
	"its-hmny.dev/nand2tetris/pkg/vm"
)

var program = []asm.Statement{
	asm.LabelDecl{Name: "LOOP"},
	asm.AInstruction{Location: "R0"},
	asm.CInstruction{Comp: "A+1", Dest: "D"},
	asm.CInstruction{Comp: "D", Jump: "JNE"},
}

var VmProgram = `
	push constant 7
	push constant 8 // Test comment, should work
	add
	// Another comment that should work as well
	push constant 3
	sub
`

func main() {

	// Instantiate a parser for the vm program
	parser := vm.NewParser()
	// Parses the input file content and extract an AST from it
	_, success := parser.Parse(bytes.NewReader([]byte(VmProgram)))
	if !success {
		fmt.Print("ERROR: Unable to complete 'parsing' pass\n")
		os.Exit(-1)
	}

	// // Now, instantiates a code generator for the Asm language
	// codegen := asm.NewCodeGenerator(program)
	// // Iterates over each program instruction and spits out the textual code
	// compiled, err := codegen.Generate()
	// if err != nil {
	// 	fmt.Printf("ERROR: Unable to complete 'codegen' pass:\n\t %s", err)
	// 	os.Exit(-1)
	// }

	// for _, comp := range compiled {
	// 	fmt.Printf("%s\n", comp)
	// }
}
