package jack

import (
	"fmt"
	"strings"
)

type TypeChecker struct {
	program Program
	scopes  ScopeTable // Keeps track of the scopes and declared variables inside each one
}

func NewTypeChecker(program Program) TypeChecker {
	return TypeChecker{program: program}
}

func (tc *TypeChecker) Check() (bool, error) {
	if tc.program == nil {
		return false, fmt.Errorf("the given 'program' is empty or nil")
	}

	for name, class := range tc.program {
		_, err := tc.HandleClass(class)
		if err != nil {
			return false, fmt.Errorf("error handling lowering of class '%s': %w", name, err)
		}

	}

	return true, nil
}

// Specialized function to type-check a 'jack.Class' and nested fields.
func (tc *TypeChecker) HandleClass(class Class) (bool, error) {
	tc.scopes.PushClassScope(class.Name) // Keep track of the current scope being processed
	defer tc.scopes.PopClassScope()      // Reset the function name after processing

	for _, field := range class.Fields.Entries() {
		_, err := tc.HandleVarStmt(VarStmt{Vars: []Variable{field}})
		if err != nil {
			return false, fmt.Errorf("error handling field '%s' in class '%s': %w", field.Name, class.Name, err)
		}
	}

	for _, subroutine := range class.Subroutines.Entries() {
		_, err := tc.HandleSubroutine(subroutine)
		if err != nil {
			return false, fmt.Errorf("error handling subroutine '%s' in class '%s': %w", subroutine.Name, class.Name, err)
		}
	}

	return true, nil
}

// Specialized function to type-check a 'jack.Subroutine' and nested fields.
func (tc *TypeChecker) HandleSubroutine(subroutine Subroutine) (any, any) {
	tc.scopes.PushSubRoutineScope(subroutine.Name) // Keep track of the current subroutine function being processed
	defer tc.scopes.PopSubroutineScope()           // Reset the function name after processing

	// We add to the current scope also all of the arguments of the subroutine
	for _, arg := range subroutine.Arguments.Entries() {
		// Like this we're actually supporting shadowing of variables, so if a variable
		// with the same name is already present in the current scope, we just temporarily
		// override it with the most update one instead of returning an error (like Go does
		tc.scopes.RegisterVariable(arg)
	}

	for _, stmt := range subroutine.Statements {
		_, err := tc.HandleStatement(stmt)
		if err != nil {
			return false, fmt.Errorf("error handling nested statement %T': %w", stmt, err)
		}
	}

	return false, fmt.Errorf("not implemented yet")
}

// Generalized function to type-check multiple statements types.
func (tc *TypeChecker) HandleStatement(stmt Statement) (bool, error) {
	return false, fmt.Errorf("not implemented yet")
}
