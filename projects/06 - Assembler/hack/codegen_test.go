package hack_test

import (
	"fmt"
	"testing"

	"its-hmny.dev/n2t-assembler/hack"
)

func TestAInstruction(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	table := map[string]uint16{"Test1": 0, "Test2": 67, "hmny": 9393, "n2t": 754, "JUMP": 90}
	codegen := hack.CodeGenerator{Program: []hack.Instruction{}, Table: table}

	test := func(inst hack.AInstruction, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.TranslateAInst(inst)
		// Each address always is exactly 16 bit long and should match the 'expected'
		if len(res) == 16 && res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Raw memory access", func(t *testing.T) {
		// This A Instruction reference correct raw location/address, to be correct a raw address
		// must be strictly below 2^16, since onl 15 bits are available to index the Hack memory.
		test(hack.AInstruction{LocType: hack.Raw, LocName: "38"}, fmt.Sprintf("%016b", 38), false)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "42"}, fmt.Sprintf("%016b", 42), false)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "64"}, fmt.Sprintf("%016b", 64), false)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "128"}, fmt.Sprintf("%016b", 128), false)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "32767"}, fmt.Sprintf("%016b", 32767), false)
		// This are just some example of invalid (Out of Bounds) address that shouldn't be translated.
		test(hack.AInstruction{LocType: hack.Raw, LocName: "32768"}, "", true)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "65538"}, "", true)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "66500"}, "", true)
		test(hack.AInstruction{LocType: hack.Raw, LocName: "70000"}, "", true)
	})

	t.Run("Hack built-in labels", func(t *testing.T) {
		// Named specific purpose registries
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "SP"}, fmt.Sprintf("%016b", 0), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "LCL"}, fmt.Sprintf("%016b", 1), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "ARG"}, fmt.Sprintf("%016b", 2), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "THIS"}, fmt.Sprintf("%016b", 3), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "THAT"}, fmt.Sprintf("%016b", 4), false)
		// Named general purpose registers (R0 to R15)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R0"}, fmt.Sprintf("%016b", 0), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R1"}, fmt.Sprintf("%016b", 1), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R2"}, fmt.Sprintf("%016b", 2), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R3"}, fmt.Sprintf("%016b", 3), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R4"}, fmt.Sprintf("%016b", 4), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R5"}, fmt.Sprintf("%016b", 5), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R6"}, fmt.Sprintf("%016b", 6), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R7"}, fmt.Sprintf("%016b", 7), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R8"}, fmt.Sprintf("%016b", 8), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R9"}, fmt.Sprintf("%016b", 9), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R10"}, fmt.Sprintf("%016b", 10), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R11"}, fmt.Sprintf("%016b", 11), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R12"}, fmt.Sprintf("%016b", 12), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R13"}, fmt.Sprintf("%016b", 13), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R14"}, fmt.Sprintf("%016b", 14), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "R15"}, fmt.Sprintf("%016b", 15), false)
		// Memory mapped I/O address testing (SCREEN is a range but only the first byte is named)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "KBD"}, fmt.Sprintf("%016b", 24576), false)
		test(hack.AInstruction{LocType: hack.BuiltIn, LocName: "SCREEN"}, fmt.Sprintf("%016b", 16384), false)
	})

	t.Run("User-defined labels", func(t *testing.T) {
		// User defined labels that are present in the injected Symbol Table
		test(hack.AInstruction{LocType: hack.Label, LocName: "Test1"}, fmt.Sprintf("%016b", table["Test1"]), false)
		test(hack.AInstruction{LocType: hack.Label, LocName: "Test2"}, fmt.Sprintf("%016b", table["Test2"]), false)
		test(hack.AInstruction{LocType: hack.Label, LocName: "hmny"}, fmt.Sprintf("%016b", table["hmny"]), false)
		test(hack.AInstruction{LocType: hack.Label, LocName: "n2t"}, fmt.Sprintf("%016b", table["n2t"]), false)
		test(hack.AInstruction{LocType: hack.Label, LocName: "JUMP"}, fmt.Sprintf("%016b", table["JUMP"]), false)
		// User defined labels that are not available in the Symbol Table and should cause an error
		test(hack.AInstruction{LocType: hack.Label, LocName: "NOP"}, "", true)
		test(hack.AInstruction{LocType: hack.Label, LocName: "MISSING"}, "", true)
		test(hack.AInstruction{LocType: hack.Label, LocName: "NEXT"}, "", true)
		test(hack.AInstruction{LocType: hack.Label, LocName: "DUNNO"}, "", true)
	})
}
