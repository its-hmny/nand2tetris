package main

import (
	"bytes"
	"fmt"
	"os"

	"its-hmny.dev/nand2tetris/assembler"
	"its-hmny.dev/nand2tetris/hack"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("USAGE: assembler [INPUT] [OUTPUT]")
		os.Exit(-1)
	}

	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR: Unable to open input file: %s\n", err)
		os.Exit(-1)
	}

	output, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Printf("ERROR: Unable to open output file: %s\n", err)
		os.Exit(-1)
	}
	defer output.Close()

	// Instantiate a parser for the Assembler program
	parser := assembler.NewParser()
	// Parses the input file content and extract an AST from it
	ast, success := parser.Parse(bytes.NewReader(input))
	if !success {
		fmt.Print("ERROR: Unable to complete 'parsing' pass\n")
		os.Exit(-1)
	}

	// Instantiate a parser for the Assembler to Hack lowerer
	lowerer := assembler.NewHackLowerer()
	// Lowers the AST to an in-memory/IR format that follows the Hack specs.
	program, table, err := lowerer.FromAST(ast)
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'lowering' pass: %s\n", err)
		os.Exit(-1)
	}

	// Now, instantiates a code generator for the Hack (compiled) program
	translator := hack.NewCodeGenerator(program, table)
	// Iterates over each program instruction and spits out the relative translation
	compiled, err := translator.Translate()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'codegen' pass:\n\t %s", err)
		os.Exit(-1)
	}

	for _, comp := range compiled {
		line := fmt.Sprintf("%s\n", comp)
		output.Write([]byte(line))
	}
}
