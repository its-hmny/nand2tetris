package vm_test

import (
	"testing"

	"its-hmny.dev/nand2tetris/pkg/vm"
)

func TestMemoryOp(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.MemoryOp, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateMemoryOp(inst)
		// Each address always is exactly 16 bit long and should match the 'expected'
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: 5}, "push constant 5", false)
		test(vm.MemoryOp{Operation: vm.Pop, Segment: vm.Local, Offset: 3}, "pop local 3", false)
		test(vm.MemoryOp{Operation: vm.Push, Segment: vm.Argument, Offset: 2}, "push argument 2", false)
		test(vm.MemoryOp{Operation: vm.Pop, Segment: vm.Static, Offset: 1}, "pop static 1", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
		// ! Since we have offset defined as uint16, and we use type-alias for string (think enums)
		// ! these are the only error we need to check for on the MemoryOp generation tests.
		// Offset 8 for temp segment is out of range (valid: 0-7), should fail
		test(vm.MemoryOp{Operation: vm.Push, Segment: vm.Temp, Offset: 8}, "", true)
		// Offset 0 for pointer segment is valid, should succeed
		test(vm.MemoryOp{Operation: vm.Pop, Segment: vm.Pointer, Offset: 2}, "", true)
		// Both operation and segment are invalid strings
		// ? test(vm.MemoryOp{Operation: vm.OperationType("randomOp"), Segment: vm.Constant, Offset: 0}, "", true)
		// ? test(vm.MemoryOp{Operation: vm.Push, Segment: vm.SegmentType("randomSegment"), Offset: 0}, "", true)
		// ? test(vm.MemoryOp{Operation: vm.OperationType("foo"), Segment: vm.SegmentType("bar"), Offset: 0}, "", true)
	})
}

func TestArithmeticOp(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.ArithmeticOp, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateArithmeticOp(inst)
		// Each address always is exactly 16 bit long and should match the 'expected'
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.ArithmeticOp{Operation: vm.Add}, "add", false)
		test(vm.ArithmeticOp{Operation: vm.Sub}, "sub", false)
		test(vm.ArithmeticOp{Operation: vm.Neg}, "neg", false)
		test(vm.ArithmeticOp{Operation: vm.Eq}, "eq", false)
		test(vm.ArithmeticOp{Operation: vm.Gt}, "gt", false)
		test(vm.ArithmeticOp{Operation: vm.Lt}, "lt", false)
		test(vm.ArithmeticOp{Operation: vm.And}, "and", false)
		test(vm.ArithmeticOp{Operation: vm.Or}, "or", false)
		test(vm.ArithmeticOp{Operation: vm.Not}, "not", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
		// ! Since we have Operation defined type-alias for string (think enums)
		// ! these are the only error we need to check for on the ArithmeticOp tests.
		// ? test(vm.ArithmeticOp{Operation: vm.ArithOpType("randomStr")}, "", true)
		// ? test(vm.ArithmeticOp{Operation: vm.ArithOpType("invalidOp")}, "", true)
		// ? test(vm.ArithmeticOp{Operation: vm.ArithOpType("123")}, "", true)
		// ? test(vm.ArithmeticOp{Operation: vm.ArithOpType("!@#")}, "", true)
		// ? test(vm.ArithmeticOp{Operation: vm.ArithOpType("add123")}, "", true)
	})
}

func TestLabelDecl(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.LabelDecl, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateLabelDecl(inst)
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.LabelDecl{Name: "END"}, "label END", false)
		test(vm.LabelDecl{Name: "CHECK"}, "label CHECK", false)
		test(vm.LabelDecl{Name: "LOOP_START"}, "label LOOP_START", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
		test(vm.LabelDecl{Name: ""}, "", true) // Empty label name
	})
}

func TestGotoOp(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.GotoOp, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateGotoOp(inst)
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.GotoOp{Jump: vm.Unconditional, Label: "END"}, "goto END", false)
		test(vm.GotoOp{Jump: vm.Conditional, Label: "CHECK"}, "if-goto CHECK", false)
		test(vm.GotoOp{Jump: vm.Unconditional, Label: "LOOP_START"}, "goto LOOP_START", false)
		test(vm.GotoOp{Jump: vm.Conditional, Label: "FUNC_RET"}, "if-goto FUNC_RET", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
		// ? test(vm.GotoOp{Jump: vm.JumpType(""), Label: "END"}, "", true)            // Empty jump type
		// ? test(vm.GotoOp{Jump: vm.JumpType("gibberish"), Label: "LABEL"}, "", true) // Gibberish jump type
		test(vm.GotoOp{Jump: vm.Unconditional, Label: ""}, "", true) // Empty label
		test(vm.GotoOp{Jump: vm.Conditional, Label: ""}, "", true)   // Empty label with valid jump
		// ? test(vm.GotoOp{Jump: vm.JumpType("ifgoto"), Label: "LABEL"}, "", true)    // Slightly wrong jump type
		// ? test(vm.GotoOp{Jump: vm.JumpType("goto!"), Label: "LABEL"}, "", true)     // Invalid jump type with symbol
	})
}

func TestFuncDecl(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.FuncDecl, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateFuncDecl(inst)
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.FuncDecl{Name: "Main", NLocal: 0}, "function Main 0", false)
		test(vm.FuncDecl{Name: "ComputeSum", NLocal: 2}, "function ComputeSum 2", false)
		test(vm.FuncDecl{Name: "LoopHandler", NLocal: 10}, "function LoopHandler 10", false)
		test(vm.FuncDecl{Name: "f", NLocal: 1}, "function f 1", false)
		test(vm.FuncDecl{Name: "VeryLongNameWithNumbers123", NLocal: 7}, "function VeryLongNameWithNumbers123 7", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
		test(vm.FuncDecl{Name: "", NLocal: 2}, "", true) // Empty function name
	})
}

func TestReturnOp(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.ReturnOp, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateReturnOp(inst)
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.ReturnOp{}, "return", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
	})
}

func TestFuncCallOp(t *testing.T) {
	// Instantiate a basic simple table with some entries and shared codegen for every test cases
	codegen := vm.NewCodeGenerator(vm.Program{})

	test := func(inst vm.FuncCallOp, expected string, fail bool) {
		// Run the translation function on the given A Instruction
		res, err := codegen.GenerateFuncCallOp(inst)
		if res != expected {
			t.Fail()
		}
		// 'err' should be not nil if 'fail' is passed as true from the caller
		if err != nil && !fail {
			t.Fail()
		}
	}

	t.Run("Valid data", func(t *testing.T) {
		test(vm.FuncCallOp{Name: "Main", NArgs: 0}, "call Main 0", false)
		test(vm.FuncCallOp{Name: "ComputeSum", NArgs: 2}, "call ComputeSum 2", false)
		test(vm.FuncCallOp{Name: "LoopHandler", NArgs: 10}, "call LoopHandler 10", false)
		test(vm.FuncCallOp{Name: "f", NArgs: 1}, "call f 1", false)
		test(vm.FuncCallOp{Name: "VeryLongNameWithNumbers123", NArgs: 7}, "call VeryLongNameWithNumbers123 7", false)
	})

	t.Run("Invalid data", func(t *testing.T) {
		test(vm.FuncCallOp{Name: "", NArgs: 2}, "", true) // Empty function name
	})
}
