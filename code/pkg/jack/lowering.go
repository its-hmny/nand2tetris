package jack

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	program     utils.OrderedMap[string, Class] // The program to lower, it must be not nil nor empty
	scopes      ScopeTable                      // Keeps track of the scopes and declared variables inside each one
	nRandomizer uint                            // Counter to randomize 'vm.LabelDecl(s)' with same name
}

// Initializes and returns to the caller a brand new 'Lowerer' struct.
// Requires the argument Program to be not nil nor empty.
func NewLowerer(p Program) Lowerer {
	// ? Why do we convert from a jack.Program (wrapper type of a map[string]Class to an OrderedMap[string, Class]?
	// Without doing this is impossible to have reproducible builds (and also meaningful test cases) because
	// the Go built-in map is not ordered and non-deterministic, so the order of iteration of the classes can
	// change on different runs, then what happens is that the label declarations will be different too since
	// they are randomized with just a counter (the counter will have different values because it will be
	// incremented a different number of times based on the order of the classes).
	//
	// The solution is simple: we order the map by its class name and store it in that order in the OrderedMap
	// so that the order we decided we'll be maintained throughout the entire lowering process. The end result
	// is that for the same input code we obtain always the same output code.

	//* 1. From unsorted map to unsorted slice of MapEntry[string, Class] (used later bu OrderedMap)
	classes := []utils.MapEntry[string, Class]{}
	for _, class := range p {
		classes = append(classes, utils.MapEntry[string, Class]{Key: class.Name, Value: class})
	}

	//* 2. We sort the slice by classname so that we have a reproducible order to use// 	//  map to unsorted slice of MapEntry[string, Class] (used later bu OrderedMap)
	sort.Slice(classes, func(i, j int) bool { return sort.StringsAreSorted([]string{classes[i].Key, classes[j].Key}) })

	//* 3. From sorted slice we create an order map where the insertion order and the alphabetic are the same
	return Lowerer{program: utils.NewOrderedMapFromList(classes), scopes: ScopeTable{}}
}

// Triggers the lowering process. It iterates class by class and then statement by statement
// and recursively calling the necessary helper function based on the construct type (much like
// a recursive descent parser but for lowering), this means the AST is visited in DFS order.
func (l *Lowerer) Lowerer() (vm.Program, error) {
	program := vm.Program{}
	if l.program.Size() == 0 {
		return nil, fmt.Errorf("the given 'program' is empty or nil")
	}

	for name, class := range l.program.Entries() {
		operations, err := l.HandleClass(class)
		if err != nil {
			return nil, fmt.Errorf("error handling lowering of class '%s': %w", name, err)
		}

		program[name] = vm.Module(operations)
	}

	return program, nil
}

// Specialized function to convert a 'jack.Class' node to a list of 'vm.Operation'.
func (l *Lowerer) HandleClass(class Class) ([]vm.Operation, error) {
	l.scopes.PushClassScope(class.Name) // Keep track of the current scope being processed
	defer l.scopes.PopClassScope()      // Reset the function name after processing

	operations := []vm.Operation{}

	for _, field := range class.Fields.Entries() {
		ops, err := l.HandleVarStmt(VarStmt{Vars: []Variable{field}})
		if err != nil {
			return nil, fmt.Errorf("error handling field '%s' in class '%s': %w", field.Name, class.Name, err)
		}
		operations = append(operations, ops...)
	}

	for _, subroutine := range class.Subroutines.Entries() {
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
	l.scopes.PushSubRoutineScope(subroutine.Name) // Keep track of the current subroutine function being processed
	defer l.scopes.PopSubroutineScope()           // Reset the function name after processing

	// When dealing with methods subroutine, where the object instance fields are available to be both read and written,
	// we will receive also the 'this' pointer as the first argument. The subroutine itself (in its prelude) will pop
	// that address from the argument memory segment and set the 'this' pointer accordingly.
	if subroutine.Type == Method {
		// ! The name is left purposefully empty, because is just there as placeholder for the upcoming/real
		// ! arguments that will be registered later on by iterating on the 'Subroutine.Arguments' field.
		l.scopes.RegisterVariable(Variable{Name: "__obj", VarType: Parameter, DataType: DataType{Main: Object, Subtype: ""}})
	}

	// We add to the current scope also all of the arguments of the subroutine
	for _, arg := range subroutine.Arguments {
		// Like this we're actually supporting shadowing of variables, so if a variable
		// with the same name is already present in the current scope, we just temporarily
		// override it with the most update one instead of returning an error (like Go does
		l.scopes.RegisterVariable(arg)
	}

	fName, fBody := l.scopes.GetScope(), []vm.Operation{}
	for _, stmt := range subroutine.Statements {
		ops, err := l.HandleStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error handling nested statement %T': %w", stmt, err)
		}
		fBody = append(fBody, ops...)
	}

	fDecl := vm.FuncDecl{Name: fName, NLocal: uint8(l.scopes.local.entries.Count())}

	// By convention, constructors will allocate the required memory for the object instance themselves and then set the
	// desired values for each address based on their own code logic. This is different, for example, from C++ constructors
	// where the memory is allocated externally by the caller (on the heap or the stack based on the code) and the constructor
	// only deals with initializing each field of the object instance to the desired value,
	if subroutine.Type == Constructor {
		// TODO (hmny): Pretty sure this can simplified and made more clear
		className := strings.Split(l.scopes.GetScope(), ".")[0] // Get the class name from the scope
		class, exists := l.program.Get(className)
		if !exists {
			return nil, fmt.Errorf("class '%s' not found", className)
		}

		nFields := uint16(0)
		for _, field := range class.Fields.Entries() {
			if field.VarType == Field { // Count only the fields, not the static ones
				nFields++
			}
		}

		preludeOps := []vm.Operation{
			// Each field is exactly one word long, so we can just allocate enough memory as fields declared in the class
			vm.MemoryOp{Operation: vm.Push, Segment: vm.Constant, Offset: nFields},
			vm.FuncCallOp{Name: "Memory.alloc", NArgs: 1},
			// We then set the 'this' pointer to the base pointer of the newly allocated memory
			vm.MemoryOp{Operation: vm.Pop, Segment: vm.Pointer, Offset: 0},
		}

		return append(append([]vm.Operation{fDecl}, preludeOps...), fBody...), nil
	}

	// By convention we'll receive the object instance pointer as the first argument on the stack. In order to
	// access correctly the object instance fields, we need to set the 'this' pointer based on the address received.
	if subroutine.Type == Method {
		preludeOps := []vm.Operation{
			vm.MemoryOp{Operation: vm.Push, Segment: vm.Argument, Offset: 0},
			vm.MemoryOp{Operation: vm.Pop, Segment: vm.Pointer, Offset: 0},
		}

		return append(append([]vm.Operation{fDecl}, preludeOps...), fBody...), nil
	}

	return append([]vm.Operation{fDecl}, fBody...), nil
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
	ops, err := l.HandleFuncCallExpr(statement.FuncCall)
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
		l.scopes.RegisterVariable(variable)
	}
	return []vm.Operation{}, nil // No operations needed for variable declaration, just update the scope
}

// Specialized function to convert a 'jack.LetStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleLetStmt(statement LetStmt) ([]vm.Operation, error) {
	// This is just the value to be assigned, nothing difficult about it
	rhsOps, err := l.HandleExpression(statement.Rhs)
	if err != nil {
		return nil, fmt.Errorf("error handling RHS expression: %w", err)
	}

	// If it's a VarExpr then we somewhat reuse the same logic as HandleVarExpr, but we need to write memory instead of reading
	if expr, isVarExpr := statement.Lhs.(VarExpr); isVarExpr {
		offset, variable, err := l.scopes.ResolveVariable(expr.Var)
		if err != nil {
			return nil, fmt.Errorf("error resolving variable '%s' in array expression: %w", expr.Var, err)
		}

		switch variable.VarType {
		case Local:
			return append(rhsOps, vm.MemoryOp{Operation: vm.Pop, Segment: vm.Local, Offset: offset}), nil
		case Parameter:
			return append(rhsOps, vm.MemoryOp{Operation: vm.Pop, Segment: vm.Argument, Offset: offset}), nil
		case Field:
			return append(rhsOps, vm.MemoryOp{Operation: vm.Pop, Segment: vm.This, Offset: offset}), nil
		case Static:
			return append(rhsOps, vm.MemoryOp{Operation: vm.Pop, Segment: vm.Static, Offset: offset}), nil
		default:
			return nil, fmt.Errorf("variable type '%s' is not supported yet", variable.VarType)
		}
	}

	// For ArrayExpr instead we reuse the pointer + offset logic from HandleArrayExpr but after that we write
	// a bit of glue code to save the RHS on temporary memory before loading the new address and writing it
	if expr, isArrayExpr := statement.Lhs.(ArrayExpr); isArrayExpr {
		baseOps, err := l.HandleVarExpr(VarExpr{Var: expr.Var})
		if err != nil {
			return nil, fmt.Errorf("error handling base variable expression: %w", err)
		}

		// Handle the index expression to get the offset of the array element
		indexOps, err := l.HandleExpression(expr.Index)
		if err != nil {
			return nil, fmt.Errorf("error handling index expression: %w", err)
		}

		// Calculates the specific element of array memory location that will be accessed later on
		refOps := append(append(indexOps, baseOps...), vm.ArithmeticOp{Operation: vm.Add})

		writeOps := []vm.Operation{ // Will move the value to temp and the pop it into the array's cell (That pointer)
			vm.MemoryOp{Operation: vm.Pop, Segment: vm.Temp, Offset: 0},
			vm.MemoryOp{Operation: vm.Pop, Segment: vm.Pointer, Offset: 1},
			vm.MemoryOp{Operation: vm.Push, Segment: vm.Temp, Offset: 0},
			vm.MemoryOp{Operation: vm.Pop, Segment: vm.That, Offset: 0},
		}

		return append(append(refOps, rhsOps...), writeOps...), nil
	}

	return nil, fmt.Errorf("LHS expression must be either a 'VarExpr' or an 'ArrayExpr', got: %T", statement.Lhs)
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
		vm.ArithmeticOp{Operation: vm.Not},
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
			vm.ArithmeticOp{Operation: vm.Not},
			vm.GotoOp{Label: fmt.Sprintf("ELSE_%d", l.nRandomizer), Jump: vm.Conditional}),
			thenOps...),
			vm.LabelDecl{Name: fmt.Sprintf("ELSE_%d", l.nRandomizer)},
		), nil
	}

	// If there is an else block, we need to do a two way fork in the control flow
	defer func() { l.nRandomizer += 3 }() // ! Increment the randomizer for next use

	return append(append(append(append(append(
		condOps,
		vm.GotoOp{Label: fmt.Sprintf("THEN_%d", l.nRandomizer), Jump: vm.Conditional},
		vm.GotoOp{Label: fmt.Sprintf("ELSE_%d", l.nRandomizer+1), Jump: vm.Unconditional},
		vm.LabelDecl{Name: fmt.Sprintf("THEN_%d", l.nRandomizer)}),
		thenOps...),
		vm.GotoOp{Label: fmt.Sprintf("END_%d", l.nRandomizer+2), Jump: vm.Unconditional},
		vm.LabelDecl{Name: fmt.Sprintf("ELSE_%d", l.nRandomizer+1)}),
		elseOps...),
		vm.LabelDecl{Name: fmt.Sprintf("END_%d", l.nRandomizer+2)},
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
	if expression.Var == "this" {
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Pointer, Offset: 0}}, nil
	}

	offset, variable, err := l.scopes.ResolveVariable(expression.Var)
	if err != nil {
		return nil, fmt.Errorf("error resolving variable '%s' in array expression: %w", expression.Var, err)
	}

	switch variable.VarType {
	case Local:
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Local, Offset: offset}}, nil
	case Parameter:
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Argument, Offset: offset}}, nil
	case Field:
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.This, Offset: offset}}, nil
	case Static:
		return []vm.Operation{vm.MemoryOp{Operation: vm.Push, Segment: vm.Static, Offset: offset}}, nil
	default:
		return nil, fmt.Errorf("variable type '%s' is not supported yet2", variable.VarType)
	}
}

// Specialized function to convert a 'jack.LiteralExpr' to a list of 'vm.Operation'.
func (l *Lowerer) HandleLiteralExpr(expression LiteralExpr) ([]vm.Operation, error) {
	switch expression.Type.Main {
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

	case Object:
		if expression.Value != "null" {
			return nil, fmt.Errorf("object literal are not supported '%s'", expression.Value)
		}
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
	case Negation:
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

	if !expression.IsExtCall { // Instance-to-instance function call
		// TODO (hmny): Pretty sure this can simplified and made more clear
		className := strings.Split(l.scopes.GetScope(), ".")[0] // Get the class name from the scope

		// Looks up whether the class and subroutine are defined and exists in the program.
		class, exists := l.program.Get(className)
		if !exists {
			return nil, fmt.Errorf("class defintion not found for '%s'", className)
		}
		routine, exists := class.Subroutines.Get(expression.FuncName)
		if !exists {
			return nil, fmt.Errorf("subroutine '%s' not found in class '%s'", expression.FuncName, className)
		}

		fName := fmt.Sprintf("%s.%s", className, expression.FuncName)

		if routine.Type == Method {
			// We push the 'this' pointer (already initialized) as the first argument to not break compatibility
			thisOp := vm.MemoryOp{Operation: vm.Push, Segment: vm.Pointer, Offset: 0}
			return append([]vm.Operation{thisOp}, append(argsInit, vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen + 1)})...), nil
		}

		return append(argsInit, vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen)}), nil
	}

	// We have an external function call and we check whether the target is a specific class instance.
	// In order to check whether we're hitting or not a class instance we check if in the scope(s) there's
	// an active variable with the same name as our expression.Var. This will also give us information about
	// how to populate the 'this', given that we will call only subroutine of Type = Method in this code path..
	if _, variable, _ := l.scopes.ResolveVariable(expression.Var); variable != (Variable{}) {
		if variable.DataType.Main != Object {
			return nil, fmt.Errorf("variable '%s' is not an object", expression.Var)
		}

		thisArg, err := l.HandleVarExpr(VarExpr{Var: expression.Var})
		if err != nil {
			return nil, fmt.Errorf("error handling variable expression for 'this' pointer: %w", err)
		}

		fName := fmt.Sprintf("%s.%s", variable.DataType.Subtype, expression.FuncName)
		return append(append(thisArg, argsInit...), vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen + 1)}), nil
	}

	// If we manage to reach here we are calling either a constructor or a function (like a static method).
	// This means that there will be no 'this' pointer to set and we can just call the function directly basically.
	// In case of a constructor the new problem is to allocate memory externally and then call the constructor to
	// set it as per its code logic, that's why we further fork the codepath based on the subroutine type.
	if class, isClass := l.program.Get(expression.Var); expression.IsExtCall && isClass {
		routine, exists := class.Subroutines.Get(expression.FuncName)
		if !exists {
			return nil, fmt.Errorf("subroutine '%s' not found in class '%s'", expression.FuncName, class.Name)
		}

		if routine.Type == Function {
			fName := fmt.Sprintf("%s.%s", class.Name, expression.FuncName)
			return append(argsInit, vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen)}), nil
		}

		if routine.Type == Constructor {
			fName := fmt.Sprintf("%s.new", class.Name) // All constructors are named 'new' in Jack
			return append(argsInit, vm.FuncCallOp{Name: fName, NArgs: uint8(argsLen)}), nil
		}

		return nil, fmt.Errorf("subroutine '%s' in class '%s' is not a function or constructor, got %s", expression.FuncName, class.Name, routine.Type)
	}

	return nil, fmt.Errorf("unrecognized function call expression: %s", expression.FuncName)
}
