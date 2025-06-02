package vm

import (
	"fmt"
)

// ----------------------------------------------------------------------------
// Code Generator

// Takes a 'vm.Program' and spits out its source code counterparts.
//
// The translation can be done without any additional data structure but the program.
type CodeGenerator struct {
	program Program // The set of modules to convert in VM code format
}

// Initializes and returns to the caller a brand new 'CodeGenerator' struct.
// Requires that argument Program 'p' (what we want to translate) is non -nil.
func NewCodeGenerator(p Program) CodeGenerator {
	return CodeGenerator{program: p}
}

// Translates each instruction in the 'program' to the VM string format.
//
// Each instruction will pass through the following step: evaluation, validation and then conversion
// to its string representation so that it can be further elaborated by the function caller
// (e.g. dumping .hack code to a file, runtime interpretation, ...).
func (cg *CodeGenerator) Generate() (map[string][]string, error) {
	vm := map[string][]string{}

	for modName, module := range cg.program {
		for _, operation := range module {
			var generated string = ""
			var err error = nil

			switch tOperation := operation.(type) {
			case MemoryOp:
				generated, err = cg.GenerateMemoryOp(tOperation)
			case ArithmeticOp:
				generated, err = cg.GenerateArithmeticOp(tOperation)
			case LabelDecl:
				generated, err = cg.GenerateLabelDecl(tOperation)
			case GotoOp:
				generated, err = cg.GenerateGotoOp(tOperation)
			case FuncDecl:
				generated, err = cg.GenerateFuncDecl(tOperation)
			case ReturnOp:
				generated, err = cg.GenerateReturnOp(tOperation)
			case FuncCallOp:
				generated, err = cg.GenerateFuncCallOp(tOperation)

			}

			if err != nil {
				return nil, err
			}
			vm[modName] = append(vm[modName], generated)
		}
	}

	return vm, nil
}

// Specialized function to convert a 'MemoryOp' operation to the VM format.
func (cg *CodeGenerator) GenerateMemoryOp(op MemoryOp) (string, error) {
	// Bound checking on segment that do have an upperbound to the allowed offsets
	if op.Segment == Pointer && op.Offset > 1 {
		return "", fmt.Errorf("invalid 'pointer' offset, got %d", op.Offset)
	}
	if op.Segment == Temp && op.Offset > 7 {
		return "", fmt.Errorf("invalid 'temp' offset, got %d", op.Offset)
	}

	return fmt.Sprintf("%s %s %d", string(op.Operation), string(op.Segment), op.Offset), nil
}

// Specialized function to convert a 'ArithmeticOp' operation to the VM format.
func (cg *CodeGenerator) GenerateArithmeticOp(op ArithmeticOp) (string, error) {
	return string(op.Operation), nil
}

// Specialized function to convert a 'LabelDecl' operation to the VM format.
func (cg *CodeGenerator) GenerateLabelDecl(op LabelDecl) (string, error) {
	if op.Name == "" {
		return "", fmt.Errorf("unable to produce empty label declaration")
	}

	return fmt.Sprintf("label %s", op.Name), nil
}

// Specialized function to convert a 'GotoOp' operation to the VM format.
func (cg *CodeGenerator) GenerateGotoOp(op GotoOp) (string, error) {
	if op.Label == "" {
		return "", fmt.Errorf("unable to produce empty jump label")
	}

	return fmt.Sprintf("%s %s", string(op.Jump), op.Label), nil
}

// Specialized function to convert a 'FuncDecl' operation to the VM format.
func (cg *CodeGenerator) GenerateFuncDecl(op FuncDecl) (string, error) {
	if op.Name == "" {
		return "", fmt.Errorf("unable to produce empty function declaration")
	}

	return fmt.Sprintf("function %s %d", op.Name, op.NLocal), nil
}

// Specialized function to convert a 'ReturnOp' operation to the VM format.
func (cg *CodeGenerator) GenerateReturnOp(op ReturnOp) (string, error) {
	return "return", nil
}

// Specialized function to convert a 'FuncCallOp' operation to the VM format.
func (cg *CodeGenerator) GenerateFuncCallOp(op FuncCallOp) (string, error) {
	if op.Name == "" {
		return "", fmt.Errorf("unable to produce empty function call")
	}

	return fmt.Sprintf("call %s %d", op.Name, op.NArgs), nil
}
