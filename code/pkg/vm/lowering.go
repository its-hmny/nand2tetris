package vm

import (
	"fmt"

	"its-hmny.dev/nand2tetris/pkg/asm"
)

var (
	SegmentTable = map[SegmentType]string{Local: "LCL", Argument: "ARG", This: "THIS", That: "THAT"}
)

// ----------------------------------------------------------------------------
// Vm Lowerer

// The Lowerer takes a 'vm.Program' and produces its 'asm.Program' counterpart.
//
// Since we get a tree-like struct we are able to traverse it using a Depth First Search (DFS) algorithm
// on it. For each operation node visited we produce a list of 'hack.Instruction' as counterpart (either
// A Instruction, C Instruction or LabelDecl) as well as validating the input before proceeding.
type Lowerer struct{ program Program }

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument Program to be not nil nor empty.
func NewLowerer(p Program) Lowerer {
	return Lowerer{program: p}
}

// Triggers the lowering process. It iterates operation by operation and recursively
// calls the specified helper function based on the operation type (much like a recursive
// descend parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lowerer() (asm.Program, error) {
	program := []asm.Instruction{}

	if l.program == nil || len(l.program) == 0 {
		return nil, fmt.Errorf("the given 'program' is empty")
	}

	for _, module := range l.program {
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

			default: // Error case, unrecognized operation type
				return nil, fmt.Errorf("unrecognized operation '%T'", tOp)
			}
		}
	}

	return program, nil
}

// Specialized function to convert a 'vm.MemoryOp' node to a list of 'asm.Instruction'.
func (Lowerer) HandleMemoryOp(op MemoryOp) ([]asm.Instruction, error) {
	translated := []asm.Instruction{} // Accumulator of the translated instructions

	if op.Segment == Constant {
		translated = append(translated,
			// Takes the raw location, saves A on D
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			asm.CInstruction{Dest: "D", Comp: "A", Jump: ""},
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
			asm.CInstruction{Dest: "D", Comp: "M", Jump: ""},
			// Adds the offset and goto to the pointed location
			asm.AInstruction{Location: fmt.Sprint(op.Offset)},
			asm.CInstruction{Dest: "A", Comp: "D+A", Jump: ""},
			// Saves on D on for usage by the next instruction
			asm.CInstruction{Dest: "D", Comp: "M", Jump: ""},
		)
	}

	// TODO(hmny): Missing handling of Pop operations
	// TODO(hmny): Missing handling of 'pointer' and 'temp' segments

	// ! On D reg I find the required value to put on stack
	translated = append(translated,
		// Takes SP and goto it location,
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "A", Comp: "M", Jump: ""},
		// Saves on M the D result
		asm.CInstruction{Dest: "M", Comp: "D", Jump: ""},
		// Increments SP to new memory location
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "M", Comp: "M+1", Jump: ""},
	)

	return translated, nil
}

// Specialized function to convert a 'vm.ArithmeticOp' node to a list of 'asm.Instruction'.
func (Lowerer) HandleArithmeticOp(op ArithmeticOp) ([]asm.Instruction, error) {
	if op.Operation == Add {
		return []asm.Instruction{
			// Takes SP and goto it location,
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "AM", Comp: "M-1", Jump: ""},
			// Saves on D the M register the first operand
			asm.CInstruction{Dest: "D", Comp: "M", Jump: ""},
			// Go back one, M contains the second operand
			asm.CInstruction{Dest: "A", Comp: "A-1", Jump: ""}, // TODO
			// Do the arithmetic operation
			asm.CInstruction{Dest: "M", Comp: "D+M", Jump: ""}, // TODO
			// No need to decrement the stack pointer anymore
			// ? asm.AInstruction{Location: "SP"},
			// ? asm.CInstruction{Dest: "M", Comp: "M-1", Jump: ""},
		}, nil
	}

	return nil, fmt.Errorf("not implemented fully")
}
