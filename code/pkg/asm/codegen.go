package asm

import (
	"errors"
	"fmt"

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
//
// TODO(hmny): Add comment to document behavior
func (CodeGenerator) GenerateAInst(stmt AInstruction) (string, error) {
	if stmt.Location == "" {
		return "", errors.New("unable ro produce empty label declaration")
	}

	return fmt.Sprintf("@%s", stmt.Location), nil
}

// Specialized function to convert a C Instruction to the Asm format.
//
// TODO(hmny): Add comment to document behavior
func (cg *CodeGenerator) GenerateCInst(stmt CInstruction) (string, error) {
	if stmt.Comp == "" {
		return "", errors.New("expected 'comp' directive in C Instruction")
	}

	if stmt.Dest != "" && stmt.Jump == "" {
		return fmt.Sprintf("%s=%s", stmt.Dest, stmt.Comp), nil
	}
	if stmt.Jump != "" && stmt.Dest == "" {
		return fmt.Sprintf("%s;%s", stmt.Comp, stmt.Jump), nil
	}

	// TODO(hmny): Missing check on the well formed-ness of Comp, Dest and Jump

	return "", errors.New("expected either 'dest' or 'jump' directive in C Instruction")
}

// Specialized function to convert an Label Declaration to the Asm format.
//
// TODO(hmny): Add comment to document behavior
func (cg *CodeGenerator) GenerateLabelDecl(stmt LabelDecl) (string, error) {
	if _, found := hack.BuiltInTable[stmt.Name]; found {
		return "", fmt.Errorf("unable to override built-in label '%s'", stmt.Name)
	}

	// TODO(hmny): Missing check on the well formed-ness of the label name
	return fmt.Sprintf("(%s)", stmt.Name), nil
}
