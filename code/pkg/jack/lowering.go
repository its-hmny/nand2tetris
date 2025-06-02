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
		for _, field := range class.Fields {
			fmt.Printf("Processing field '%s' for class '%s'\n", field.Name, name)
		}

		for _, subroutine := range class.Subroutines {
			fmt.Printf("Processing subroutine '%s' for class '%s'\n", subroutine.Name, name)
		}
	}

	return program, nil
}
