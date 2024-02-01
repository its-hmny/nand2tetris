package vm

import (
	"fmt"
	"log"
	"strconv"

	pc "github.com/prataprc/goparsec"
)

type ASTLowerer struct{}

func NewVMLowerer() ASTLowerer { return ASTLowerer{} }

func (hl *ASTLowerer) FromAST(root pc.Queryable) (Module, error) {
	module := Module{Statements: []Statement{}}

	if root.GetName() != "module" {
		return Module{}, fmt.Errorf("expected node 'program', found %s", root.GetName())
	}

	for _, child := range root.GetChildren() {
		switch child.GetName() {
		// Traverse the AST subtree and extracts the MemoryOp struct defined inside.
		case "memory_op":
			inst, err := hl.handleMemoryOp(child)
			if inst == nil || err != nil {
				return Module{}, err
			}
			module.Statements = append(module.Statements, inst)

		case "arithmetic_op":
			inst, err := hl.handleArithmeticOp(child)
			if inst == nil || err != nil {
				return Module{}, err
			}
			module.Statements = append(module.Statements, inst)

		// ? // Traverse the AST subtree and returns the in-memory A Instruction defined inside.
		// ? // After that, adds the instruction to the Program for the 'codegen' phase.
		// ? case "a-inst":
		// ? 	inst, err := hl.HandleAInst(child)
		// ? 	if inst == nil || err != nil {
		// ? 		return nil, err
		// ? 	}
		// ? 	program = append(program, inst)

		// Comment nodes in the AST are just skipped since not required for 'codegen' phase.
		case "comment":
			continue

		// Error case, unrecognized top-level node in the AST
		default:
			return Module{}, fmt.Errorf("unrecognized node '%s'", child.GetName())
		}
	}

	return module, nil
}

func (ASTLowerer) handleMemoryOp(node pc.Queryable) (Statement, error) {
	if node.GetName() != "memory_op" {
		return nil, fmt.Errorf("expected node 'memory_op', got %s %d", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		return nil, fmt.Errorf("expected node with 3 leaf, got %d", len(node.GetChildren()))
	}

	memoptype, segment, index := node.GetChildren()[0], node.GetChildren()[1], node.GetChildren()[2]

	offset, err := strconv.ParseUint(index.GetValue(), 10, 16)
	if err != nil {
		log.Fatalf("failed to parse 'offset' in MemoryOp, got '%s'", index.GetValue())
	}

	return MemoryOp{
		Offset:    uint16(offset),
		Segment:   SegmentType(segment.GetValue()),
		Operation: OperationType(memoptype.GetValue()),
	}, nil
}

func (ASTLowerer) handleArithmeticOp(node pc.Queryable) (Statement, error) {
	if node.GetName() != "arithmetic_op" {
		log.Fatalf("expected node 'arithmetic_op', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 1 {
		log.Fatalf("expected node 'arithmetic_op' with 1 children, got %d", len(node.GetChildren()))
	}

	operand := node.GetChildren()[0]
	return ArithmeticOp{Operation: ArithOpType(operand.GetValue())}, nil
}
