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
		ops, err := l.HandleStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error handling nested statement %T': %w", stmt, err)
		}
		fBody = append(fBody, ops...)
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
	ops, err := l.HandleExpression(statement.FuncCall)
	if err != nil {
		return nil, fmt.Errorf("error handling nested function call expression: %w", err)
	}

	// Do statements do not return a value, so we can just return the operations
	// TODO (hmny): Not sure about which segment I'll need to pop from, for now I assume Temp
	// TODO (hmny): Not sure about which offset I'll need to pop off, for now I assume 1
	return []vm.Operation{ops, vm.MemoryOp{Operation: vm.Pop, Segment: vm.Temp, Offset: 1}}, nil
}

// Specialized function to convert a 'jack.VarStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleVarStmt(statement VarStmt) ([]vm.Operation, error) {
	// ! Variable declaration does not produce any operation, it is just a declaration that will be
	// ! used later in the program. We could return an empty slice or nil, but let's be explicit.
	return []vm.Operation{}, nil
}

// Specialized function to convert a 'jack.LetStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleLetStmt(statement LetStmt) ([]vm.Operation, error) {
	rhsOps, err := l.HandleExpression(statement.Rhs)
	if err != nil {
		return nil, fmt.Errorf("error handling RHS expression: %w", err)
	}

	lhsOps, err := l.HandleExpression(statement.Lhs)
	if err != nil {
		return nil, fmt.Errorf("error handling LHS expression: %w", err)
	}

	// TODO(hmny): Add some glue code here to move RHS in LHS location
	return []vm.Operation{rhsOps, lhsOps}, nil
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

	// TODO(hmny): Randomize the label names to avoid collisions

	return []vm.Operation{
		vm.LabelDecl{Name: "WHILE_START"},
		condOps,
		vm.ArithmeticOp{Operation: vm.Neg},
		vm.GotoOp{Label: "WHILE_END", Jump: vm.Conditional},
		blockOps,
		vm.GotoOp{Label: "WHILE_START", Jump: vm.Unconditional},
		vm.LabelDecl{Name: "WHILE_END"},
	}, nil
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

	// TODO(hmny): Randomize the label names to avoid collisions

	// If there's no else block, we can just implement one way fork in the control flow
	if len(elseOps) == 0 {
		return []vm.Operation{
			condOps,
			vm.ArithmeticOp{Operation: vm.Neg},
			vm.GotoOp{Label: "ELSE", Jump: vm.Conditional},
			thenOps,
			vm.LabelDecl{Name: "ELSE"},
		}, nil
	}

	// If there is an else block, we need to do a two way fork in the control flow
	return []vm.Operation{
		condOps,
		vm.GotoOp{Label: "THEN", Jump: vm.Conditional},
		vm.GotoOp{Label: "ELSE", Jump: vm.Unconditional},
		vm.LabelDecl{Name: "THEN"},
		thenOps,
		vm.LabelDecl{Name: "ELSE"},
		elseOps,
	}, nil
}

// Specialized function to convert a 'jack.ReturnStmt' to a list of 'vm.Operation'.
func (l *Lowerer) HandleReturnStmt(statement ReturnStmt) ([]vm.Operation, error) {
	ops, err := l.HandleExpression(statement.Expr)
	if err != nil {
		return nil, fmt.Errorf("error handling return expression: %w", err)
	}

	return append(ops, vm.ReturnOp{}), nil
}


