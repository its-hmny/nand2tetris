package asm

import (
	"fmt"
	"strconv"

	"its-hmny.dev/nand2tetris/pkg/hack"
)

// ----------------------------------------------------------------------------
// Code Generator

// Takes some a set of 'asm.Statement' and spits out their textual counterparts.
//
// The translation can be done without any additional data structure but the program.
type CodeGenerator struct {
	program []Statement // The set of statements to convert in Hack binary format
}

// Initializes and returns to the caller a brand new 'CodeGenerator' struct.
// Requires that argument Program 'p' (what we want to translate) is non-nil.
func NewCodeGenerator(p []Statement) CodeGenerator {
	return CodeGenerator{program: p}
}

// Translate each statement in the 'program' field to the Asm textual format.
//
// Each instruction will pass through the following step: evaluation, validation and
// then conversion to its textual representation (a string) so that it can be further
// elaborated by the caller (e.g. dumping to a file, runtime interpretation, ...).
func (cg *CodeGenerator) Generate() ([]string, error) {
	asm := make([]string, 0, len(cg.program))

	for _, statement := range cg.program {
		var generated string = ""
		var err error = nil

		switch tStatement := statement.(type) {
		case AInstruction:
			generated, err = cg.GenerateAInst(tStatement)
		case CInstruction:
			generated, err = cg.GenerateCInst(tStatement)
		case LabelDecl:
			generated, err = cg.GenerateLabelDecl(tStatement)
		}

		if err != nil {
			return nil, err
		}
		asm = append(asm, generated)
	}

	return asm, nil
}

// Specialized function to convert an A Instruction to the Asm format.
func (CodeGenerator) GenerateAInst(stmt AInstruction) (string, error) {
	// Pre-check on the label/built-in/raw label (is required not to be empty)
	if stmt.Location == "" {
		return "", fmt.Errorf("unable to produce empty label declaration")
	}
	// Pre-check on the raw literal access (has to be an addressable memory location)
	addr, err := strconv.ParseUint(stmt.Location, 10, 16)
	if err == nil && uint16(addr) >= hack.MaxAddressableMemory {
		return "", fmt.Errorf("unable to use out-of-bound raw address")
	}

	return fmt.Sprintf("@%s", stmt.Location), nil
}

// Specialized function to convert a C Instruction to the Asm format.
func (CodeGenerator) GenerateCInst(stmt CInstruction) (string, error) {
	// Pre-check on the 'comp' (required), 'dest' and 'jump' (either one or the other)
	if _, found := hack.CompTable[stmt.Comp]; stmt.Comp == "" || !found {
		return "", fmt.Errorf("expected valid 'comp' directive in CInst, got: '%s'", stmt.Comp)
	}
	if stmt.Jump != "" && stmt.Dest != "" {
		return "", fmt.Errorf("expected either 'dest' or 'jump' directive in CInst")
	}

	// The instruction has either a valid 'jump' or valid 'dest' directive
	if _, found := hack.DestTable[stmt.Dest]; stmt.Dest != "" && found {
		return fmt.Sprintf("%s=%s", stmt.Dest, stmt.Comp), nil
	}
	if _, found := hack.JumpTable[stmt.Jump]; stmt.Jump != "" && found {
		return fmt.Sprintf("%s;%s", stmt.Comp, stmt.Jump), nil
	}

	return "", fmt.Errorf("neither 'dest' or 'jump' directives are valid in C Inst")
}

// Specialized function to convert an Label Declaration to the Asm format.
func (CodeGenerator) GenerateLabelDecl(stmt LabelDecl) (string, error) {
	if stmt.Name == "" {
		return "", fmt.Errorf("unable to declare empty label")
	}
	if _, found := hack.BuiltInTable[stmt.Name]; found {
		return "", fmt.Errorf("unable to override built-in label '%s'", stmt.Name)
	}
	return fmt.Sprintf("(%s)", stmt.Name), nil
}
