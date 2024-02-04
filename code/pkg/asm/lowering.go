package asm

import (
	"fmt"
	"strconv"

	"its-hmny.dev/nand2tetris/pkg/hack"
)

// ----------------------------------------------------------------------------
// Asm Lowerer

// The Lowerer takes an 'asm.Program' and produces its 'hack.Program' counterpart.
//
// Since we get a tree-like struct we are able to traverse it using a Depth First Search (DFS) algorithm
// on it. For each instruction node visited we produce it's 'hack.Instruction' counterpart (either
// A Instruction or C Instruction) as well as validating the input before proceeding.
type Lowerer struct{ program Program }

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument Program to be not nil nor empty.
func NewLowerer(p Program) Lowerer {
	return Lowerer{program: p}
}

// Triggers the lowering process. It iterates instruction by instruction and recursively
// calls the specified helper function based on the instruction type (much like a recursive
// descend parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lower() (hack.Program, hack.SymbolTable, error) {
	converted, table := []hack.Instruction{}, map[string]uint16{}

	if l.program == nil || len(l.program) == 0 {
		return nil, nil, fmt.Errorf("the given 'program' is empty")
	}

	for _, asmInst := range l.program {
		switch tAsmInst := asmInst.(type) {
		case AInstruction: // Converts 'asm.AInstruction' to 'hack.AInstruction'
			hackInst, err := l.HandleAInst(tAsmInst)
			if hackInst == nil || err != nil {
				return nil, nil, err
			}
			converted = append(converted, hackInst)

		case CInstruction: // Converts 'asm.CInstruction' to 'hack.CInstruction'
			hackInst, err := l.HandleCInst(tAsmInst)
			if hackInst == nil || err != nil {
				return nil, nil, err
			}
			converted = append(converted, hackInst)

		case LabelDecl: // Adds 'asm.LabelDecl' to the 'hack.SymbolTable'
			label, err := l.HandleLabelDecl(tAsmInst)
			if label == "" || err != nil {
				return nil, nil, err
			}
			table[label] = uint16(len(converted))

		default: // Error case, unrecognized operation type
			return nil, nil, fmt.Errorf("unrecognized instruction '%T'", asmInst)
		}
	}

	return converted, table, nil
}

// Specialized function to convert a 'asm.AInstruction' node to an 'hack.AInstruction'.
func (Lowerer) HandleAInst(inst AInstruction) (hack.Instruction, error) {
	// Based on one of the following cases below (the type of the symbol) we do different things:
	// 1) If it's present in the BuiltInTable is we set the 'LocType'to 'BuiltIn' accordingly
	if _, found := hack.BuiltInTable[inst.Location]; found {
		return hack.AInstruction{LocType: hack.BuiltIn, LocName: inst.Location}, nil
	}
	// 2) If it can be parsed as an int we set the 'LocType' to 'Raw' accordingly
	if _, err := strconv.ParseInt(inst.Location, 10, 16); err == nil {
		return hack.AInstruction{LocType: hack.Raw, LocName: inst.Location}, nil
	}
	// 3) Else it's a user defined label and we set 'LocType' to 'Label' accordingly
	return hack.AInstruction{LocType: hack.Label, LocName: inst.Location}, nil
}

// Specialized function to convert a 'asm.CInstruction' node to an 'hack.CInstruction'.
func (Lowerer) HandleCInst(inst CInstruction) (hack.Instruction, error) {
	if inst.Comp == "" { // Pre-check: CInstruction.Comp should always be provided
		return nil, fmt.Errorf("'Comp' sub-instruction should always be provided")
	}

	if inst.Dest != "" && inst.Jump == "" {
		return hack.CInstruction{Dest: inst.Dest, Comp: inst.Comp}, nil
	}
	if inst.Jump != "" && inst.Dest == "" {
		return hack.CInstruction{Comp: inst.Comp, Jump: inst.Jump}, nil
	}

	return nil, fmt.Errorf("expected either node 'Dest' or 'Jump' sub-instructions")
}

// Specialized function to extract from a 'asm.LabelDecl' node to the identifier of the label.
func (Lowerer) HandleLabelDecl(inst LabelDecl) (string, error) {
	return inst.Name, nil
}
