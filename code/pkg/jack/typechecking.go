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
func (tc *TypeChecker) HandleSubroutine(subroutine Subroutine) (bool, error) {
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
	switch tStmt := stmt.(type) {
	case DoStmt:
		return tc.HandleDoStmt(tStmt)
	case VarStmt:
		return tc.HandleVarStmt(tStmt)
	case LetStmt:
		return tc.HandleLetStmt(tStmt)
	case IfStmt:
		return tc.HandleIfStmt(tStmt)
	case WhileStmt:
		return tc.HandleWhileStmt(tStmt)
	case ReturnStmt:
		return tc.HandleReturnStmt(tStmt)
	default:
		return false, fmt.Errorf("unrecognized statement: %T", stmt)
	}
}

// Specialized function to type-check a 'jack.DoStmt' and nested fields.
func (tc *TypeChecker) HandleDoStmt(statement DoStmt) (bool, error) {
	_, err := tc.HandleFuncCallExpr(statement.FuncCall)
	if err != nil {
		return false, fmt.Errorf("error handling nested function call expression: %w", err)
	}

	return true, nil // Since the return value is discarded type-checking will always succeed
}

// Specialized function to type-check a 'jack.VarStmt' and nested fields.
func (tc *TypeChecker) HandleVarStmt(statement VarStmt) (bool, error) {
	for _, variable := range statement.Vars {
		// Like this we're actually supporting shadowing of variables, so if a variable
		// with the same name is already present in the current scope, we just temporarily
		// override it with the most update one instead of returning an error (like Go does BTW).
		tc.scopes.RegisterVariable(variable)
	}
	return true, nil // No type-checking needed for variable declaration, just return true
}

// Specialized function to type-check a 'jack.LetStmt' and nested fields.
func (tc *TypeChecker) HandleLetStmt(statement LetStmt) (bool, error) {
	rhs, err := tc.HandleExpression(statement.Rhs)
	if err != nil {
		return false, fmt.Errorf("error handling RHS expression: %w", err)
	}

	// If it's a VarExpr then we somewhat reuse the same logic as HandleVarExpr, but we need to write memory instead of reading
	if expr, isVarExpr := statement.Lhs.(VarExpr); isVarExpr {
		// TODO (hmny): Should check 'rhs' against the type of var
		return false, fmt.Errorf("VarExpr not supported yet")
	}

	// For ArrayExpr instead we reuse the pointer + offset logic from HandleArrayExpr but after that we write
	// a bit of glue code to save the RHS on temporary memory before loading the new address and writing it
	if expr, isArrayExpr := statement.Lhs.(ArrayExpr); isArrayExpr {
		// TODO (hmny): Should check 'rhs' against the type of array
		return false, fmt.Errorf("ArrayExpr not supported yet")
	}

	return false, fmt.Errorf("LHS expression must be either a 'VarExpr' or an 'ArrayExpr', got: %T", statement.Lhs)
}

// Specialized function to type-check a 'jack.IfStmt' and nested fields.
func (tc *TypeChecker) HandleIfStmt(statement IfStmt) (bool, error) {
	cond, err := tc.HandleExpression(statement.Condition)
	if err != nil {
		return false, fmt.Errorf("error handling if condition expression: %w", err)
	}
	if cond != Bool {
		return false, fmt.Errorf("if expression should be boolean expression, got %s", cond)
	}

	for _, stmt := range statement.ThenBlock {
		_, err := tc.HandleStatement(stmt)
		if err != nil {
			return false, fmt.Errorf("error handling statement in 'then' block: %w", err)
		}
	}

	for _, stmt := range statement.ElseBlock {
		_, err := tc.HandleStatement(stmt)
		if err != nil {
			return false, fmt.Errorf("error handling statement in 'else' block: %w", err)
		}
	}

	return true, nil
}

// Specialized function to type-check a 'jack.WhileStmt' and nested fields.
func (tc *TypeChecker) HandleWhileStmt(statement WhileStmt) (bool, error) {
	cond, err := tc.HandleExpression(statement.Condition)
	if err != nil {
		return false, fmt.Errorf("error handling while condition expression: %w", err)
	}
	if cond != Bool {
		return false, fmt.Errorf("while expression should be boolean expression, got %s", cond)
	}

	for _, stmt := range statement.Block {
		_, err := tc.HandleStatement(stmt)
		if err != nil {
			return false, fmt.Errorf("error handling statement in while block: %w", err)
		}
	}

	return true, nil
}

// Specialized function to type-check a 'jack.ReturnStmt' and nested fields.
func (tc *TypeChecker) HandleReturnStmt(statement ReturnStmt) (bool, error) {
	className := strings.Split(tc.scopes.GetScope(), ".")[0]
	subroutineName := strings.Split(tc.scopes.GetScope(), ".")[1]

	// Retrieve the current class and current subroutine information (checking for existence)
	class, exists := tc.program[className]
	if !exists {
		return false, fmt.Errorf("class %s doesn't exists", className)
	}
	subroutine, exists := class.Subroutines.Get(subroutineName)
	if !exists {
		return false, fmt.Errorf("routine %s doesn't exists for class %s", subroutineName, className)
	}

	// No expression means just void and hence type check always pass
	if subroutine.Return == Void && statement.Expr == nil {
		return true, nil
	}
	if subroutine.Return == Void && statement.Expr != nil {
		return false, fmt.Errorf("return type of function is void but an expr has been provided")
	}

	// When the subroutine has a return type defined we need to check it against the actual return expression
	ret, err := tc.HandleExpression(statement.Expr)
	if err != nil {
		return false, fmt.Errorf("error handling return expression: %w", err)
	}
	if ret != subroutine.Return {
		return false, fmt.Errorf("expected return type %s, got %s", subroutine.Return, ret)
	}

	return true, nil
}

// Generalized function to type-check multiple expression their final 'jack.DataType'.
func (tc *TypeChecker) HandleExpression(expr Expression) (DataType, error) {
	return DataType(""), fmt.Errorf("not implemented yet")
}
