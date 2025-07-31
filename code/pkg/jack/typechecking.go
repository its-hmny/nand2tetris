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
			return false, fmt.Errorf("error handling typechecking of class '%s': %w", name, err)
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
	for _, arg := range subroutine.Arguments {
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

	return true, nil
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
		_, variable, err := tc.scopes.ResolveVariable(expr.Var)
		if err != nil {
			return false, fmt.Errorf("error resolving variable '%s' in let expression: %w", expr.Var, err)
		}
		if !variable.DataType.Matches(rhs) {
			return false, fmt.Errorf("expected variable '%s' to be of type %s, got %s", expr.Var, variable.DataType, rhs)
		}

		return true, nil
	}

	// For ArrayExpr instead we reuse the pointer + offset logic from HandleArrayExpr but after that we write
	// a bit of glue code to save the RHS on temporary memory before loading the new address and writing it
	if expr, isArrayExpr := statement.Lhs.(ArrayExpr); isArrayExpr {
		_, variable, err := tc.scopes.ResolveVariable(expr.Var)
		if err != nil {
			return false, fmt.Errorf("error resolving variable '%s' in let expression: %w", expr.Var, err)
		}
		if !variable.DataType.Matches(DataType{Main: Array, Subtype: ""}) { // TODO (hmny): Array should be its own MainType and not a derived one
			return false, fmt.Errorf("expected variable '%s' to be of type %s, got %s", expr.Var, variable.DataType, rhs)
		}

		index, err := tc.HandleExpression(expr.Index)
		if err != nil {
			return false, fmt.Errorf("error handling index expression: %w", err)
		}
		if !index.Matches(DataType{Main: Int}) {
			return false, fmt.Errorf("array index expression must be 'int', got %s", expr.Index)
		}

		return true, nil
	}

	return false, fmt.Errorf("LHS expression must be either a 'VarExpr' or an 'ArrayExpr', got: %T", statement.Lhs)
}

// Specialized function to type-check a 'jack.IfStmt' and nested fields.
func (tc *TypeChecker) HandleIfStmt(statement IfStmt) (bool, error) {
	cond, err := tc.HandleExpression(statement.Condition)
	if err != nil {
		return false, fmt.Errorf("error handling if condition expression: %w", err)
	}
	if !cond.Matches(DataType{Main: Bool}) {
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
	if !cond.Matches(DataType{Main: Bool}) {
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
	if subroutine.Return.Matches(DataType{Main: Void}) && statement.Expr == nil {
		return true, nil
	}
	if subroutine.Return.Matches(DataType{Main: Void}) && statement.Expr != nil {
		return false, fmt.Errorf("return type of function is void but an expr has been provided")
	}

	// When the subroutine has a return type defined we need to check it against the actual return expression
	ret, err := tc.HandleExpression(statement.Expr)
	if err != nil {
		return false, fmt.Errorf("error handling return expression: %w", err)
	}
	if !subroutine.Return.Matches(ret) {
		return false, fmt.Errorf("expected return type %s, got %s", subroutine.Return, ret)
	}

	return true, nil
}

// Generalized function to type-check multiple expression their final 'jack.DataType'.
func (tc *TypeChecker) HandleExpression(expr Expression) (DataType, error) {
	switch tExpr := expr.(type) {
	case VarExpr:
		return tc.HandleVarExpr(tExpr)
	case LiteralExpr:
		return tc.HandleLiteralExpr(tExpr)
	case ArrayExpr:
		return tc.HandleArrayExpr(tExpr)
	case CastExpr:
		return tc.HandleCastExpr(tExpr)
	case UnaryExpr:
		return tc.HandleUnaryExpr(tExpr)
	case BinaryExpr:
		return tc.HandleBinaryExpr(tExpr)
	case FuncCallExpr:
		return tc.HandleFuncCallExpr(tExpr)
	default:
		return DataType{}, fmt.Errorf("unrecognized expression: %T", expr)
	}
}

// Specialized function to extract the DataType of a 'jack.VarExpr'.
func (tc *TypeChecker) HandleVarExpr(expression VarExpr) (DataType, error) {
	if expression.Var == "this" {
		// TODO (hmny): Pretty sure this can simplified and made more clear
		className := strings.Split(tc.scopes.GetScope(), ".")[0] // Get the class name from the scope
		return DataType{Main: Object, Subtype: className}, nil
	}

	_, variable, err := tc.scopes.ResolveVariable(expression.Var)
	if err != nil {
		return DataType{}, fmt.Errorf("error resolving variable '%s' in array expression: %w", expression.Var, err)
	}

	return variable.DataType, nil
}

// Specialized function to extract the DataType of a 'jack.LiteralExpr'.
func (tc *TypeChecker) HandleLiteralExpr(expression LiteralExpr) (DataType, error) {
	switch expression.Type.Main {
	case Int, Bool, Char, String:
		return expression.Type, nil // Classic passthrough for built-in data types
	case Object:
		if expression.Value != "null" {
			return DataType{}, fmt.Errorf("object literal are not supported '%s'", expression.Value)
		}
		return DataType{Main: Wildcard}, nil // TODO (hmny): Not sure if this is the correct way to handle null literal tbh
	default:
		return DataType{}, fmt.Errorf("unrecognized literal expression type: %s", expression.Type)
	}
}

// Specialized function to extract the DataType of a 'jack.ArrayExpr'.
func (tc *TypeChecker) HandleArrayExpr(expression ArrayExpr) (DataType, error) {
	array, err := tc.HandleVarExpr(VarExpr{Var: expression.Var})
	if err != nil {
		return DataType{}, fmt.Errorf("error handling base variable expression: %w", err)
	}
	if !array.Matches(DataType{Main: Array, Subtype: ""}) {
		return DataType{}, fmt.Errorf("variable %s must be an array, got %s", expression.Var, array.Main)
	}

	// Handle the index expression to get the offset of the array element
	index, err := tc.HandleExpression(expression.Index)
	if err != nil {
		return DataType{}, fmt.Errorf("error handling index expression: %w", err)
	}
	if !index.Matches(DataType{Main: Int}) {
		return DataType{}, fmt.Errorf("array index expression must be 'int', got %s", index)
	}

	return DataType{Main: Wildcard}, nil
}

// Specialized function to extract the DataType of a 'jack.CastExpr'.
func (tc *TypeChecker) HandleCastExpr(expression CastExpr) (DataType, error) {
	_, err := tc.HandleExpression(expression.Rhs)
	if err != nil {
		return DataType{}, fmt.Errorf("error handling nested expression: %w", err)
	}

	return expression.Type, nil
}

// Specialized function to extract the DataType of a 'jack.UnaryExpr'.
func (tc *TypeChecker) HandleUnaryExpr(expression UnaryExpr) (DataType, error) {
	nested, err := tc.HandleExpression(expression.Rhs)
	if err != nil {
		return DataType{}, fmt.Errorf("error handling nested expression: %w", err)
	}

	switch expression.Type {
	case Negation:
		if !nested.Matches(DataType{Main: Int}) {
			return DataType{}, fmt.Errorf("nested expression must be 'int', got %s", nested)
		}
		return DataType{Main: Int}, nil
	case BoolNot:
		if !nested.Matches(DataType{Main: Bool}) {
			return DataType{}, fmt.Errorf("nested expression must be 'bool', got %s", nested)
		}
		return DataType{Main: Bool}, nil
	default:
		return DataType{}, fmt.Errorf("unrecognized unary expression type: %s", expression.Type)
	}
}

// Specialized function to extract the DataType of a 'jack.BinaryExpr'.
func (tc *TypeChecker) HandleBinaryExpr(expression BinaryExpr) (DataType, error) {
	lhs, err := tc.HandleExpression(expression.Lhs)
	if err != nil {
		return DataType{}, fmt.Errorf("error handling nested LHS expression: %w", err)
	}

	rhs, err := tc.HandleExpression(expression.Rhs)
	if err != nil {
		return DataType{}, fmt.Errorf("error handling nested RHS expression: %w", err)
	}

	if !rhs.Matches(lhs) {
		return DataType{}, fmt.Errorf("RHS and LHS should have same type, got %s and %s", rhs, lhs)
	}

	switch expression.Type {
	case Plus, Minus, Divide, Multiply:
		return rhs, nil // Also lhs should be fine since they are the same DataType
	case BoolOr, BoolAnd, BoolNot:
		return DataType{Main: Bool}, nil
	case Equal, LessThan, GreatThan:
		return DataType{Main: Bool}, nil
	default:
		return DataType{}, fmt.Errorf("unrecognized binary expression type: %s", expression.Type)
	}
}

// Specialized function to extract the DataType of a 'jack.FuncCallExpr'.
func (tc *TypeChecker) HandleFuncCallExpr(expression FuncCallExpr) (DataType, error) {
	className := ""

	if _, variable, _ := tc.scopes.ResolveVariable(expression.Var); expression.IsExtCall && variable != (Variable{}) {
		// 1. We're calling a method of a specific object instance (e.g. a variable not a class name)
		if variable.DataType.Main != Object {
			return DataType{}, fmt.Errorf("variable '%s' is not an object type", expression.Var)
		}
		className = variable.DataType.Subtype

	} else if class, isClass := tc.program[expression.Var]; expression.IsExtCall && isClass {
		// 2. We're calling a function or constructor (static method) of a specific class
		className = class.Name
	} else if !expression.IsExtCall {
		// 3. Internal call to another method for the same class instance
		className = strings.Split(tc.scopes.GetScope(), ".")[0]
	} else {
		return DataType{}, fmt.Errorf("unsupported function call expression")
	}

	// Retrieve the current class and current subroutine information (checking for existence)
	class, exists := tc.program[className]
	if !exists {
		return DataType{}, fmt.Errorf("class %s doesn't exists", className)
	}
	subroutine, exists := class.Subroutines.Get(expression.FuncName)
	if !exists {
		return DataType{}, fmt.Errorf("subroutine %s doesn't exists for class %s", expression.FuncName, className)
	}

	for idx, expr := range expression.Arguments {
		arg, err := tc.HandleExpression(expr)
		if err != nil {
			return DataType{}, fmt.Errorf("error handling argument expression: %w", err)
		}

		if expected := subroutine.Arguments[idx].DataType; !arg.Matches(expected) {
			return DataType{}, fmt.Errorf("error handling arg no. %d, expected %s but got %s", idx, expected, arg)
		}
	}

	return subroutine.Return, nil
}
