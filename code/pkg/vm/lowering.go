package vm

import (
	"fmt"

	"its-hmny.dev/nand2tetris/pkg/asm"
)

var SegmentTable = map[SegmentType]string{
	// Stack Segment mapped to their own 'asm' labels
	Local: "LCL", Argument: "ARG", This: "THIS", That: "THAT",
	// Stack Segment mapped to static raw location in 'asm'
	Pointer: "3", Temp: "5",
}

var ArithmeticTable = map[ArithOpType]func(uint) []asm.Instruction{
	// Mappers to []asm.Instruction for the comparison operations in VM language (eq, gt, lt)
	Eq: func(counter uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			asm.AInstruction{Location: fmt.Sprintf("EQUAL_%d", counter)},
			asm.CInstruction{Comp: "D", Jump: "JEQ"},
			asm.CInstruction{Dest: "D", Comp: "0"},
			asm.AInstruction{Location: fmt.Sprintf("END_%d", counter)},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
			asm.LabelDecl{Name: fmt.Sprintf("EQUAL_%d", counter)},
			asm.CInstruction{Dest: "D", Comp: "-1"},
			asm.LabelDecl{Name: fmt.Sprintf("END_%d", counter)},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
	Gt: func(counter uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			asm.AInstruction{Location: fmt.Sprintf("GREATER_%d", counter)},
			asm.CInstruction{Comp: "D", Jump: "JLT"},
			asm.CInstruction{Dest: "D", Comp: "0"},
			asm.AInstruction{Location: fmt.Sprintf("END_%d", counter)},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
			asm.LabelDecl{Name: fmt.Sprintf("GREATER_%d", counter)},
			asm.CInstruction{Dest: "D", Comp: "-1"},
			asm.LabelDecl{Name: fmt.Sprintf("END_%d", counter)},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
	Lt: func(counter uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			asm.AInstruction{Location: fmt.Sprintf("LESS_%d", counter)},
			asm.CInstruction{Comp: "D", Jump: "JGT"},
			asm.CInstruction{Dest: "D", Comp: "0"},
			asm.AInstruction{Location: fmt.Sprintf("END_%d", counter)},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
			asm.LabelDecl{Name: fmt.Sprintf("LESS_%d", counter)},
			asm.CInstruction{Dest: "D", Comp: "-1"},
			asm.LabelDecl{Name: fmt.Sprintf("END_%d", counter)},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Mappers to []asm.Instruction for the arithmetic operations in VM language (add, sub, neg)
	Add: func(uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "D+M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
	Sub: func(uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
	Neg: func(uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "-D"},
		}
	},

	// Mappers to []asm.Instruction for the bitwise operations in VM language (not, and, or)
	Not: func(uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "!D"},
		}
	},
	And: func(uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D&M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
	Or: func(uint) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D|M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
}

// ----------------------------------------------------------------------------
// Vm Lowerer

// The Lowerer takes a 'vm.Program' and produces its 'asm.Program' counterpart.
//
// Since we get a tree-like struct we are able to traverse it using a Depth First Search (DFS) algorithm
// on it. For each operation node visited we produce a list of 'hack.Instruction' as counterpart (either
// A Instruction, C Instruction or LabelDecl) as well as validating the input before proceeding.
type Lowerer struct {
	program Program

	// Keeps track of the module (.vm file) we're lowering at the moment
	// Used to randomize and make unique the static variables during lowering
	vmModule string
	// Keeps track of the scope (either global or function) we're lowering at the moment
	// Used to randomize and make unique the label declaration during lowering
	vmScope string

	nRandomizer uint // Counter to randomize 'asm.LabelDecl(s)' with same name
}

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument Program to be not nil nor empty.
func NewLowerer(p Program) Lowerer {
	return Lowerer{program: p, vmScope: "global"}
}

// Triggers the lowering process. It iterates operation by operation and recursively
// calls the specified helper function based on the operation type (much like a recursive
// descend parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lowerer() (asm.Program, error) {
	program := []asm.Instruction{}

	if l.program == nil || len(l.program) == 0 {
		return nil, fmt.Errorf("the given 'program' is empty")
	}

	for name, module := range l.program {
		l.vmModule = name // Updates the tracker, signaling we're lowering another module

		for _, op := range module {
			switch tOp := op.(type) {
			case MemoryOp: // Converts 'vm.MemoryOp' to a list of 'asm.Instruction'
				inst, err := l.HandleMemoryOp(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			case ArithmeticOp: // Converts 'vm.ArithmeticOp' to a list of 'asm.Instruction'
				inst, err := l.HandleArithmeticOp(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			case LabelDeclaration: // Converts 'vm.LabelDeclaration' to a list of 'asm.Instruction'
				inst, err := l.HandleLabelDecl(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			case GotoOp: // Converts 'vm.GotoOp' to a list of 'asm.Instruction'
				inst, err := l.HandleGotoOp(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			case FuncDecl: // Converts 'vm.FuncDecl' to a list of 'asm.Instruction'
				inst, err := l.HandleFuncDecl(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			case ReturnOp: // Converts 'vm.ReturnOp' to a list of 'asm.Instruction'
				inst, err := l.HandleReturnOp(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			default: // Error case, unrecognized operation type
				return nil, fmt.Errorf("unrecognized operation '%T'", tOp)
			}
		}
	}

	return program, nil
}

// Specialized function to convert a 'vm.MemoryOp' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleMemoryOp(op MemoryOp) ([]asm.Instruction, error) {
	switch op.Operation {
	case Pop:
		return l.HandlePopOp(op)
	case Push:
		return l.HandlePushOp(op)
	default:
		return nil, fmt.Errorf("unrecognized MemoryOp instruction %s", op.Operation)
	}
}

// Specialized function to convert a 'vm.MemoryOp' (subtype Push) node to a list of 'asm.Instruction'.
func (l *Lowerer) HandlePushOp(op MemoryOp) ([]asm.Instruction, error) {
	translated := []asm.Instruction{} // Accumulator of the translated instructions

	if op.Segment == Constant {
		translated = append(translated,
			// Takes the raw location with the A Instruction, saves A reg on D reg
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			// Saves on D the M reg value, then copies it on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "A"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	if op.Segment == Local || op.Segment == Argument || op.Segment == This || op.Segment == That {
		label, found := SegmentTable[op.Segment]
		if !found {
			return nil, fmt.Errorf("could not map %s to Asm label", op.Segment)
		}

		translated = append(translated,
			// Takes the base pointer for the requested segment
			asm.AInstruction{Location: label},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Saves on D the M reg value, then copies it on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	if op.Segment == Pointer || op.Segment == Temp {
		label, found := SegmentTable[op.Segment]
		if !found {
			return nil, fmt.Errorf("could not map %s to Asm label", op.Segment)
		}

		translated = append(translated,
			// Takes the raw mapped location for the requested segment
			asm.AInstruction{Location: label},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Saves on D the M reg value, then copies it on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	if op.Segment == Static {
		translated = append(translated,
			asm.AInstruction{Location: fmt.Sprintf("%s.%d", l.vmModule, op.Offset)},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	// This is the set of operations that is common to every push on the stack.
	// We expect that on the D register will have the value to push and proceed accordingly.
	translated = append(translated,
		// Takes out the value from R13 and saves onto the D reg
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		// Takes SP and goto it location,
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		// Saves on M the D result
		asm.CInstruction{Dest: "M", Comp: "D"},
		// Increments SP to new memory location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
	)

	return translated, nil
}

// Specialized function to convert a 'vm.MemoryOp' (subtype Pop) node to a list of 'asm.Instruction'.
func (l *Lowerer) HandlePopOp(op MemoryOp) ([]asm.Instruction, error) {
	translated := []asm.Instruction{}

	if op.Segment == Constant {
		return nil, fmt.Errorf("cannot push on read-only segment 'constant'")
	}

	if op.Segment == Local || op.Segment == Argument || op.Segment == This || op.Segment == That {
		label, found := SegmentTable[op.Segment]
		if !found {
			return nil, fmt.Errorf("could not map %s to Asm instructions", op.Segment)
		}

		translated = append(translated,
			// Takes the base pointer for the requested segment
			asm.AInstruction{Location: label},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves on D on for usage by the next instruction
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	if op.Segment == Pointer || op.Segment == Temp {
		label, found := SegmentTable[op.Segment]
		if !found {
			return nil, fmt.Errorf("could not map %s to Asm instructions", op.Segment)
		}

		translated = append(translated,
			// Takes the base pointer for the requested segment
			asm.AInstruction{Location: label},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves on D on for usage by the next instruction
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	if op.Segment == Static {
		translated = append(translated,
			asm.AInstruction{Location: fmt.Sprintf("%s.%d", l.vmModule, op.Offset)},
			asm.CInstruction{Dest: "D", Comp: "A"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	// This is the set of operations that is common to every pop on the stack.
	// We save on the D register the value to be stored on the heap and proceed based on the specific segment.
	translated = append(translated,
		// Takes SP and goto its location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "AM", Comp: "M-1"},
		// Saves on D the M reg value, then copies it on R13 (for persistence)
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
	)

	return translated, nil
}

// Specialized function to convert a 'vm.ArithmeticOp' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleArithmeticOp(op ArithmeticOp) ([]asm.Instruction, error) {
	prelude := []asm.Instruction{
		// Takes SP and goto it location (also decrementing it)
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "AM", Comp: "M-1"},
		// Saves onto D the value and then copies it onto R13
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "M", Comp: "D"},
	}

	// For every binary operation we push the second operand onto R14 reg
	if op.Operation != Not && op.Operation != Neg {
		prelude = append(prelude,
			// Takes SP and goto it location (also decrementing it)
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "AM", Comp: "M-1"},
			// Saves onto D the value and then copies it onto  R14
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	// The 'arithmetic' section does the computation and stores everything on R15.
	arithmetic, found := ArithmeticTable[op.Operation]
	if !found {
		return nil, fmt.Errorf("could not map %s to Asm instructions", op.Operation)
	}

	if op.Operation == Eq || op.Operation == Lt || op.Operation == Gt {
		l.nRandomizer += 1
	}

	// The 'postlude' section takes the value in R15 and push it onto the Stack
	postlude := []asm.Instruction{
		// Takes out the value from R15 and saves onto the D reg
		asm.AInstruction{Location: "R15"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		// Takes SP and goto it location,
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		// Saves on M the D result
		asm.CInstruction{Dest: "M", Comp: "D"},
		// Increments SP to new memory location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
	}

	return append(append(prelude, arithmetic(l.nRandomizer)...), postlude...), nil
}

// Specialized function to convert a 'vm.LabelDeclaration' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleLabelDecl(op LabelDeclaration) ([]asm.Instruction, error) {
	if op.Name == "" {
		return nil, fmt.Errorf("unexpected empty label value")
	}
	if l.vmScope == "" {
		return nil, fmt.Errorf("unexpected empty 'vmScope' value")
	}

	// The vm.LabelDecl is scoped to either the function or the global scope, by appending the
	// name of the current scope as prefix we 'implement' this scoping in the asm counterpart
	// that doesn't support this kind of high-level constructs (it has a unified global scope).
	return []asm.Instruction{asm.LabelDecl{Name: fmt.Sprintf("%s$%s", l.vmScope, op.Name)}}, nil
}

// Specialized function to convert a 'vm.GotoOp' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleGotoOp(op GotoOp) ([]asm.Instruction, error) {
	if op.Label == "" {
		return nil, fmt.Errorf("unexpected empty label value")
	}
	if l.vmScope == "" {
		return nil, fmt.Errorf("unexpected empty 'vmScope' value")
	}

	if op.Jump == Conditional {
		return []asm.Instruction{
			// Takes SP, goto its location and decrements it
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "AM", Comp: "M-1"},
			// Saves on D the M reg value for later
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Loads the jump location, since labels are scoped in VM language we also 'randomize' it.
			asm.AInstruction{Location: fmt.Sprintf("%s$%s", l.vmScope, op.Label)},
			// Makes the jump if D reg contains a 'truthy' value (different from 0)
			asm.CInstruction{Comp: "D", Jump: "JGT"},
		}, nil
	}

	if op.Jump == Unconditional {
		return []asm.Instruction{
			// Loads the jump location, since labels are scoped in VM language we also 'randomize' it.
			asm.AInstruction{Location: fmt.Sprintf("%s$%s", l.vmScope, op.Label)},
			// Makes the jump always (unconditionally)
			asm.CInstruction{Comp: "0", Jump: "JMP"},
		}, nil
	}

	return nil, fmt.Errorf("unrecognized jump type, got %s", op.Jump)
}

// Specialized function to convert a 'vm.FuncDecl' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleFuncDecl(op FuncDecl) ([]asm.Instruction, error) {
	if op.Name == "" {
		return nil, fmt.Errorf("unexpected empty 'FuncDecl.Name' value")
	}

	// Allocates first the label that will represent the entrypoint of the function
	prelude := []asm.Instruction{asm.LabelDecl{Name: op.Name}}

	// Wipes clean the 'local' segment section with zeroes (just enough as indicated by the FuncDecl)
	for offset := range op.ArgsNum {
		prelude = append(prelude,
			// Takes the LCL pointer location for the 'local' segment
			asm.AInstruction{Location: "LCL"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Saves at that location the zero (wiping out the memory)
			asm.CInstruction{Dest: "M", Comp: "0"})
	}

	return prelude, nil
}

// Specialized function to convert a 'vm.ReturnOp' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleReturnOp(op ReturnOp) ([]asm.Instruction, error) {
	postlude := []asm.Instruction{
		// ? FRAME = LCL
		asm.AInstruction{Location: "LCL"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		// On R13 we keep the temporary frame pointer (on the book named as FRAME)
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// ? RET = *(FRAME-5)
		asm.AInstruction{Location: "5"},
		asm.CInstruction{Dest: "A", Comp: "D-A"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		// On R14 we keep the temporary return address (on the book named as RET)
		asm.AInstruction{Location: "R14"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// ? *ARG = pop()
		asm.AInstruction{Location: "ARG"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		// Saves on D on for usage by the next instruction
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// Takes SP and goto its location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "AM", Comp: "M-1"},
		// Saves on D the M reg value, then copies it on R13 (for persistence)
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// ? SP = ARG+1
		// ARG[0] will coincide with the top of the stack (also SP will point here).
		asm.AInstruction{Location: "ARG"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "D+1"},
	}

	// ? THAT = *(FRAME-1), THIS = *(FRAME-2), ARG = *(FRAME-3), LCL = *(FRAME-4)
	for offset, segment := range []string{"THAT", "THIS", "ARG", "LCL"} {
		postlude = append(postlude,
			asm.AInstruction{Location: "LCL"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: fmt.Sprint(offset + 1)},
			asm.CInstruction{Dest: "A", Comp: "D-A"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: segment},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	postlude = append(postlude,
		// ? goto RET
		asm.AInstruction{Location: "R14"},
		asm.CInstruction{Comp: "0", Jump: "JMP"},
	)

	return postlude, nil
}
