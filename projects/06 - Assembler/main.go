package main

import (
	"fmt"
	"reflect"

	"its-hmny.dev/n2t-assembler/hack"
)

var program = []hack.Instruction{
	// This first block should produce the sum of R1 + R2 and save it on R3
	hack.AInstruction{LocType: hack.BuiltIn, LocName: "R1"},
	hack.CInstruction{Comp: "M", Dest: "D"},
	hack.AInstruction{LocType: hack.BuiltIn, LocName: "R2"},
	hack.CInstruction{Comp: "D+M", Dest: "D"},
	hack.AInstruction{LocType: hack.BuiltIn, LocName: "R3"},
	hack.CInstruction{Comp: "D", Dest: "M"},
	hack.AInstruction{LocType: hack.Raw, LocName: "6"},
	hack.CInstruction{Comp: "0", Jump: "JMP"},
}

var table = map[string]uint16{
	"Test1": 86, "Test2": 256, "Test3": 24,
}

func main() {
	fmt.Println("============== nand2tetris Hack Assembler ==============")

	translator := hack.CodeGenerator{Program: program, Table: table}
	compiled, err := translator.Translate()

	if err != nil {
		fmt.Printf("ERR: Unable to complete 'codegen' pass:\n\t %s", err)
	}

	// For the time being we simply dump the program on stdout before exiting, each
	// and every instruction is printed both in its in-memory format and raw binary
	for n := range compiled {
		assembler, hack := program[n], compiled[n]

		fmt.Printf("%d) %s: =>\n", n, reflect.TypeOf(assembler).Name())
		fmt.Printf("\tAsm:  %+v\n", assembler)
		fmt.Printf("\tHack: %s\n", hack)
	}
	}
}
