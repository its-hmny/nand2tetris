package vm

import (
	"fmt"

	"its-hmny.dev/nand2tetris/pkg/asm"
)

// ----------------------------------------------------------------------------
// Translation tables

// This section contains the translation tables, cornerstone of the lowering phase.
//
// This table provides a simple yet effective way to map every memory and arithmetic operations,
// that both supports respectively memory segment and arithmetic intrinsics.
// Notably, we have a the following tables defined:
//   - 'PushTable': Specifies how to translate Push MemoryOp to their (segment specific) asm code
//   - 'PopTable': Specifies how to translate Pop MemoryOp to their (segment specific) asm code
//   - 'ArithmeticTable': Specifies how to all the arithmetic operation to their asm code

// The 'PushTable' as the name suggests allows to map a memory push operation for a specific segment
// to their asm counterpart, the convention here is that the value is taken from the memory location
// and saves it on the 'well-known' R13 register (reserved for internal usage) so that it can be
// pushed on the stack by a shared and reusable piece of codegen.
var PushTable = map[SegmentType]func(uint, string) []asm.Instruction{
	// Direct access segment, go to the raw location and uses the A reg as value
	Constant: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Loads the raw memory location using direct access
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Copies teh value on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves the value
	Local: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "LCL"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Copies teh value on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves the value
	Argument: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "ARG"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Copies teh value on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves the value
	This: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "THIS"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Copies teh value on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves the value
	That: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "THAT"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Copies teh value on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Direct access segment, takes the raw location + offset and saves the value
	Pointer: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the raw memory location for the segment
			asm.AInstruction{Location: "3"},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Copies teh value on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Direct access segment, takes the raw location + offset and saves the value
	Temp: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the raw mapped location for the segment
			asm.AInstruction{Location: "5"},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Copies teh value on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Static access segment, declares a unique variable shared across the vm.Module
	Static: func(offset uint, module string) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: fmt.Sprintf("%s.%d", module, offset)},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
}

// The 'PopTable' as the name suggests allows to map a memory pop operation for a specific segment
// to their asm counterpart, the convention here is that the memory location where we have to write
// the value on the top of the stack pointer, this location is saved on the 'well-known' R13 register
// (reserved for internal usage) so that it can be written by a reusable piece of codegen.
var PopTable = map[SegmentType]func(uint, string) []asm.Instruction{
	// Indirect access segment, takes the pointer + offset and saves it
	Local: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "LCL"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and saves the location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves it on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves it
	Argument: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "ARG"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and saves the location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves it on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves it
	This: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "THIS"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and saves the location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves it on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Indirect access segment, takes the pointer + offset and saves it
	That: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the base pointer for the segment
			asm.AInstruction{Location: "THAT"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and saves the location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves it on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Direct access segment, takes the raw location + offset and saves it
	Pointer: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the raw location for the segment
			asm.AInstruction{Location: "3"},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Adds the offset and saves the location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves it on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Direct access segment, takes the raw location + offset and saves it
	Temp: func(offset uint, _ string) []asm.Instruction {
		return []asm.Instruction{
			// Takes the raw location for the segment
			asm.AInstruction{Location: "5"},
			asm.CInstruction{Dest: "D", Comp: "A"},
			// Adds the offset and saves the location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "D", Comp: "D+A"},
			// Saves it on R13 (for persistence)
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	// Static access segment, declares a unique variable shared across the vm.Module
	Static: func(offset uint, module string) []asm.Instruction {
		return []asm.Instruction{
			asm.AInstruction{Location: fmt.Sprintf("%s.%d", module, offset)},
			asm.CInstruction{Dest: "D", Comp: "A"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},
}

// The 'ArithmeticTable' as the name suggests allows to map an arithmetic operation to a a set of
// specific asm instructions counterparts. The convention here is that the two operands are provided
// and saved respectively on the R13 and R14 registers while the result is saved in R15 (all of them
// reserved for internal usage) so that the remaining parts of the computation are op independent.
//
// NOTE: Comparison operation (Eq, Lt, Gt) rely on asm.LabelDecl in order to do their lowering and of
// course this kind of label have to eb unique to avoid jumping across the code like crazy when running
// the asm output, to do so the function accepts a 'counter' input that randomizes each label declaration.
var ArithmeticTable = map[ArithOpType]func(uint) []asm.Instruction{
	Eq: func(counter uint) []asm.Instruction {
		return []asm.Instruction{
			// Takes R13 and R14 and subtracts one from the other
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			// If (A - B) == 0 then goto EQUAL else goto END
			asm.AInstruction{Location: fmt.Sprintf("EQUAL_%d", counter)},
			asm.CInstruction{Comp: "D", Jump: "JEQ"},
			asm.CInstruction{Dest: "D", Comp: "0"},
			asm.AInstruction{Location: fmt.Sprintf("END_%d", counter)},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
			// Then branch R15 = 255
			asm.LabelDecl{Name: fmt.Sprintf("EQUAL_%d", counter)},
			asm.CInstruction{Dest: "D", Comp: "-1"},
			// Else branch R15 = 0
			asm.LabelDecl{Name: fmt.Sprintf("END_%d", counter)},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	Gt: func(counter uint) []asm.Instruction {
		return []asm.Instruction{
			// Takes R13 and R14 and subtracts one from the other
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			// If (A - B) > 0 then goto GREATER else goto END
			asm.AInstruction{Location: fmt.Sprintf("GREATER_%d", counter)},
			asm.CInstruction{Comp: "D", Jump: "JLT"},
			asm.CInstruction{Dest: "D", Comp: "0"},
			asm.AInstruction{Location: fmt.Sprintf("END_%d", counter)},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
			asm.LabelDecl{Name: fmt.Sprintf("GREATER_%d", counter)},
			// Then branch R15 = 255
			asm.CInstruction{Dest: "D", Comp: "-1"},
			asm.LabelDecl{Name: fmt.Sprintf("END_%d", counter)},
			// Else branch R15 = 0
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	Lt: func(counter uint) []asm.Instruction {
		return []asm.Instruction{
			// Takes R13 and R14 and subtracts one from the other
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D-M"},
			// If (A - B) < 0 then goto LESS else goto END
			asm.AInstruction{Location: fmt.Sprintf("LESS_%d", counter)},
			asm.CInstruction{Comp: "D", Jump: "JGT"},
			asm.CInstruction{Dest: "D", Comp: "0"},
			asm.AInstruction{Location: fmt.Sprintf("END_%d", counter)},
			asm.CInstruction{Comp: "0", Jump: "JMP"},
			// Then branch R15 = 255
			asm.LabelDecl{Name: fmt.Sprintf("LESS_%d", counter)},
			asm.CInstruction{Dest: "D", Comp: "-1"},
			asm.LabelDecl{Name: fmt.Sprintf("END_%d", counter)},
			// Else branch R15 = 0
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	Add: func(uint) []asm.Instruction {
		return []asm.Instruction{
			// Takes R13 and R14 and adds one to the other
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
			// Takes R13 and R14 and subtracts one from the other
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
			// Takes R13 and negates it
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "-D"},
		}
	},

	And: func(uint) []asm.Instruction {
		return []asm.Instruction{
			// Takes R13 and R14 and applies a bitwise and to one another
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
			// Takes R13 and R14 and applies a bitwise and to one another
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "D", Comp: "D|M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		}
	},

	Not: func(uint) []asm.Instruction {
		return []asm.Instruction{
			// Takes R13 and applies bitwise not to it
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R15"},
			asm.CInstruction{Dest: "M", Comp: "!D"},
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

// Triggers the lowering process. It iterates operation by operation and recursively calls
// the specified helper function based on the operation type (much like a recursive
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
				l.vmScope = tOp.Name
				program = append(program, inst...)

			case ReturnOp: // Converts 'vm.ReturnOp' to a list of 'asm.Instruction'
				inst, err := l.HandleReturnOp(tOp)
				if inst == nil || err != nil {
					return nil, err
				}
				program = append(program, inst...)

			case FuncCallOp: // Converts 'vm.FuncCallOp' to a list of 'asm.Instruction'
				inst, err := l.HandleFuncCallOp(tOp)
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
// Acts as a sort of 'dispatcher' between the Push and Pop OperationTypes that have
// really divergent underlying implementations (and asm counterparts),
func (l *Lowerer) HandleMemoryOp(op MemoryOp) ([]asm.Instruction, error) {
	switch op.Operation {
	case Pop:
		// Can't pop data onto the 'Constant' segment (is readonly of course)
		if op.Segment == Constant {
			return nil, fmt.Errorf("cannot push on read-only segment 'constant'")
		}

		// Retrieves the specific lowerer implementation based on the op.Segment
		generator, found := PopTable[op.Segment]
		if !found {
			return nil, fmt.Errorf("cannot find entry '%s' in lowering table", op.Segment)
		}

		// This is the set of operations that is common to every pop on the stack.
		// We save on the D register the value to be stored on the heap and proceed based on the specific segment.
		return append(generator(uint(op.Offset), l.vmModule),
			// Takes SP and goto its location
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "AM", Comp: "M-1"},
			// Saves on D the M reg value, then copies it on R13 (for persistence)
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "R13"},
			asm.CInstruction{Dest: "A", Comp: "M"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		), nil

	case Push:
		// Retrieves the specific lowerer implementation based on the op.Segment
		generator, found := PushTable[op.Segment]
		if !found {
			return nil, fmt.Errorf("cannot find entry '%s' in lowering table", op.Segment)
		}

		// This is the set of operations that is common to every push on the stack.
		// We expect that on the D register will have the value to push and proceed accordingly.
		return append(generator(uint(op.Offset), l.vmModule),
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
		), nil

	default:
		return nil, fmt.Errorf("unrecognized MemoryOp instruction %s", op.Operation)
	}
}

// Specialized function to convert a 'vm.ArithmeticOp' node to a list of 'asm.Instruction'.
func (l *Lowerer) HandleArithmeticOp(op ArithmeticOp) ([]asm.Instruction, error) {
	// We push the first operand onto R13 reg
	prelude := []asm.Instruction{
		// Decrements SP and goto its location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "AM", Comp: "M-1"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		// Saves the value onto R13
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "M", Comp: "D"},
	}

	// For every binary operation we push the second operand onto R14 reg
	if op.Operation != Not && op.Operation != Neg {
		prelude = append(prelude,
			// Decrements SP and goto its location
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "AM", Comp: "M-1"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Saves the value onto  R14
			asm.AInstruction{Location: "R14"},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	// If the op.Operation is a comparison one we have to 'randomize' the label
	if op.Operation == Eq || op.Operation == Lt || op.Operation == Gt {
		l.nRandomizer += 1
	}

	// Retrieves the specific lowerer implementation based on the op.Operation
	generator, found := ArithmeticTable[op.Operation]
	if !found {
		return nil, fmt.Errorf("could not map %s to Asm instructions", op.Operation)
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

	// Joins prelude + computation + postlude into a single slice
	return append(append(prelude, generator(l.nRandomizer)...), postlude...), nil
}

// Specialized function to convert a 'vm.LabelDeclaration' node to a list of 'asm.Instruction'.
// Manages the 'scoping' of the labels (a label is reachable only from within its declaration scope)
// during the lowering to the asm counterpart by prepending the label with the scope name.
func (l *Lowerer) HandleLabelDecl(op LabelDeclaration) ([]asm.Instruction, error) {
	if op.Name == "" { // Invariant: the label name should always be provided
		return nil, fmt.Errorf("unexpected empty label value")
	}
	if l.vmScope == "" { // Invariant: the scope name should always be provided
		return nil, fmt.Errorf("unexpected empty 'vmScope' value")
	}

	// The vm.LabelDecl is scoped to either the function or the global scope, by appending the name
	// of the current scope as prefix we 'implement' this scoping in the asm counterpart that doesn't
	// support this kind of high-level constructs (as it has a unified global scope/namespace).
	return []asm.Instruction{asm.LabelDecl{Name: fmt.Sprintf("%s$%s", l.vmScope, op.Name)}}, nil
}

// Specialized function to convert a 'vm.GotoOp' node to a list of 'asm.Instruction'.
// Manages the 'scoping' of the labels (a label is reachable only from within its declaration scope)
// during the lowering to the asm counterpart by prepending the label with the scope name.
func (l *Lowerer) HandleGotoOp(op GotoOp) ([]asm.Instruction, error) {
	if op.Label == "" { // Invariant: the label name should always be provided
		return nil, fmt.Errorf("unexpected empty label value")
	}
	if l.vmScope == "" { // Invariant: the scope name should always be provided
		return nil, fmt.Errorf("unexpected empty 'vmScope' value")
	}

	if op.Jump == Conditional {
		return []asm.Instruction{
			// Decrements the SP and goto the pointed location
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "AM", Comp: "M-1"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Loads the jump location, 'scoping' the label/destination.
			asm.AInstruction{Location: fmt.Sprintf("%s$%s", l.vmScope, op.Label)},
			// Makes the jump if D reg contains a 'truthy' value (!= 0)
			asm.CInstruction{Comp: "D", Jump: "JGT"},
		}, nil
	}

	if op.Jump == Unconditional {
		return []asm.Instruction{
			// Loads the jump location, 'scoping' the label/destination.
			asm.AInstruction{Location: fmt.Sprintf("%s$%s", l.vmScope, op.Label)},
			// Makes the unconditional jump (always jumps)
			asm.CInstruction{Comp: "0", Jump: "JMP"},
		}, nil
	}

	return nil, fmt.Errorf("unrecognized jump type, got %s", op.Jump)
}

// Specialized function to convert a 'vm.FuncDecl' node to a list of 'asm.Instruction'.
// The first instructions to be executed when calling a function cleans up the 'local' segment
// memory location with zeroes to avoid errors due to uninitialized memory (the number of
// local variables has to be predefined in the function declaration itself 'op.NLocals').
func (l *Lowerer) HandleFuncDecl(op FuncDecl) ([]asm.Instruction, error) {
	if op.Name == "" {
		return nil, fmt.Errorf("unexpected empty function name value")
	}

	// First, allocates the label for the function entrypoint
	translated := []asm.Instruction{asm.LabelDecl{Name: op.Name}}

	for offset := range op.NLocal { // Wipes the 'local' segment initializing all memory to 0
		translated = append(translated,
			// Takes the LCL pointer location for the 'local' segment
			asm.AInstruction{Location: "LCL"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A"},
			// Saves at that location the zero (wiping out the memory)
			asm.CInstruction{Dest: "M", Comp: "0"},
			// Increments the Stack Pointer to avoid double write errors
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "M", Comp: "M+1"},
		)
	}

	return translated, nil
}

// Specialized function to convert a 'vm.ReturnOp' node to a list of 'asm.Instruction'.
// When returning from a function call we have to restore all memory segments pointer
// back to the one used by the caller, while also pushing the return value on the stack.
func (l *Lowerer) HandleReturnOp(op ReturnOp) ([]asm.Instruction, error) {
	translated := []asm.Instruction{
		// We save the base frame pointer on R13 (the beginning of the call frame)
		asm.AInstruction{Location: "LCL"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// We save the return address on R14 (ReturnAddr = FrameBase-5)
		asm.AInstruction{Location: "5"},
		asm.CInstruction{Dest: "A", Comp: "D-A"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R14"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// Line-by-line translation of 'pop argument 0'
		asm.AInstruction{Location: "ARG"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "AM", Comp: "M-1"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "R13"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// We restore the Stack Pointer back to the top of the stack of the caller
		asm.AInstruction{Location: "ARG"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "D+1"},
	}

	// We restore as well the 'that', 'this', 'argument' and 'local' pointers back to the memory location
	// used by the caller (this address are saved by the caller when calling the current function)
	for offset, segment := range []string{"THAT", "THIS", "ARG", "LCL"} {
		translated = append(translated,
			asm.AInstruction{Location: "LCL"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: fmt.Sprint(offset + 1)},
			asm.CInstruction{Dest: "A", Comp: "D-A"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: segment},
			asm.CInstruction{Dest: "M", Comp: "D"},
		)
	}

	// At last, once everything is restored we jump back to the return address
	translated = append(translated,
		asm.AInstruction{Location: "R14"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Comp: "0", Jump: "JMP"},
	)

	return translated, nil
}

func (l *Lowerer) HandleFuncCallOp(op FuncCallOp) ([]asm.Instruction, error) {
	l.nRandomizer++
	return []asm.Instruction{
		// Takes the return address for the caller and push it on the stack
		asm.AInstruction{Location: fmt.Sprintf("%s-ret-%d", l.vmScope, l.nRandomizer)},
		asm.CInstruction{Dest: "D", Comp: "A"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
		// Takes the current 'local' segment pointer for the caller and push it on the stack
		asm.AInstruction{Location: "LCL"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
		// Takes the current 'argument' segment pointer for the caller and push it on the stack
		asm.AInstruction{Location: "ARG"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
		// Takes the current 'this' segment pointer for the caller and push it on the stack
		asm.AInstruction{Location: "THIS"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
		// Takes the current 'that' segment pointer for the caller and push it on the stack
		asm.AInstruction{Location: "THAT"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1"},
		// Sets the callee function 'argument' segment pointer to its location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "5"},
		asm.CInstruction{Dest: "D", Comp: "D-A"},
		asm.AInstruction{Location: fmt.Sprint(op.NArgs)},
		asm.CInstruction{Dest: "D", Comp: "D-A"},
		asm.AInstruction{Location: "ARG"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// Sets the callee function 'local' segment pointer to its location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.AInstruction{Location: "LCL"},
		asm.CInstruction{Dest: "M", Comp: "D"},
		// Transfer the execution control to the callee function with a jump to its entrypoint
		asm.AInstruction{Location: op.Name},
		asm.CInstruction{Comp: "0", Jump: "JMP"},
		// Declare a label that will reference the caller's return address
		asm.LabelDecl{Name: fmt.Sprintf("%s-ret-%d", l.vmScope, l.nRandomizer)},
	}, nil
}
