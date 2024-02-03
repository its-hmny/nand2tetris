package asm

import (
	"fmt"
	"strconv"

	"its-hmny.dev/nand2tetris/pkg/hack"
)

// ----------------------------------------------------------------------------
// Code Generator

// Takes some a set of 'asm.Instruction' and spits out their textual counterparts.
//
// The translation can be done without any additional data structure but the program.
type CodeGenerator struct {
	program Program // The set of statements to convert in asm textual format
}

// Initializes and returns to the caller a brand new 'CodeGenerator' struct.
// Requires that argument Program 'p' (what we want to translate) is non-nil.
func NewCodeGenerator(p Program) CodeGenerator {
	return CodeGenerator{program: p}
}

// Translate each statement in the 'program' field to the Asm textual format.
//
// Each instruction will pass through the following step: evaluation, validation and
// then conversion to its textual representation (a string) so that it can be further
// elaborated by the caller (e.g. dumping to a file, runtime interpretation, ...).
func (cg *CodeGenerator) Generate() ([]string, error) {
	asm := make([]string, 0, len(cg.program))

	for _, instruction := range cg.program {
		var generated string = ""
		var err error = nil

		switch tInstruction := instruction.(type) {
		case AInstruction:
			generated, err = cg.GenerateAInst(tInstruction)
		case CInstruction:
			generated, err = cg.GenerateCInst(tInstruction)
		case LabelDecl:
			generated, err = cg.GenerateLabelDecl(tInstruction)
		}

		if err != nil {
			return nil, err
		}
		asm = append(asm, generated)
	}

	return asm, nil
}

// Specialized function to convert an A Instruction to the Asm format.
func (CodeGenerator) GenerateAInst(inst AInstruction) (string, error) {
	// Pre-check on the label/built-in/raw label (is required not to be empty)
	if inst.Location == "" {
		return "", fmt.Errorf("unable to produce empty label declaration")
	}
	// Pre-check on the raw literal access (has to be an addressable memory location)
	addr, err := strconv.ParseUint(inst.Location, 10, 16)
	if err == nil && uint16(addr) >= hack.MaxAddressableMemory {
		return "", fmt.Errorf("unable to use out-of-bound raw address")
	}

	return fmt.Sprintf("@%s", inst.Location), nil
}

// Specialized function to convert a C Instruction to the Asm format.
func (CodeGenerator) GenerateCInst(inst CInstruction) (string, error) {
	// Pre-check on the 'comp' (required), 'dest' and 'jump' (either one or the other)
	if _, found := hack.CompTable[inst.Comp]; inst.Comp == "" || !found {
		return "", fmt.Errorf("expected valid 'comp' directive in CInst, got: '%s'", inst.Comp)
	}
	if inst.Jump != "" && inst.Dest != "" {
		return "", fmt.Errorf("expected either 'dest' or 'jump' directive in CInst")
	}

	// The instruction has either a valid 'jump' or valid 'dest' directive
	if _, found := hack.DestTable[inst.Dest]; inst.Dest != "" && found {
		return fmt.Sprintf("%s=%s", inst.Dest, inst.Comp), nil
	}
	if _, found := hack.JumpTable[inst.Jump]; inst.Jump != "" && found {
		return fmt.Sprintf("%s;%s", inst.Comp, inst.Jump), nil
	}

	return "", fmt.Errorf("neither 'dest' or 'jump' directives are valid in C Inst")
}

// Specialized function to convert an Label Declaration to the Asm format.
func (CodeGenerator) GenerateLabelDecl(inst LabelDecl) (string, error) {
	if inst.Name == "" {
		return "", fmt.Errorf("unable to declare empty label")
	}
	if _, found := hack.BuiltInTable[inst.Name]; found {
		return "", fmt.Errorf("unable to override built-in label '%s'", inst.Name)
	}
	return fmt.Sprintf("(%s)", inst.Name), nil
}
