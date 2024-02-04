package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/teris-io/cli"
	"its-hmny.dev/nand2tetris/pkg/asm"
	"its-hmny.dev/nand2tetris/pkg/hack"
)

var Description = strings.ReplaceAll(`
The Hack Assembler takes assembly language code written in the Hack assembly language
and translates it into machine code that can be executed by the Hack computer. The process
involves parsing the assembly code, resolving symbols, and generating machine code.
`, "\n", " ")

var HackAssembler = cli.New(Description).
	WithArg(cli.NewArg("input", "The assembler (.asm) file to be compiled")).
	WithArg(cli.NewArg("output", "The compiled binary output (.hack)")).
	WithAction(Handler)

func Handler(args []string, options map[string]string) int {
	input, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Printf("ERROR: Unable to open input file: %s\n", err)
		return -1
	}

	output, err := os.Create(args[1])
	if err != nil {
		fmt.Printf("ERROR: Unable to open output file: %s\n", err)
		return -1
	}
	defer output.Close()

	// Instantiate a parser for the Asm program
	parser := asm.NewParser(bytes.NewReader(input))
	// Parses the input file content and extract an AST (as a 'asm.Program') from it.
	asmProgram, err := parser.Parse()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'parsing' pass: %s\n", err)
		return -1
	}

	// Instantiate a lowerer to convert the program from Asm to Hack
	lowerer := asm.NewLowerer(asmProgram)
	// Lowers the asm.Program to an in-memory/IR representation of its Hack counterpart 'hack.Program'.
	hackProgram, table, err := lowerer.Lower()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'lowering' pass: %s\n", err)
		return -1
	}

	// Now, instantiates a code generator for the Hack (compiled) program
	codegen := hack.NewCodeGenerator(hackProgram, table)
	// Iterates over each instruction and spits out the relative textual representation.
	compiled, err := codegen.Generate()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'codegen' pass:\n\t %s", err)
		return -1
	}

	for _, comp := range compiled {
		line := fmt.Sprintf("%s\n", comp)
		output.Write([]byte(line))
	}

	return 0
}

func main() { os.Exit(HackAssembler.Run(os.Args, os.Stdout)) }
