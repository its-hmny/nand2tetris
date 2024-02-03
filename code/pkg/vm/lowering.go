package vm

import (
	"fmt"
	"log"
	"strconv"

	pc "github.com/prataprc/goparsec"
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

// The Lowerer takes an Abstract Syntax Tree (AST) and produces its 'asm.Program' counterpart.
//
// Since we get a tree we are able to traverse it using a simple Depth First Search (DFS) algorithm
// on it. For each instruction node visited we produce it's 'asm.Instruction' counterpart (either
// A Instruction, C Instruction) or Label Declaration as well as validating the input before proceeding.
type Lowerer struct{ root pc.Queryable }

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument pc.Queryable to be not nil.
func NewLowerer(r pc.Queryable) Lowerer {
	return Lowerer{root: r}
}

// Triggers the lowering process on the given AST root. It iterates on the top-level children
// of the AST and recursively calls the specified helper function based on the child type (much
// like a recursive descend parser but for lowering), this means the AST is visited in DFS order.
func (hl *Lowerer) Lowerer() (asm.Program, error) {
	program := []asm.Instruction{}

	if hl.root.GetName() != "module" {
		return nil, fmt.Errorf("expected node 'program', found %s", hl.root.GetName())
	}

	for _, child := range hl.root.GetChildren() {
		switch child.GetName() {
		// Traverse the AST subtree and extracts the MemoryOp struct defined inside.
		case "memory_op":
			inst, err := hl.handleMemoryOp(child)
			if inst == nil || err != nil {
				return nil, err
			}
			program = append(program, inst...)

		case "arithmetic_op":
			inst, err := hl.handleArithmeticOp(child)
			if inst == nil || err != nil {
				return nil, err
			}
			program = append(program, inst...)

		// Comment nodes in the AST are just skipped since not required for 'codegen' phase.
		case "comment":
			continue

		// Error case, unrecognized top-level node in the AST
		default:
			return nil, fmt.Errorf("unrecognized node '%s'", child.GetName())
		}
	}

	return program, nil
}

// Specialized function to convert a "memory_op" node to a list of 'asm.Instruction'.
func (Lowerer) handleMemoryOp(node pc.Queryable) ([]asm.Instruction, error) {
	if node.GetName() != "memory_op" {
		return nil, fmt.Errorf("expected node 'memory_op', got %s", node.GetName())
	}

	children := node.GetChildren()
	if len(children) != 3 {
		return nil, fmt.Errorf("expected node with 3 leaf, got %d", len(children))
	}

	operation := OperationType(children[0].GetValue())
	segment := SegmentType(children[1].GetValue())
	offset, err := strconv.ParseUint(children[2].GetValue(), 10, 16)
	if err != nil {
		log.Fatalf("failed to parse 'offset' in MemoryOp, got '%s'", children[2].GetValue())
	}

	if operation == Push {
		return []asm.Instruction{
			LocationResolver[segment](uint16(offset)),
			asm.CInstruction{Dest: "D", Comp: "M"},
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "M", Comp: "D"},
			asm.CInstruction{Dest: "A", Comp: "A+1"},
		}, nil
	}

	if operation == Pop {
		return []asm.Instruction{
			asm.AInstruction{Location: "SP"},
			asm.CInstruction{Dest: "D", Comp: "M"},
			LocationResolver[segment](uint16(offset)),
			asm.CInstruction{Dest: "M", Comp: "D"},
			asm.CInstruction{Dest: "A", Comp: "A-1"},
		}, nil
	}

	return nil, fmt.Errorf("unrecognized OperationType '%s'", operation)
}

// Specialized function to convert a "arithmetic_op" node to a list of 'asm.Instruction'.
func (Lowerer) handleArithmeticOp(node pc.Queryable) ([]asm.Instruction, error) {
	if node.GetName() != "arithmetic_op" {
		log.Fatalf("expected node 'arithmetic_op', got %s ", node.GetName())
	}
	children := node.GetChildren()
	if len(children) != 1 {
		log.Fatalf("expected node 'arithmetic_op' with 1 children, got %d", len(children))
	}

	operation := ArithOpType(children[0].GetValue())

	return []asm.Instruction{
		asm.AInstruction{Location: "SP"},
		asm.CInstruction{Dest: "D", Comp: "M"},
		asm.CInstruction{Dest: "A", Comp: "A-1"},
		IntrinsicResolver[operation](),
		asm.CInstruction{Dest: "M", Comp: "D"},
	}, nil
}
