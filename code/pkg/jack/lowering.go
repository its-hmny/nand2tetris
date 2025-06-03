package jack

import (
	"fmt"

	"its-hmny.dev/nand2tetris/pkg/vm"
)

// ----------------------------------------------------------------------------
// Jack Lowerer

// The Lowerer takes a 'jack.Program' and produces its 'vm.Program' counterpart.
//
// Since we get a tree-like struct we are able to traverse it using a Depth First Search (DFS) algorithm
// on it. For each operation node visited we produce a list of 'wm.Operation' as counterpart as well as
// validating the input before proceeding with the processing.
type Lowerer struct {
	program Program
}

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument Program to be not nil nor empty.
func NewLowerer(p Program) Lowerer {
	return Lowerer{program: p}
}

// Triggers the lowering process. It iterates class by class and then statement by statement
// and recursively calling the necessary helper function based on the construct type (much like
// a recursive descent parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lowerer() (vm.Program, error) {
	program := vm.Program{}

	if l.program == nil || len(l.program) == 0 {
		return nil, fmt.Errorf("the given 'program' is empty")
	}

	for name, class := range l.program {
		operations := []vm.Operation{}

		for _, field := range class.Fields {
			ops, err := l.HandleFieldDecl(field)
			if err != nil {
				return nil, fmt.Errorf("error handling field declaration in class '%s': %w", name, err)
			}
			operations = append(operations, ops...)
		}

		for _, subroutine := range class.Subroutines {
			ops, err := l.HandleSubroutine(subroutine)
			if err != nil {
				return nil, fmt.Errorf("error handling subroutine '%s' in class '%s': %w", subroutine.Name, name, err)
			}
			operations = append(operations, ops...)
		}

		program[name] = vm.Module(operations)
	}

	return program, nil
}

// Specialized function to convert a 'jack.FieldDecl' node to a list of 'vm.Operation'.
func (l *Lowerer) HandleFieldDecl(field Variable) ([]vm.Operation, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// Specialized function to convert a 'jack.Subroutine' node to a list of 'vm.Operation'.
func (l *Lowerer) HandleSubroutine(routine Subroutine) ([]vm.Operation, error) {
	localVars := map[string]bool{}
	for _, stmt := range routine.Statements {
		// For multiple var declarations we register each new variable in the map/set
		if varStmt, isVarStmt := stmt.(VarStmt); isVarStmt {
			for _, variable := range varStmt.Vars {
				localVars[variable.Name] = true
			}
		}
		// For let declarations we register the new variable in the map/set
		if letStmt, isLetStmt := stmt.(LetStmt); isLetStmt {
			// ! Here we support only VarExpr because ArrayExpr would mean that the array has been
			// ! already declared either in a previous VarStmt or LetStmt, making it redundant.
			localVars[letStmt.Lhs.(VarExpr).Var] = true
		}
	}

	fDecl, fBody := vm.FuncDecl{Name: routine.Name, NLocal: uint8(len(localVars))}, []vm.Operation{}

	for _, stmt := range routine.Statements {
	}

	return append([]vm.Operation{fDecl}, fBody...), nil
}

	return append([]vm.Operation{fDecl}, fBody...), nil
}

