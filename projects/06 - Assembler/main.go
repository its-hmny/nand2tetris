package main

import (
	"fmt"
	"reflect"

	"its-hmny.dev/n2t-assembler/hack"
)

var program = []hack.Instruction{
	// Built-in & well known registries with resolution on the built-in table
	hack.AInstruction{LocType: hack.BuiltIn, LocName: "R1"},
	hack.AInstruction{LocType: hack.BuiltIn, LocName: "R2"},
	// User-defined label with simple address binary translation
	hack.AInstruction{LocType: hack.Raw, LocName: "38"},
	hack.AInstruction{LocType: hack.Raw, LocName: "42"},
	hack.AInstruction{LocType: hack.Raw, LocName: "64"},
	hack.AInstruction{LocType: hack.Raw, LocName: "128"},
	hack.AInstruction{LocType: hack.Raw, LocName: fmt.Sprint(hack.MaxAddressAllowed - 1)},
	// User-defined labels with resolution on the symbol table
	hack.AInstruction{LocType: hack.Label, LocName: "Test1"},
	hack.AInstruction{LocType: hack.Label, LocName: "Test2"},
	hack.AInstruction{LocType: hack.Label, LocName: "Test3"},
}

var table = map[string]uint16{
	"Test1": 86, "Test2": 256, "Test3": 24,
}

func main() {
	fmt.Println("============== nand2tetris Hack Assembler ==============")

	translator := hack.CodeGenerator{Program: program, Table: table}
	compiled, err := translator.Dump()

	if err != nil {
		fmt.Printf("ERR: Unable to complete 'codegen' pass:\n\t %s", err)
	}

	// For the time being we simply dump the program on stdout before exiting, each
	// and every instruction is printed both in its in-memory format and raw binary
	for n := range compiled {
		assembler, hack := program[n], compiled[n]

		fmt.Printf("%d) %s: =>\n", n, reflect.TypeOf(assembler).Name())
		fmt.Printf("\t\t %+v\n", assembler)
		fmt.Printf("\t\t %s\n", hack)
	}
}
