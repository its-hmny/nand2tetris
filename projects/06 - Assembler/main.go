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
	M=D+1;JGT
@TEST
A=M+1;JEQ
	
	(END)
		@END
		D=0; JMP
`

var HackProgram = []hack.Instruction{
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

var SymbolTable = map[string]uint16{
	"Test1": 86, "Test2": 256, "Test3": 24,
}

func main() {
	fmt.Println("============== nand2tetris Hack Assembler ==============")

	fmt.Println("================== Assembler parsing ===================")
	parser := assembler.Parser{Reader: strings.NewReader(AsmProgram)}

	status, err := parser.Parse()
	if !status || err != nil {
		fmt.Printf("ERR: Unable to complete 'codegen' pass:\n\t %s", err)
	}

	fmt.Println("==================== Hack codegen =====================")
	translator := hack.CodeGenerator{Program: HackProgram, Table: SymbolTable}

	compiled, err := translator.Translate()
	if err != nil {
		fmt.Printf("ERR: Unable to complete 'codegen' pass:\n\t %s", err)
	}

	// For the time being we simply dump the program on stdout before exiting, each
	// and every instruction is printed both in its in-memory format and raw binary
	for n := range compiled {
		assembler, hack := HackProgram[n], compiled[n]

		fmt.Printf("%d) %s: =>\n", n, reflect.TypeOf(assembler).Name())
		fmt.Printf(" Asm:  %+v\n", assembler)
		fmt.Printf(" Hack: %s\n\n", hack)
	}

	fmt.Println("==> Dumping compilation output to file...")

	out, _ := os.Create("./Test.hack")
	defer out.Close()

	for _, inst := range compiled {
		out.Write([]byte(fmt.Sprintf("%s\n", inst)))
	}
}
