package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/teris-io/cli"
	"its-hmny.dev/nand2tetris/pkg/asm"
	"its-hmny.dev/nand2tetris/pkg/vm"
)

var Description = strings.ReplaceAll(`
The VM Translator translates programs (composed of multiple modules/files) written in 
the VM language into Hack assembly code that can be further elaborated. The VM language
is a higher-level (bytecode'like) language tailored for use with the Hack computer arch.
`, "\n", " ")

var VmTranslator = cli.New(Description).
	// TODO(hmny): 'input' should be registered as optional and put last to support multi-args
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

	// Instantiate a parser for the Vm program
	parser := vm.NewParser(bytes.NewReader(input))
	// Parses the input file content and extract an AST (as a 'vm.Module') from it.
	vmModule, err := parser.Parse()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'parsing' pass: %s\n", err)
		os.Exit(-1)
	}

	// Instantiate a lowerer to convert the program from Vm to Asm
	lowerer := vm.NewLowerer([]vm.Module{vmModule})
	// Lowers the vm.Program to an in-memory/IR representation of its Asm counterpart 'asm.Program'.
	asmProgram, err := lowerer.Lowerer()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'lowering' pass: %s\n", err)
		return -1
	}

	// Now, instantiates a code generator for the Asm (compiled) program
	codegen := asm.NewCodeGenerator(asmProgram)
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

func main() { os.Exit(VmTranslator.Run(os.Args, os.Stdout)) }
