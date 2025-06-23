package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"its-hmny.dev/nand2tetris/pkg/jack"
	"its-hmny.dev/nand2tetris/pkg/vm"

	"github.com/teris-io/cli"
)

var Description = strings.ReplaceAll(`
TODO (hmny): Must add CLI tool description
`, "\n", " ")

var JackCompiler = cli.New(Description).
	// 'AsOptional()' allows to have more than one input .vm file
	WithArg(cli.NewArg("inputs", "The source (.jack) files to be compiled").
		AsOptional().WithType(cli.TypeString)).
	WithOption(cli.NewOption("stdlib", "Uses the built-in ABI of the standard library for lowering").
		WithType(cli.TypeBool)).
	WithAction(Handler)

func Handler(args []string, options map[string]string) int {
	if len(args) < 1 {
		fmt.Printf("ERROR: Not enough arguments provided, use --help\n")
		return -1
	}

	// The first is the aggregation of all the Translation Units (TUs) found during the input walk (just the paths)
	// The second is the container of the full program (a basic collection of parsed modules that can be explored)
	// ! While the Jack language spec follows the same semantic as Java every file is a class and every class is a
	// ! jack.Module, that said in future or other language the same could not apply. By TU we identify the source
	// ! that needs to be parsed, by module we identify the biggest entity extractable from said file. In jack this
	// ! a class but for other language it may be a module (Go), a namespace (C#) or just some basic functions (C).
	TUs, program := []string{}, jack.Program{}

	for _, input := range args {
		filepath.Walk(input, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(path) != ".jack" {
				return nil // We recurse on dirs and ignore other filetypes
			}

			TUs = append(TUs, path)
			return nil
		})
	}

	for _, tu := range TUs {
		content, err := os.ReadFile(tu)
		if err != nil {
			fmt.Printf("ERROR: Unable to open input file: %s\n", err)
			return -1
		}

		// Instantiate a parser for the Vm program
		parser := jack.NewParser(bytes.NewReader(content))
		// Parses the input file content and extract an AST (as a 'vm.Module') from it.
		program[path.Base(tu)], err = parser.Parse()
		if err != nil {
			fmt.Printf("ERROR: Unable to complete 'parsing' pass: %s\n", err)
			return -1
		}
	}

	// Adds to the jack.Program the stdlib ABI, this will help resolve stdlib functions w/o adding
	// them to the final executable (they are ignored after the codegen phase). This will enable
	// in future to compile project w/o defining the stdlib and assuming it can be 'linked' if needed.
	if _, enabled := options["stdlib"]; enabled {
		for name, class := range jack.StandardLibraryABI {
			program[name] = class
		}
	}

	// Instantiate a lowerer to convert the program from Jack to Vm
	lowerer := jack.NewLowerer(program)
	// Lowers the jack.Program to an in-memory/IR representation of its Vm counterpart 'vm.Program'.
	vmProgram, err := lowerer.Lowerer()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'lowering' pass: %s\n", err)
		return -1
	}

	// Now, instantiates a code generator for the Vm (compiled) program
	codegen := vm.NewCodeGenerator(vmProgram)
	// Iterates over each instruction and spits out the relative textual representation.
	compiled, err := codegen.Generate()
	if err != nil {
		fmt.Printf("ERROR: Unable to complete 'codegen' pass: %s\n", err)
		return -1
	}

	for _, tu := range TUs {
		module, ok := compiled[path.Base(tu)]
		if !ok {
			fmt.Printf("ERROR: Unable to compiled module for class file '%s'\n", tu)
			return -1
		}

		output, err := os.Create(fmt.Sprintf("%s.vm", strings.TrimSuffix(tu, ".jack")))
		if err != nil {
			fmt.Printf("ERROR: Unable to open output file: %s\n", err)
			return -1
		}
		defer output.Close()

		for _, ops := range module {
			line := fmt.Sprintf("%s\n", ops)
			output.Write([]byte(line))
		}
	}

	return 0
}

func main() { os.Exit(JackCompiler.Run(os.Args, os.Stdout)) }
