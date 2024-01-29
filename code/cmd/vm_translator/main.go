package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/teris-io/cli"
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
	// ? WithArg(cli.NewArg("output", "The compiled binary output (.hack)")).
	WithAction(Handler)

func Handler(args []string, options map[string]string) int {
	input, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Printf("ERROR: Unable to open input file: %s\n", err)
		return -1
	}

	// ? output, err := os.Create(args[1])
	// ? if err != nil {
	// ? 	fmt.Printf("ERROR: Unable to open output file: %s\n", err)
	// ? 	return -1
	// ? }
	// ? defer output.Close()

	// Instantiate a parser for the vm program
	parser := vm.NewParser()
	// Parses the input file content and extract an AST from it
	_, success := parser.Parse(bytes.NewReader(input))
	if !success {
		fmt.Print("ERROR: Unable to complete 'parsing' pass\n")
		os.Exit(-1)
	}

	return 0
}

func main() { os.Exit(VmTranslator.Run(os.Args, os.Stdout)) }
