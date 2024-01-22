package asm_test

import (
	"testing"

	"its-hmny.dev/nand2tetris/pkg/asm"
)

func TestAInstructions(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := asm.NewCodeGenerator([]asm.Statement{})

	test := func(inst asm.AInstruction, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateAInst(inst)
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
		test(asm.AInstruction{Location: "38"}, "@38", false)
		test(asm.AInstruction{Location: "42"}, "@42", false)
		test(asm.AInstruction{Location: "64"}, "@64", false)
		test(asm.AInstruction{Location: "1024"}, "@1024", false)
		// This are just some example of invalid (Out of Bounds) address that shouldn't be translated.
		test(asm.AInstruction{Location: "32768"}, "", true)
		test(asm.AInstruction{Location: "65538"}, "", true)
		test(asm.AInstruction{Location: "66500"}, "", true)
		test(asm.AInstruction{Location: "70000"}, "", true)
	})

	t.Run("Hack built-in labels", func(t *testing.T) {
		// Named specific purpose registries
		test(asm.AInstruction{Location: "SP"}, "@SP", false)
		test(asm.AInstruction{Location: "LCL"}, "@LCL", false)
		test(asm.AInstruction{Location: "ARG"}, "@ARG", false)
		test(asm.AInstruction{Location: "THIS"}, "@THIS", false)
		test(asm.AInstruction{Location: "THAT"}, "@THAT", false)
		// Named general purpose registers (R0 to R15)
		test(asm.AInstruction{Location: "R0"}, "@R0", false)
		test(asm.AInstruction{Location: "R1"}, "@R1", false)
		test(asm.AInstruction{Location: "R2"}, "@R2", false)
		test(asm.AInstruction{Location: "R3"}, "@R3", false)
		test(asm.AInstruction{Location: "R4"}, "@R4", false)
		test(asm.AInstruction{Location: "R5"}, "@R5", false)
		test(asm.AInstruction{Location: "R6"}, "@R6", false)
		test(asm.AInstruction{Location: "R7"}, "@R7", false)
		test(asm.AInstruction{Location: "R8"}, "@R8", false)
		test(asm.AInstruction{Location: "R9"}, "@R9", false)
		test(asm.AInstruction{Location: "R10"}, "@R10", false)
		test(asm.AInstruction{Location: "R11"}, "@R11", false)
		test(asm.AInstruction{Location: "R12"}, "@R12", false)
		test(asm.AInstruction{Location: "R13"}, "@R13", false)
		test(asm.AInstruction{Location: "R14"}, "@R14", false)
		test(asm.AInstruction{Location: "R15"}, "@R15", false)
		// Memory mapped I/O address testing (SCREEN is a range but only the first byte is named)
		test(asm.AInstruction{Location: "KBD"}, "@KBD", false)
		test(asm.AInstruction{Location: "SCREEN"}, "@SCREEN", false)
	})

	t.Run("User-defined labels", func(t *testing.T) {
		// User defined labels that are present in the injected Symbol Table
		test(asm.AInstruction{Location: "Test1"}, "@Test1", false)
		test(asm.AInstruction{Location: "Test2"}, "@Test2", false)
		test(asm.AInstruction{Location: "hmny"}, "@hmny", false)
		test(asm.AInstruction{Location: "n2t"}, "@n2t", false)
		test(asm.AInstruction{Location: "JUMP"}, "@JUMP", false)
	})
}

func TestCInstructions(t *testing.T) {
	// Instantiate a shared codegen instance for every test cases
	codegen := asm.NewCodeGenerator([]asm.Statement{})

	test := func(inst asm.CInstruction, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateCInst(inst)
		// Each address always is exactly 16 bit long and should match the 'expected'
		if len(res) == 16 && res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Comps and Jumps", func(t *testing.T) {
		// Basic constant and identities operations with jump directives
		test(asm.CInstruction{Comp: "0", Jump: "JGT"}, "0;JGT", false)
		test(asm.CInstruction{Comp: "1", Jump: "JEQ"}, "1;JEQ", false)
		test(asm.CInstruction{Comp: "-1", Jump: "JEQ"}, "-1;JEQ", false)
		test(asm.CInstruction{Comp: "D", Jump: "JGE"}, "D;JGE", false)
		test(asm.CInstruction{Comp: "A", Jump: "JGT"}, "A;JGT", false)
		// Binary and numerical negation operations with jump directives
		test(asm.CInstruction{Comp: "!A", Jump: "JLT"}, "!A;JLT", false)
		test(asm.CInstruction{Comp: "!M", Jump: "JNE"}, "!M;JNE", false)
		test(asm.CInstruction{Comp: "-D", Jump: "JNE"}, "-D;JNE", false)
		test(asm.CInstruction{Comp: "-A", Jump: "JLE"}, "-A;JLE", false)
		test(asm.CInstruction{Comp: "-M", Jump: "JLE"}, "-M;JLE", false)
	})

	t.Run("Comps and Jumps", func(t *testing.T) {
		// Register with register operations with dest directives
		test(asm.CInstruction{Comp: "D-A", Dest: "M"}, "M=D-A", false)
		test(asm.CInstruction{Comp: "D-M", Dest: "M"}, "M=D-M", false)
		test(asm.CInstruction{Comp: "A-D", Dest: "D"}, "D=A-D", false)
		test(asm.CInstruction{Comp: "M-D", Dest: "D"}, "D=M-D", false)
		// Bitwise register with register operations with dest directives
		test(asm.CInstruction{Comp: "D&A", Dest: "A"}, "A=D&A", false)
		test(asm.CInstruction{Comp: "D&M", Dest: "A"}, "A=D&M", false)
		test(asm.CInstruction{Comp: "D|A", Dest: "MD"}, "MD=D|A", false)
		test(asm.CInstruction{Comp: "D|M", Dest: "MD"}, "MD=D|M", false)
		// Basic constant and identities operations with dest directives
		test(asm.CInstruction{Comp: "M", Dest: "AM"}, "AM=M", false)
		test(asm.CInstruction{Comp: "0", Dest: "AD"}, "AD=0", false)
		test(asm.CInstruction{Comp: "-1", Dest: "AMD"}, "AMD=-1", false)
		test(asm.CInstruction{Comp: "D", Dest: "AMD"}, "AMD=D", false)
		test(asm.CInstruction{Comp: "A", Dest: "AMD"}, "AMD=A", false)

	})

	t.Run("Malformed Inst", func(t *testing.T) {
		// Comp only C Instruction, should fail and return an error
		test(asm.CInstruction{Comp: "D+1", Jump: ""}, "", true)
		test(asm.CInstruction{Comp: "A+1", Jump: ""}, "", true)
		test(asm.CInstruction{Comp: "A-1", Jump: ""}, "", true)
		test(asm.CInstruction{Comp: "M-1", Jump: ""}, "", true)
		// Comp only C Instruction, should fail and return an error
		test(asm.CInstruction{Comp: "A", Dest: ""}, "", true)
		test(asm.CInstruction{Comp: "1", Dest: ""}, "", true)
		test(asm.CInstruction{Comp: "D+1", Dest: ""}, "", true)
		test(asm.CInstruction{Comp: "A+1", Dest: ""}, "", true)
		// C Instruction with either 'Dest' or 'Jump' or both but not 'Comp'
		test(asm.CInstruction{Dest: "AM", Jump: "JNE"}, "", true)
		test(asm.CInstruction{Dest: "AD", Jump: "JEQ"}, "", true)
		test(asm.CInstruction{Dest: "AMD", Jump: ""}, "", true)
		test(asm.CInstruction{Dest: "AMD", Jump: ""}, "", true)
		test(asm.CInstruction{Dest: "", Jump: "JGT"}, "", true)
		test(asm.CInstruction{Dest: "", Jump: "JLT"}, "", true)
	})
}

func TestLabelDecl(t *testing.T) {
	// Instantiate a shared codegen instance for every test cases
	codegen := asm.NewCodeGenerator([]asm.Statement{})

	test := func(inst asm.LabelDecl, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateLabelDecl(inst)
		// Each address always is exactly 16 bit long and should match the 'expected'
		if len(res) == 16 && res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Fuzzy labels", func(t *testing.T) {
		// Fuzzy label declaration
		test(asm.LabelDecl{Name: "test123"}, "(test123)", false)
		test(asm.LabelDecl{Name: "ping"}, "(ping)", false)
		test(asm.LabelDecl{Name: "PONG"}, "(PONG)", false)
		test(asm.LabelDecl{Name: "TEST"}, "(TEST)", false)
		test(asm.LabelDecl{Name: "DUNNO"}, "(DUNNO)", false)
		// Malformed or conflicting label generation
		test(asm.LabelDecl{Name: ""}, "", true)
		test(asm.LabelDecl{Name: "SP"}, "", true)
		test(asm.LabelDecl{Name: "R1"}, "", true)
		test(asm.LabelDecl{Name: "LCL"}, "", true)
		test(asm.LabelDecl{Name: "R15"}, "", true)
	})
}
