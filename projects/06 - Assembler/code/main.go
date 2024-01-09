package main

import (
	"fmt"
	"reflect"
	"strings"

	"its-hmny.dev/n2t-assembler/assembler"
	"its-hmny.dev/n2t-assembler/hack"
)

var AsmProgram = `
	@42
	// Test comment
	M=D+1 // Another test comment
	// Test comment 2
	@END
	M+1;JEQ
	
	(END)
		@END
		0; JMP
`

func main() {
	// Instantiate a parser for the Assembler program
	parser := assembler.NewParser()
	// Parses the input file content and extract an AST from it
	ast, success := parser.Parse(strings.NewReader(AsmProgram))
	if !success {
		fmt.Printf("ERROR: Unable to complete 'parsing' pass\n")
	}

	// Instantiate a parser for the Assembler to Hack lowerer
	lowerer := assembler.NewHackLowerer()
	// Lowers the AST to an in-memory/IR format that follows the Hack specs.
	program, table, err := lowerer.FromAST(ast)
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'lowering' pass: %s\n", err)
	}

	// Now, instantiates a code generator for the Hack (compiled) program
	translator := hack.NewCodeGenerator(program, table)
	// Iterates over each program instruction and spits out the relative translation
	compiled, err := translator.Translate()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'codegen' pass:\n\t %s", err)
	}

	// For the time being we simply dump the program on stdout before exiting, each
	// and every instruction is printed both in its in-memory format and raw binary
	for n := range compiled {
		assembler, hack := translator.Program[n], compiled[n]

		fmt.Printf("%s: =>\n", reflect.TypeOf(assembler).Name())
		fmt.Printf(" Asm:  %+v\n", assembler)
		fmt.Printf(" Hack: %s\n\n", hack)
	}
}
