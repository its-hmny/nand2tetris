package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
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
	// 'AsOptional()' allows to have more than one input .vm file
	WithArg(cli.NewArg("inputs", "The bytecode (.vm) file to be compiled").
		AsOptional().WithType(cli.TypeString)).
	WithOption(cli.NewOption("output", "The compiled binary output (.asm)").
		WithType(cli.TypeString)).
	WithOption(cli.NewOption("bootstrap", "Includes bootstrap code in the final .asm file").
		WithType(cli.TypeBool)).
	WithAction(Handler)

func Handler(args []string, options map[string]string) int {
	if len(args) < 1 || options["output"] == "" {
		fmt.Printf("ERROR: Not enough arguments provided, use --help\n")
		return -1
	}

	output, err := os.Create(options["output"])
	if err != nil {
		fmt.Printf("ERROR: Unable to open output file: %s\n", err)
		return -1
	}
	defer output.Close()

	// Allocates a 'vm.Program' struct to save all the parsed translation unit
	// (the .vm files) that will be parsed and lowered independently and then
	// sent to the codegen phases (that will create a monolithic compiled output).
	program := vm.Program{}

	// For every file provided by the user we do the following things
	for _, input := range args {
		content, err := os.ReadFile(input)
		if err != nil {
			fmt.Printf("ERROR: Unable to open input file: %s\n", err)
			return -1
		}

		// Instantiate a parser for the Vm program
		parser := vm.NewParser(bytes.NewReader(content))
		// Parses the input file content and extract an AST (as a 'vm.Module') from it.
		program[path.Base(input)], err = parser.Parse()
		if err != nil {
			fmt.Printf("ERROR: Unable to complete 'parsing' pass: %s\n", err)
			return -1
		}
	}

	// Instantiate a lowerer to convert the program from Vm to Asm
	lowerer := vm.NewLowerer(program)
	// Lowers the vm.Program to an in-memory/IR representation of its Asm counterpart 'asm.Program'.
	asmProgram, err := lowerer.Lowerer()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'lowering' pass: %s\n", err)
		return -1
	}

	// When the user opts in to include the 'bootstrap' code as the first instructions of our
	// translated program, this code does the following things:
	// - Sets the Stack Pointer to its base location at memory location 256
	// - Jump to the Sys.init function that (defined by the one of the 'vm.Module')
	if _, enabled := options["bootstrap"]; enabled {
		asmProgram = append([]asm.Instruction{
			asm.AInstruction{Location: "261"},
			asm.CInstruction{Dest: "D", Comp: "A"},
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "M", Comp: "D"},
			asm.AInstruction{Location: "Sys.init"},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
		}, asmProgram...)
	}

	// Now, instantiates a code generator for the Asm (compiled) program
	codegen := asm.NewCodeGenerator(asmProgram)
	// Iterates over each instruction and spits out the relative textual representation.
	compiled, err := codegen.Generate()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'codegen' pass: %s\n", err)
		return -1
	}

	for _, comp := range compiled {
		line := fmt.Sprintf("%s\n", comp)
		output.Write([]byte(line))
	}

	return 0
}

func main() { os.Exit(VmTranslator.Run(os.Args, os.Stdout)) }
