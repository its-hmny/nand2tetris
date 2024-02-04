package vm

import (
	"fmt"

	"its-hmny.dev/nand2tetris/pkg/asm"
)

var LocationResolver = map[SegmentType]func(uint16) asm.AInstruction{
	Constant: func(constant uint16) asm.AInstruction {
		return asm.AInstruction{Location: fmt.Sprint(constant)}
	},
}

var IntrinsicResolver = map[ArithOpType]func() asm.CInstruction{
	Add: func() asm.CInstruction {
		return asm.CInstruction{Dest: "D", Comp: "D+M"}
	},
}

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
	return nil, fmt.Errorf("not implemented yet")
}

// Specialized function to convert a 'vm.ArithmeticOp' node to a list of 'asm.Instruction'.
func (Lowerer) HandleArithmeticOp(op ArithmeticOp) ([]asm.Instruction, error) {
	return nil, fmt.Errorf("not implemented yet")
}
