package jack

import (
	"fmt"
	"strconv"

	"its-hmny.dev/nand2tetris/pkg/utils"
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

	// Keeps track of the class (.jack file) we're lowering at the moment,
	// used to convert internal function calls (e.g. div(X, y) => Math.div(x,y))
	classModule string
	// Keeps track of the scope (in this case only a subroutine) we're lowering
	// at the moment, used to track active variable declaration in the ScopeTable
	classScope string

	nRandomizer uint // Counter to randomize 'vm.LabelDecl(s)' with same name

	scopes ScopeTable // List of active scopes to lookup variables given the current context
}

type ActiveEntry struct {
	ScopeName string
	Variable  Variable
}

type ScopeTable map[VarType]*utils.Stack[ActiveEntry]

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument Program to be not nil nor empty.
func NewLowerer(p Program) Lowerer {
	builtin := ScopeTable{
		Local:     utils.NewStack[ActiveEntry](),
		Field:     utils.NewStack[ActiveEntry](),
		Parameter: utils.NewStack[ActiveEntry](),
		Static: utils.NewStack[ActiveEntry](
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Sys", Type: Static, DataType: Object, ClassName: "Sys"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Math", Type: Static, DataType: Object, ClassName: "Math"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Array", Type: Static, DataType: Object, ClassName: "Array"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Output", Type: Static, DataType: Object, ClassName: "Output"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Memory", Type: Static, DataType: Object, ClassName: "Memory"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Screen", Type: Static, DataType: Object, ClassName: "Screen"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "String", Type: Static, DataType: Object, ClassName: "String"}},
			ActiveEntry{ScopeName: "Global", Variable: Variable{Name: "Keyboard", Type: Static, DataType: Object, ClassName: "Keyboard"}},
		),
	}

	return Lowerer{program: p, scopes: builtin}
}

// Triggers the lowering process. It iterates class by class and then statement by statement
// and recursively calling the necessary helper function based on the construct type (much like
// a recursive descent parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lowerer() (vm.Program, error) {
	program := vm.Program{}
	if len(l.program) == 0 {
		return nil, fmt.Errorf("the given 'program' is empty or nil")
	}

	for name, class := range l.program {
		operations, err := l.HandleClass(class)
		if err != nil {
			return nil, fmt.Errorf("error handling lowering of class '%s': %w", name, err)
		}

		program[name] = vm.Module(operations)
	}

	return program, nil
}

// Searches for a variable by name iteratively in all scopes, starting from the most nested one going upwards.
func (l *Lowerer) ResolveVariable(name string) (uint16, Variable, error) {
	for _, vartype := range []VarType{Local, Parameter, Field, Static} {
		for idx, entry := range l.scopes[vartype].Iterator() {
			if entry.Variable.Name == name {
				return uint16(idx), entry.Variable, nil
			}

		}
	}

	return 0, Variable{}, fmt.Errorf("variable '%s' undeclared, not found in any scope", name)
}

// Specialized function to convert a 'jack.Class' node to a list of 'vm.Operation'.
func (l *Lowerer) HandleClass(class Class) ([]vm.Operation, error) {
	l.classModule, l.classScope = class.Name, "Global"      // Keep track of the current scope being processed
	defer func() { l.classModule, l.classScope = "", "" }() // Reset the function name after processing

	operations := []vm.Operation{}

	for _, field := range class.Fields {
		ops, err := l.HandleVarStmt(VarStmt{Vars: []Variable{field}})
		if err != nil {
			return nil, fmt.Errorf("error handling field '%s' in class '%s': %w", field.Name, class.Name, err)
		}
		operations = append(operations, ops...)
	}

	for _, subroutine := range class.Subroutines {
		ops, err := l.HandleSubroutine(subroutine)
		if err != nil {
			return nil, fmt.Errorf("error handling subroutine '%s' in class '%s': %w", subroutine.Name, class.Name, err)
		}
		operations = append(operations, ops...)
	}

	return operations, nil
}

// Specialized function to convert a 'jack.Subroutine' node to a list of 'vm.Operation'.
func (l *Lowerer) HandleSubroutine(subroutine Subroutine) ([]vm.Operation, error) {
	l.classModule, l.classScope = l.classModule, subroutine.Name // Keep track of the current subroutine function being processed
	defer func() { l.classModule, l.classScope = "", "" }()      // Reset the function name after processing

	fName, fBody := fmt.Sprintf("%s.%s", l.classModule, subroutine.Name), []vm.Operation{}
	for _, stmt := range subroutine.Statements {
		ops, err := l.HandleStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error handling nested statement %T': %w", stmt, err)
		}
		fBody = append(fBody, ops...)
	}

	// We need to count the number of local variable to be allocated for a function as part of its vm.FuncDecl
	nLocal, currentScope := 0, fmt.Sprintf("%s.%s", l.classModule, l.classScope)
	for _, active := range l.scopes {
		for _, entry := range active.Iterator() {
			if entry.ScopeName == currentScope {
				nLocal++
			}
		}
	}

	return append([]vm.Operation{vm.FuncDecl{Name: fName, NLocal: uint8(nLocal)}}, fBody...), nil
}

// Generalized function to lower multiple statements types returning a 'vm.Operation' list.
func (l *Lowerer) HandleStatement(stmt Statement) ([]vm.Operation, error) {
	switch tStmt := stmt.(type) {
	case DoStmt:
		return l.HandleDoStmt(tStmt)
	case VarStmt:
		return l.HandleVarStmt(tStmt)
	case LetStmt:
		return l.HandleLetStmt(tStmt)
	case IfStmt:
		return l.HandleIfStmt(tStmt)
	case WhileStmt:
		return l.HandleWhileStmt(tStmt)
	case ReturnStmt:
		return l.HandleReturnStmt(tStmt)
	default:
		return nil, fmt.Errorf("unrecognized statement: %T", stmt)
	}
}

// Specialized function to convert a 'jack.DoStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleDoStmt(statement DoStmt) ([]vm.Operation, error) {
	ops, err := l.HandleExpression(statement.FuncCall)
	if err != nil {
		return nil, fmt.Errorf("error handling nested function call expression: %w", err)
	}

	// Do statements do not return a value, so we can just drop whatever has been returned
	return append(ops, vm.MemoryOp{Operation: vm.Pop, Segment: vm.Temp, Offset: 0}), nil
}

// Specialized function to convert a 'jack.VarStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleVarStmt(statement VarStmt) ([]vm.Operation, error) {
	for _, variable := range statement.Vars {
		// Like this we're actually supporting shadowing of variables, so if a variable
		// with the same name is already present in the current scope, we just temporarily
		// override it with the most update one instead of returning an error (like Go does BTW).
		scope := fmt.Sprintf("%s.%s", l.classModule, l.classScope)
		l.scopes[variable.Type].Push(ActiveEntry{ScopeName: scope, Variable: variable})
	}
	return []vm.Operation{}, nil // No operations needed for variable declaration, just update the scope
}

// Specialized function to convert a 'jack.LetStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleLetStmt(statement LetStmt) ([]vm.Operation, error) {
	lhsOps, err := l.HandleExpression(statement.Lhs)
	if err != nil {
		return nil, fmt.Errorf("error handling LHS expression: %w", err)
	}

	rhsOps, err := l.HandleExpression(statement.Rhs)
	if err != nil {
		return nil, fmt.Errorf("error handling RHS expression: %w", err)
	}

	return append(rhsOps, lhsOps...), nil // TODO (hmny): Still some things to go here
}

// Specialized function to convert a 'jack.WhileStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleWhileStmt(statement WhileStmt) ([]vm.Operation, error) {
	condOps, err := l.HandleExpression(statement.Condition)
	if err != nil {
		return nil, fmt.Errorf("error handling while condition expression: %w", err)
	}

	blockOps := []vm.Operation{}

	for _, stmt := range statement.Block {
		ops, err := l.HandleStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error handling statement in while block: %w", err)
		}
		blockOps = append(blockOps, ops...)
	}

	defer func() { l.nRandomizer += 2 }() // ! Increment the randomizer for next use

	return append(append(append(append(
		[]vm.Operation{vm.LabelDecl{Name: fmt.Sprintf("WHILE_START_%d", l.nRandomizer)}},
		condOps...),
		vm.ArithmeticOp{Operation: vm.Neg},
		vm.GotoOp{Label: fmt.Sprintf("WHILE_END_%d", l.nRandomizer+1), Jump: vm.Conditional}),
		blockOps...),
		vm.GotoOp{Label: fmt.Sprintf("WHILE_START_%d", l.nRandomizer), Jump: vm.Unconditional},
		vm.LabelDecl{Name: fmt.Sprintf("WHILE_END_%d", l.nRandomizer+1)},
	), nil
}

// Specialized function to convert a 'jack.IfStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleIfStmt(statement IfStmt) ([]vm.Operation, error) {
	condOps, err := l.HandleExpression(statement.Condition)
	if err != nil {
		return nil, fmt.Errorf("error handling if condition expression: %w", err)
	}

	thenOps, elseOps := []vm.Operation{}, []vm.Operation{}

	for _, stmt := range statement.ThenBlock {
		ops, err := l.HandleStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error handling statement in 'then' block: %w", err)
		}
		thenOps = append(thenOps, ops...)
	}

	for _, stmt := range statement.ElseBlock {
		ops, err := l.HandleStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error handling statement in 'else' block: %w", err)
		}
		elseOps = append(elseOps, ops...)
	}

	// If there's no else block, we can just implement one way fork in the control flow
	if len(elseOps) == 0 {
		defer func() { l.nRandomizer += 1 }() // ! Increment the randomizer for next use

		return append(append(append(
			condOps,
			vm.ArithmeticOp{Operation: vm.Neg},
			vm.GotoOp{Label: fmt.Sprintf("ELSE_%d", l.nRandomizer), Jump: vm.Conditional}),
			thenOps...),
			vm.LabelDecl{Name: fmt.Sprintf("ELSE_%d", l.nRandomizer)},
		), nil
	}

	// If there is an else block, we need to do a two way fork in the control flow
	defer func() { l.nRandomizer += 2 }() // ! Increment the randomizer for next use

	return append(append(append(append(
		condOps,
		vm.GotoOp{Label: fmt.Sprintf("THEN_%d", l.nRandomizer), Jump: vm.Conditional},
		vm.GotoOp{Label: fmt.Sprintf("ELSE_%d", l.nRandomizer+1), Jump: vm.Unconditional},
		vm.LabelDecl{Name: fmt.Sprintf("THEN_%d", l.nRandomizer)}),
		thenOps...),
		vm.LabelDecl{Name: fmt.Sprintf("ELSE_%d", l.nRandomizer+1)}),
		elseOps...,
	), nil
}

// Specialized function to convert a 'jack.ReturnStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleReturnStmt(statement ReturnStmt) ([]vm.Operation, error) {
	if statement.Expr == nil { // No expression means just a zero-value return
		return []vm.Operation{
			vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: 0},
			vm.ReturnOp{},
		}, nil
	}

	ops, err := l.HandleExpression(statement.Expr)
	if err != nil {
		return nil, fmt.Errorf("error handling return expression: %w", err)
	}

	return append(ops, vm.ReturnOp{}), nil
}

// Generalized function to lower multiple expression types returning a 'vm.Operation' list.
func (l *Lowerer) HandleExpression(expr Expression) ([]vm.Operation, error) {
	switch tExpr := expr.(type) {
	case VarExpr:
		return l.HandleVarExpr(tExpr)
	case LiteralExpr:
		return l.HandleLiteralExpr(tExpr)
	case ArrayExpr:
		return l.HandleArrayExpr(tExpr)
	case UnaryExpr:
		return l.HandleUnaryExpr(tExpr)
	case BinaryExpr:
		return l.HandleBinaryExpr(tExpr)
	case FuncCallExpr:
		return l.HandleFuncCallExpr(tExpr)
	default:
		return nil, fmt.Errorf("unrecognized expression: %T", expr)
	}
}

// Specialized function to convert a 'jack.VarExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleVarExpr(expression VarExpr) ([]vm.Operation, error) {
	offset, variable, err := l.ResolveVariable(expression.Var)
	if err != nil {
		return nil, fmt.Errorf("error resolving variable '%s' in array expression: %w", expression.Var, err)
	}

	switch variable.Type {
	case Local:
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Local, Offset: offset}}, nil
	case Parameter:
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Argument, Offset: offset}}, nil
	default:
		return nil, fmt.Errorf("variable type '%s' is not supported yet2", variable.Type)
	}
}

// Specialized function to convert a 'jack.LiteralExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleLiteralExpr(expression LiteralExpr) ([]vm.Operation, error) {
	switch expression.Type {
	case Int:
		value, err := strconv.ParseUint(expression.Value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("error parsing integer literal '%s': %w", expression.Value, err)
		}

		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: uint16(value)}}, nil

	case Bool:
		value, err := strconv.ParseBool(expression.Value)
		if err != nil {
			return nil, fmt.Errorf("error parsing integer literal '%s': %w", expression.Value, err)
		}

		mapping := map[bool]uint16{true: 1, false: 0}
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: mapping[value]}}, nil

	case Char:
		if len(expression.Value) != 1 {
			return nil, fmt.Errorf("error parsing char literal '%s'", expression.Value)
		}

		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: uint16(expression.Value[0])}}, nil

	case Null: // TODO (hmny): Not sure that pushing just a simple 0 is enough or right
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: 0}}, nil

	case String:
		ops := []vm.Operation{
			// Reserves/Allocates enough space for the entire string literal via the constructor
			vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: uint16(len(expression.Value))},
			vm.FuncCallOp{Name: "String.new", NArgs: 1},
		}

		for _, char := range expression.Value {
			// Set each character in the string literal one by one until completion
			ops = append(ops, vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: uint16(char)})
			ops = append(ops, vm.FuncCallOp{Name: "String.appendChar", NArgs: 2})
		}

		return ops, nil

	default:
		return nil, fmt.Errorf("unrecognized literal expression type: %s", expression.Type)
	}
}

// Specialized function to convert a 'jack.ArrayExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleArrayExpr(expression ArrayExpr) ([]vm.Operation, error) {
	baseOps, err := l.HandleVarExpr(VarExpr{Var: expression.Var})
	if err != nil {
		return nil, fmt.Errorf("error handling base variable expression: %w", err)
	}

	// Handle the index expression to get the offset of the array element
	indexOps, err := l.HandleExpression(expression.Index)
	if err != nil {
		return nil, fmt.Errorf("error handling index expression: %w", err)
	}

	// We need to add the index to the base address of the array
	return append(append(indexOps, baseOps...),
		vm.ArithmeticOp{Operation: vm.Add},
		// Add the pointer + offset and then set the 'That' pointer to the memory location
		vm.MemoryOp{Operation: vm.Pop, Segment: vm.Pointer, Offset: 1},
		vm.MemoryOp{Operation: vm.Push, Segment: vm.That, Offset: 0},
	), nil
}

// Specialized function to convert a 'jack.UnaryExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleUnaryExpr(expression UnaryExpr) ([]vm.Operation, error) {
	ops, err := l.HandleExpression(expression.Rhs)
	if err != nil {
		return nil, fmt.Errorf("error handling nested expression: %w", err)
	}

	switch expression.Type {
	case Minus:
		return append(ops, vm.ArithmeticOp{Operation: vm.Neg}), nil
	case BoolNot:
		return append(ops, vm.ArithmeticOp{Operation: vm.Not}), nil
	default:
		return nil, fmt.Errorf("unrecognized unary expression type: %s", expression.Type)
	}
}

// Specialized function to convert a 'jack.BinaryExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleBinaryExpr(expression BinaryExpr) ([]vm.Operation, error) {
	lhsOps, err := l.HandleExpression(expression.Lhs)
	if err != nil {
		return nil, fmt.Errorf("error handling nested LHS expression: %w", err)
	}

	rhsOps, err := l.HandleExpression(expression.Rhs)
	if err != nil {
		return nil, fmt.Errorf("error handling nested RHS expression: %w", err)
	}

	switch expression.Type {
	case Plus:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Add}), nil
	case Minus:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Sub}), nil
	case Divide:
		return append(append(lhsOps, rhsOps...), vm.FuncCallOp{Name: "Math.divide", NArgs: 2}), nil
	case Multiply:
		return append(append(lhsOps, rhsOps...), vm.FuncCallOp{Name: "Math.multiply", NArgs: 2}), nil
	case BoolOr:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Or}), nil
	case BoolAnd:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.And}), nil
	case BoolNot:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Not}), nil
	case Equal:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Eq}), nil
	case LessThan:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Lt}), nil
	case GreatThan:
		return append(append(lhsOps, rhsOps...), vm.ArithmeticOp{Operation: vm.Gt}), nil
	default:
		return nil, fmt.Errorf("unrecognized binary expression type: %s", expression.Type)
	}
}

// Specialized function to convert a 'jack.FuncCallExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleFuncCallExpr(expression FuncCallExpr) ([]vm.Operation, error) {
	argsInit, argsLen := []vm.Operation{}, len(expression.Arguments)

	for _, expr := range expression.Arguments {
		ops, err := l.HandleExpression(expr)
		if err != nil {
			return nil, fmt.Errorf("error handling argument expression: %w", err)
		}

		argsInit = append(argsInit, ops...)
	}

	if expression.IsExtCall {
		_, variable, err := l.ResolveVariable(expression.Var)
		if err != nil {
			return nil, fmt.Errorf("error resolving variable '%s' in array expression: %w", expression.Var, err)
		}
		if variable.DataType != Object {
			return nil, fmt.Errorf("variable '%s' is not an object", expression.Var)
		}

		// TODO (hmny): Pretty sure here I have to do something with the this pointer
		fName := fmt.Sprintf("%s.%s", variable.ClassName, expression.FuncName)
		return append(argsInit, vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen)}), nil
	}

	fName := fmt.Sprintf("%s.%s", l.classModule, expression.FuncName)
	return append(argsInit, vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen)}), nil
}
