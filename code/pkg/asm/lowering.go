package asm

import (
	"fmt"
	"strconv"

	pc "github.com/prataprc/goparsec"
	"its-hmny.dev/nand2tetris/pkg/hack"
)

// ----------------------------------------------------------------------------
// Asm Lowerer

// The Lowerer takes an Abstract Syntax Tree (AST) and produces its 'hack.Program' counterpart.
//
// Since we get a tree we are able to traverse it using a simple Depth First Search (DFS) algorithm
// on it. For each instruction node visited we produce it's 'hack.Instruction' counterpart (either
// A Instruction or C Instruction) as well as validating the input before proceeding.
type Lowerer struct{ root pc.Queryable }

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument pc.Queryable to be not nil.
func NewLowerer(r pc.Queryable) Lowerer {
	return Lowerer{root: r}
}

// Triggers the lowering process on the given AST root. It iterates on the top-level children
// of the AST and recursively calls the specified helper function based on the child type (much
// like a recursive descend parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lower() (hack.Program, hack.SymbolTable, error) {
	program, table := []hack.Instruction{}, map[string]uint16{}

	if l.root.GetName() != "program" {
		return nil, nil, fmt.Errorf("expected node 'program', found %s", l.root.GetName())
	}

	for _, child := range l.root.GetChildren() {
		switch child.GetName() {
		// Traverse the AST subtree and returns the in-memory A Instruction defined inside.
		// After that, adds the instruction to the Program for the 'codegen' phase.
		case "a-inst":
			inst, err := l.HandleAInst(child)
			if inst == nil || err != nil {
				return nil, nil, err
			}
			program = append(program, inst)

		// Traverse the AST subtree and returns the in-memory C Instruction defined inside.
		// After that, adds the instruction to the Program for the 'codegen' phase.
		case "c-inst":
			inst, err := l.HandleCInst(child)
			if inst == nil || err != nil {
				return nil, nil, err
			}
			program = append(program, inst)

		// Traverse the AST subtree and returns the label declared in each that subtree.
		// After that, adds a new (symbol,address) tuple to the SymbolTable for the 'codegen' phase.
		case "label-decl":
			label, err := l.HandleLabelDecl(child)
			if label == "" || err != nil {
				return nil, nil, err
			}
			table[label] = uint16(len(program))

		// Comment nodes in the AST are just skipped since not required for 'codegen' phase.
		case "comment":
			continue

		// Error case, unrecognized top-level node in the AST
		default:
			return nil, nil, fmt.Errorf("unrecognized node '%s'", child.GetName())
		}
	}

	return program, table, nil
}

// Specialized function to convert a "a-inst" node to an 'hack.AInstruction'.
func (Lowerer) HandleAInst(inst pc.Queryable) (hack.Instruction, error) {
	if inst.GetName() != "a-inst" { // Prelude checks: inspects the node to verify it's an 'a-inst'
		return nil, fmt.Errorf("expected node 'a-inst', found %s", inst.GetName())
	}

	symbol := inst.GetChildren()[1] // Prelude checks: inspects the label node type (INT | SYMBOL)
	if symbol.GetName() != "INT" && symbol.GetName() != "SYMBOL" {
		return nil, fmt.Errorf("expected token 'SYMBOL' or 'INT', got %s", symbol.GetName())
	}

	// Based on one of the following cases below (the type of the symbol) we do different things:
	// 1) If it's present in the BuiltInTable is we set the 'LocType'to 'BuiltIn' accordingly
	if _, found := hack.BuiltInTable[symbol.GetValue()]; found {
		return hack.AInstruction{LocType: hack.BuiltIn, LocName: symbol.GetValue()}, nil
	}
	// 2) If it can be parsed as an int we set the 'LocType' to 'Raw' accordingly
	if _, err := strconv.ParseInt(symbol.GetValue(), 10, 16); err == nil {
		return hack.AInstruction{LocType: hack.Raw, LocName: symbol.GetValue()}, nil
	}
	// 3) Else it's a user defined label and we set 'LocType' to 'Label' accordingly
	return hack.AInstruction{LocType: hack.Label, LocName: symbol.GetValue()}, nil
}

// Specialized function to convert a "c-inst" node to an 'hack.CInstruction'.
func (Lowerer) HandleCInst(inst pc.Queryable) (hack.Instruction, error) {
	if inst.GetName() != "c-inst" { // Prelude checks: inspects the node to verify it's an 'a-inst'
		return nil, fmt.Errorf("expected node 'c-inst', found %s", inst.GetName())
	}

	dest, comp, jump := inst.GetChildren()[0], inst.GetChildren()[1], inst.GetChildren()[2]

	if dest.GetName() == "assign" && len(dest.GetChildren()) == 2 {
		dest = dest.GetChildren()[0]
		return hack.CInstruction{Dest: dest.GetValue(), Comp: comp.GetValue()}, nil
	}

	if jump.GetName() == "goto" || len(dest.GetChildren()) == 2 {
		jump = jump.GetChildren()[1]
		return hack.CInstruction{Comp: comp.GetValue(), Jump: jump.GetValue()}, nil
	}

	return nil, fmt.Errorf("expected either node 'assign' or 'goto' not found")
}

// Specialized function to extract from a "label-decl" node to the identifier of the label.
func (Lowerer) HandleLabelDecl(decl pc.Queryable) (string, error) {
	if decl.GetName() != "label-decl" { // Prelude checks: inspects the node to verify it's a 'label-decl'
		return "", fmt.Errorf("expected node 'a-inst', found %s", decl.GetName())
	}

	symbol := decl.GetChildren()[1] // Prelude checks: inspects the label node type (INT | SYMBOL)
	if symbol.GetName() != "SYMBOL" {
		return "", fmt.Errorf("expected token 'SYMBOL', got %s", symbol.GetName())
	}

	return symbol.GetValue(), nil
}
