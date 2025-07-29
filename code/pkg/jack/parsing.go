package jack

import (
	"fmt"
	"io"
	"os"
	"strings"

	pc "github.com/prataprc/goparsec"
	"its-hmny.dev/nand2tetris/pkg/utils"
)

var ast = pc.NewAST("jack_program", 0)

var (
	pClass = ast.And("class_decl", nil,
		ast.Kleene("file_header", nil, pComment),
		pc.Atom("class", "CLASS"), pIdent, pLBrace,
		ast.Kleene("fields_or_comments", nil, ast.OrdChoice("items", nil, pField, pComment)),
		ast.Kleene("routines_or_comments", nil, ast.OrdChoice("items", nil, pRoutines, pComment)),
		pRBrace,
	)

	pField = ast.And("field_decl", nil,
		pFieldType, pDataType,
		// ! The 'Many' combinator is used because both of these are valid Jack syntax:
		// ! - 'field int test;'
		// ! - 'field int numerator, denominator;'
		ast.Many("items", nil, pIdent, pComma), pSemi,
	)

	pRoutines = ast.And("routine_decl", nil,
		// Func keyword, return type and function/method name
		pRoutineType, pDataType, pIdent,
		// '(', comma separated argument type(s) and name(s), ')'
		pLParen, ast.Kleene("arguments", nil, ast.And("argument", nil, pDataType, pIdent), pComma), pRParen,
		// '{', statement and or comments (s), '}'
		pLBrace, ast.Kleene("statements_or_comments", nil, ast.OrdChoice("item", nil, &pStatement, pComment)), pRBrace,
	)

	// TODO (hmny): We need to inject comment parsing everywhere basically
	pComment = ast.OrdChoice("comment", nil,
		// Single line comments (e.g. "// This is a comment")
		ast.And("sl_comment", nil, pc.Atom("//", "//"), pc.Token(`(?m).*$`, "COMMENT")),
		// Multi line comments (e.g. "/* This is a comment */")
		ast.And("ml_comment", nil, pc.Token(`/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`, "COMMENT")),
	)
)

var (
	// Top level generic statement parser, declared like this to allow cyclical references.
	// An example of a statement that has the need to parse other nested statements is 'pWhileStmt'.
	pStatement pc.Parser

	pDoStmt = ast.And("do_stmt", nil,
		// Support both external method call and local method call syntax:
		// - 'External': call to another class method (e.g. 'do X.ExtMethod()')
		// - 'Local': call to same class/instance method (e.g. 'do InternalMethod()')
		pc.Atom("do", "DO"), pFunCallExpr, pSemi,
	)

	pVarStmt = ast.And("var_stmt", nil, pc.Atom("var", "VAR"), pDataType, ast.Many("variables", nil, pIdent, pComma), pSemi)

	pLetStmt = ast.And("let_stmt", nil, pc.Atom("let", "LET"), ast.OrdChoice("lhs", nil, pArrayExpr, pIdent), pc.Atom("=", "EQUAL"), &pExpr, pSemi)

	pReturnStmt = ast.And("return_stmt", nil, pc.Atom("return", "RETURN"), ast.Maybe("expr", nil, &pExpr), pSemi)

	pIfStmt = ast.And("if_stmt", nil,
		pc.Atom("if", "IF"), pLParen, &pExpr, pRParen, pLBrace,
		ast.Kleene("statements_or_comments", nil, ast.OrdChoice("item", nil, &pStatement, pComment)), pRBrace,
		ast.Maybe("else_opt", nil, ast.And("else_stmt", nil,
			pc.Atom("else", "ELSE"), pLBrace,
			ast.Kleene("statements_or_comments", nil, ast.OrdChoice("item", nil, &pStatement, pComment)),
			pRBrace,
		)),
	)

	pWhileStmt = ast.And("while_stmt", nil,
		pc.Atom("while", "WHILE"), pLParen, &pExpr, pRParen, pLBrace,
		ast.Kleene("statements_or_comments", nil, ast.OrdChoice("item", nil, &pStatement, pComment)), pRBrace,
	)
)

var (
	// Top level generic expression parser, declared like this to allow cyclical references.
	// An example of a expression that has the need to parse other nested expr is (1.0 * (2 / 3)).
	pExpr, pTerm pc.Parser

	// ! The order of this PCs is important: by putting Int() before Float() we'll not be able to parse a float
	// !completely because the integer part will be picked up by the Int() PC before given back control to PExpr.
	pLiteral = ast.OrdChoice("literal", nil,
		// Basic literals (int, char and bool)
		pc.Int(), pc.Char(), pc.Token("true", "TRUE"), pc.Token("false", "FALSE"),
		// also here we parse 'null' and 'this
		pc.Token("null", "NULL"), pc.Token("this", "THIS"),
		// finally we parse string literals
		pc.Token(`"(?:\\.|[^"\\])*"`, "STRING"),
	)

	pArrayExpr = ast.And("array_expr", nil, pIdent, pc.Atom("[", "RSQUARE"), &pExpr, pc.Atom("]", "LSQUARE"))

	pUnaryExpr = ast.And("unary_expr", nil,
		// Unary operations supported by the Jack language (boolean and arithmetic negation)
		ast.OrdChoice("op", nil, pc.Atom("-", "NEGATION"), pc.Atom("~", "BOOL_NEG")),
		&pTerm, // Nested subexpression or term to be evaluated
	)

	pBinaryExpr = ast.And("binary_expr", nil,
		&pTerm, // Nested subexpression or term to be evaluated
		ast.OrdChoice("op", nil,
			// Bitwise binary operations
			pc.Atom("|", "BOOL_OR"), pc.Atom("&", "BOOL_AND"),
			// Comparison operations
			pc.Atom("=", "EQUAL"), pc.Atom("<", "LESS_THAN"), pc.Atom(">", "GREATER_THAN"),
			// Arithmetic operations
			pc.Atom("+", "PLUS"), pc.Atom("-", "MINUS"), pc.Atom("/", "DIVIDE"), pc.Atom("*", "MULTIPLY"),
		),
		&pTerm, // Nested subexpression or term to be evaluated
	)

	pFunCallExpr = ast.And("funcall_expr", nil,
		// Support both external method call and local method call syntax:
		// - 'External': call to another class method (e.g. 'do X.ExtMethod()')
		// - 'Local': call to same class/instance method (e.g. 'do InternalMethod()')
		ast.Many("qualifiers", nil, pIdent, pDot),
		// '(', comma separated argument passing w/ expression to be eval'd, ')'
		pLParen, ast.Kleene("args", nil, &pExpr, pComma), pRParen,
	)
)

var (
	// Generic Identifier parser (for label and function declaration)
	// NOTE: An ident can be any sequence of letters, digits, and symbols (_, ., $, :).
	// NOTE: An ident cannot begin with a leading digit (a symbol is indeed allowed).
	pIdent = pc.Token(`[A-Za-z_$:][0-9a-zA-Z_$:]*`, "IDENT")

	pDot    = pc.Atom(".", "DOT")
	pSemi   = pc.Atom(";", "SEMI")
	pComma  = pc.Atom(",", "COMMA")
	pLParen = pc.Atom("(", "RPAREN")
	pRParen = pc.Atom(")", "RPAREN")
	pLBrace = pc.Atom("{", "LBRACE")
	pRBrace = pc.Atom("}", "RBRACE")

	// Different types of field declarations, each has its own meaning:
	// - field: For classic OOP-like fields (accessed only by the object instance)
	// - static: For Java-like static fields (accessed by all the object instances)
	pFieldType = ast.OrdChoice("method_type", nil,
		pc.Atom("field", "FIELD"), pc.Atom("static", "STATIC"),
	)

	// Different types od routine declarations, each has its own meaning:
	// - constructor: For constructor (just one per class) method (to create the object instance)
	// - function:  For Java-like static functions (w/o access to the object instance)
	// - method: For classic OOP-like class methods (w/ access to the object instance)
	pRoutineType = ast.OrdChoice("method_type", nil,
		pc.Atom("constructor", "CONSTRUCTOR"), pc.Atom("function", "FUNCTION"), pc.Atom("method", "METHOD"),
	)

	// Built-in (also known as primitive) data types allowed/provided by the Jack language.
	pDataType = ast.OrdChoice("data_type", nil,
		pc.Atom("int", "INT"), pc.Atom("char", "CHAR"), pc.Atom("boolean", "BOOL"),
		pc.Atom("null", "NULL"), pc.Atom("void", "VOID"), pIdent,
	)
)

func init() {
	pStatement = ast.OrdChoice("item", nil, pDoStmt, pVarStmt, pLetStmt, pIfStmt, pWhileStmt, pReturnStmt)

	pExpr = ast.OrdChoice("expression", nil, pBinaryExpr, pUnaryExpr, pFunCallExpr, pArrayExpr, pLiteral, pIdent, ast.And("subexpr", nil, pLParen, &pExpr, pRParen))
	pTerm = ast.OrdChoice("term", nil, pFunCallExpr, pArrayExpr, pLiteral, pIdent, ast.And("subexpr", nil, pLParen, &pExpr, pRParen))
}

// ----------------------------------------------------------------------------
// Jack Parser

// This section defines the Parser for the nand2tetris Jack language.
//
// It uses parser combinator(s) to obtain the AST from the source code (the latter can be provided)
// in multiple ways using a generic io.Reader, the library reads up the feature flags (as env vars):
// - PARSEC_DEBUG: Verbose logging to inspect which of the PCs gets triggered and match
// - EXPORT_AST:   Exports in the DEBUG_FOLDER a Graphviz representation of the AST
// - PRINT_AST:    Print on the stdout a textual representation of the AST
type Parser struct{ reader io.Reader }

// Initializes and returns to the caller a brand new 'Parser' struct.
// Requires the argument io.Reader 'r' to be valid and usable.
func NewParser(r io.Reader) Parser {
	return Parser{reader: r}
}

// Parser entrypoint divides the 2 phases of the parsing pipeline
// Text --> AST: This step is done using PCs and returns a generic traversable AST
// AST --> IR: This step is done by traversing the AST and extracting the 'vm.Module'
func (p *Parser) Parse() (Class, error) {
	content, err := io.ReadAll(p.reader)
	if err != nil {
		return Class{}, fmt.Errorf("cannot read from 'io.Reader': %s", err)
	}

	ast, success := p.FromSource(content)
	if !success {
		return Class{}, fmt.Errorf("failed to parse AST from input content")
	}

	return p.FromAST(ast)
}

// Scans the textual input stream coming from the 'reader' method and returns a traversable AST
// (Abstract Syntax Tree) that can be eventually visited to extract/transform the info available.
func (p *Parser) FromSource(source []byte) (pc.Queryable, bool) {

	// Feature flag: Enable 'goparsec' library's debug logs
	if os.Getenv("PARSEC_DEBUG") != "" {
		ast.SetDebug()
	}

	// We generate the traversable Abstract Syntax Tree from the source content
	root, _ := ast.Parsewith(pClass, pc.NewScanner(source))

	// Feature flag: Enables export of the AST as Dot file (debug.ast.fot)
	if os.Getenv("EXPORT_AST") != "" {
		file, _ := os.Create(fmt.Sprintf("%s/debug.ast.dot", os.Getenv("DEBUG_FOLDER")))
		defer file.Close()

		file.Write([]byte(ast.Dotstring("\"JACK AST\"")))
	}

	// Feature flag: Enables pretty printing of the AST on the console
	if os.Getenv("PRINT_AST") != "" {
		ast.Prettyprint()
	}
	// TODO (hmny): This hardcoding to true should be changed
	return root, true // Success is based on the reaching of 'EOF'
}

// This function takes the root node of the raw parsed AST and does a DFS on it parsing
// one by one each subtree and retuning a 'jack.Class' that can be used as in-memory and
// type-safe AST not dependent on the parsing library used.
func (p *Parser) FromAST(root pc.Queryable) (Class, error) {
	if root.GetName() != "class_decl" {
		return Class{}, fmt.Errorf("expected node 'class_decl', found %s", root.GetName())
	}
	if len(root.GetChildren()) != 7 {
		return Class{}, fmt.Errorf("expected node with 7 leaf, got %d", len(root.GetChildren()))
	}

	class := Class{
		Name:        root.GetChildren()[2].GetValue(),
		Fields:      utils.OrderedMap[string, Variable]{},
		Subroutines: utils.OrderedMap[string, Subroutine]{},
	}

	// Field declaration subtree, appends 'jack.Variable' to 'class.Fields'
	for _, node := range root.GetChildren()[4].GetChildren() {
		if node.GetName() == "sl_comment" || node.GetName() == "ml_comment" { // Skip comments
			continue
		}
		fields, err := p.HandleFieldDecl(node)
		if err != nil {
			return Class{}, err
		}
		for _, field := range fields {
			class.Fields.Set(field.Name, field)
		}
	}

	// Method declaration subtree, appends 'jack.Subroutine' to 'class.Subroutines'
	for _, node := range root.GetChildren()[5].GetChildren() {
		if node.GetName() == "sl_comment" || node.GetName() == "ml_comment" { // Skip comments
			continue
		}
		subroutine, err := p.HandleSubroutineDecl(node)
		if err != nil {
			return Class{}, err
		}
		class.Subroutines.Set(subroutine.Name, subroutine)
	}

	return class, nil
}

// Specialized function to convert a "field_decl" node to a '[]jack.Variable'.
func (Parser) HandleFieldDecl(node pc.Queryable) ([]Variable, error) {
	if node.GetName() != "field_decl" {
		return nil, fmt.Errorf("expected node 'field_decl', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 4 {
		return nil, fmt.Errorf("expected node with 4 leaf, got %d", len(node.GetChildren()))
	}

	fieldType, dataType := VarType(node.GetChildren()[0].GetValue()), node.GetChildren()[1].GetValue()

	nested, fields := node.GetChildren()[2].GetChildren(), []Variable{}
	if len(nested) < 1 {
		return nil, fmt.Errorf("expected at least one field declaration, got %d", len(nested))
	}

	// Iterate on the nested possible n declarations to extract all the variable names
	for _, child := range nested {
		if child.GetName() != "IDENT" {
			return nil, fmt.Errorf("expected node 'IDENT', got %s", child.GetName())
		}

		// Primitive data types (int, string, bool) are handled differently than complex objects
		if builtin := MainType(dataType); builtin == Int || builtin == String || builtin == Bool || builtin == Char {
			fields = append(fields, Variable{Name: child.GetValue(), VarType: fieldType, DataType: DataType{Main: builtin}})
			continue
		}

		fields = append(fields, Variable{Name: child.GetValue(), VarType: fieldType, DataType: DataType{Main: Object, Subtype: dataType}})
	}

	return fields, nil
}

// Specialized function to convert a "routine_decl" node to a 'jack.Routine'.
func (p *Parser) HandleSubroutineDecl(node pc.Queryable) (Subroutine, error) {
	if node.GetName() != "routine_decl" {
		return Subroutine{}, fmt.Errorf("expected node 'routine_decl', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 9 {
		return Subroutine{}, fmt.Errorf("expected node with 9 leaf, got %d", len(node.GetChildren()))
	}

	routineType := SubroutineType(node.GetChildren()[0].GetValue())
	returnType := MainType(node.GetChildren()[1].GetValue())
	routineName := node.GetChildren()[2].GetValue()

	// All constructors must be named 'new', so we actively check for that
	if routineType == Constructor && routineName != "new" {
		return Subroutine{}, fmt.Errorf("constructor method must be named 'new', got '%s'", routineName)
	}

	// Iterate on the nested possible n declarations to extract all the variable names
	nested, arguments := node.GetChildren()[4].GetChildren(), []Variable{}
	for _, child := range nested {
		argType, argName := child.GetChildren()[0].GetValue(), child.GetChildren()[1].GetValue()

		// Primitive data types (int, string, bool) are handled differently than complex objects
		if builtin := MainType(argType); builtin == Int || builtin == String || builtin == Bool || builtin == Char {
			arguments = append(arguments, Variable{Name: argName, VarType: Parameter, DataType: DataType{Main: builtin}})
			continue
		}

		arguments = append(arguments, Variable{Name: argName, VarType: Parameter, DataType: DataType{Main: Object, Subtype: argType}})
	}

	nested, statements := node.GetChildren()[7].GetChildren(), []Statement{}
	for _, child := range nested {
		switch child.GetName() {
		case "sl_comment", "ml_comment": // Comment nodes in the AST are just skipped
			continue
		default:
			stmt, err := p.HandleStatement(child)
			if err != nil {
				return Subroutine{}, fmt.Errorf("failed to handle statement: %w", err)
			}
			statements = append(statements, stmt)
		}
	}

	return Subroutine{Name: routineName, Type: routineType, Return: DataType{Main: returnType}, Arguments: arguments, Statements: statements}, nil
}

// Generalized function to dispatch and convert between multiple statements types returning a 'jack.Statement'.
func (p *Parser) HandleStatement(node pc.Queryable) (Statement, error) {
	switch node.GetName() {
	case "do_stmt":
		stmt, err := p.HandleDoStmt(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'do' statement: %w", err)
		}
		return stmt, nil

	case "var_stmt":
		stmt, err := p.HandleVarStmt(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'var' statement: %w", err)
		}
		return stmt, nil

	case "let_stmt":
		stmt, err := p.HandleLetStmt(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'let' statement: %w", err)
		}
		return stmt, nil

	case "if_stmt":
		stmt, err := p.HandleIfStmt(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'if' statement: %w", err)
		}
		return stmt, nil

	case "while_stmt":
		stmt, err := p.HandleWhileStmt(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'while' statement: %w", err)
		}
		return stmt, nil

	case "return_stmt":
		stmt, err := p.HandleReturnStmt(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'do' statement: %w", err)
		}
		return stmt, nil

	default:
		return nil, fmt.Errorf("unrecognized node '%s' in statement", node.GetName())
	}
}

// Specialized function to convert a "do_stmt" node to a 'jack.DoStmt'.
func (p *Parser) HandleDoStmt(node pc.Queryable) (Statement, error) {
	if node.GetName() != "do_stmt" {
		return nil, fmt.Errorf("expected node 'do_stmt', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		return nil, fmt.Errorf("expected node with 3 leaf, got %d", len(node.GetChildren()))
	}

	expr, err := p.HandleFunCallExpr(node.GetChildren()[1])
	if err != nil {
		return nil, fmt.Errorf("failed to handle nested function call expression: %w", err)
	}

	return DoStmt{FuncCall: expr.(FuncCallExpr)}, nil
}

// Specialized function to convert a "var_stmt" node to a 'jack.VarStmt'.
func (p *Parser) HandleVarStmt(node pc.Queryable) (Statement, error) {
	if node.GetName() != "var_stmt" {
		return nil, fmt.Errorf("expected node 'var_stmt', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 4 {
		return nil, fmt.Errorf("expected node with 4 leaf, got %d", len(node.GetChildren()))
	}

	dataType := node.GetChildren()[1].GetValue()

	nested, variables := node.GetChildren()[2].GetChildren(), []Variable{}
	if len(nested) < 1 {
		return nil, fmt.Errorf("expected at least one variable declaration, got %d", len(nested))
	}

	// Iterate on the nested possible 'n' declarations to extract all the variable names
	for _, child := range nested {
		if child.GetName() != "IDENT" {
			return nil, fmt.Errorf("expected node 'IDENT', got %s", child.GetName())
		}
		// Primitive data types (int, string, bool) are handled differently than complex objects
		if builtin := MainType(dataType); builtin == Int || builtin == String || builtin == Bool || builtin == Char {
			variables = append(variables, Variable{Name: child.GetValue(), VarType: Local, DataType: DataType{Main: builtin}})
			continue
		}

		variables = append(variables, Variable{Name: child.GetValue(), VarType: Local, DataType: DataType{Main: Object, Subtype: dataType}})
	}

	return VarStmt{Vars: variables}, nil
}

// Specialized function to convert a "let_stmt" node to a 'jack.LetStmt'.
func (p *Parser) HandleLetStmt(node pc.Queryable) (Statement, error) {
	if node.GetName() != "let_stmt" {
		return nil, fmt.Errorf("expected node 'let_stmt', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 5 {
		return nil, fmt.Errorf("expected node with 5 leaf, got %d", len(node.GetChildren()))
	}

	lhs, err := p.HandleExpression(node.GetChildren()[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsh expression: %w", err)
	}
	_, isVarExpr := lhs.(VarExpr)
	_, isArrayExpr := lhs.(ArrayExpr)
	if !isVarExpr && !isArrayExpr { // Ensure 'lhs' is either 'ArrayExpr' or 'VarExpr'
		return nil, fmt.Errorf("lhs expression can only be 'VarExpr' or 'ArrayExpr', got %T", lhs)
	}

	rhs, err := p.HandleExpression(node.GetChildren()[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse right-hand side expression: %w", err)
	}

	return LetStmt{Lhs: lhs, Rhs: rhs}, nil
}

// Specialized function to convert a "if_stmt" node to a 'jack.IfStmt'.
func (p *Parser) HandleIfStmt(node pc.Queryable) (Statement, error) {
	if node.GetName() != "if_stmt" {
		return nil, fmt.Errorf("expected node 'if_stmt', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 8 {
		return nil, fmt.Errorf("expected node with 8 leaf, got %d", len(node.GetChildren()))
	}

	condition, err := p.HandleExpression(node.GetChildren()[2])
	if err != nil {
		return nil, fmt.Errorf("failed to handle nested if expression: %w", err)
	}

	nested, thenStmts := node.GetChildren()[5].GetChildren(), []Statement{}
	for _, child := range nested {
		switch child.GetName() {
		case "sl_comment", "ml_comment": // Comment nodes in the AST are just skipped
			continue
		default:
			stmt, err := p.HandleStatement(child)
			if err != nil {
				return IfStmt{}, fmt.Errorf("failed to handle statement in 'then' block: %w", err)
			}
			thenStmts = append(thenStmts, stmt)
		}
	}

	// The else section of the if statement is optional and can be omitted
	if node.GetChildren()[7].GetName() == "missing" {
		return IfStmt{Condition: condition, ThenBlock: thenStmts, ElseBlock: []Statement{}}, nil
	}

	nested, elseStmts := node.GetChildren()[7].GetChildren(), []Statement{}
	for _, child := range nested[2].GetChildren() {
		switch child.GetName() {
		case "sl_comment", "ml_comment": // Comment nodes in the AST are just skipped
			continue
		default:
			stmt, err := p.HandleStatement(child)
			if err != nil {
				return IfStmt{}, fmt.Errorf("failed to handle statement in 'else' block: %w", err)
			}
			elseStmts = append(elseStmts, stmt)
		}
	}

	return IfStmt{Condition: condition, ThenBlock: thenStmts, ElseBlock: elseStmts}, nil
}

// Specialized function to convert a "while_stmt" node to a 'jack.WhileStmt'.
func (p *Parser) HandleWhileStmt(node pc.Queryable) (Statement, error) {
	if node.GetName() != "while_stmt" {
		return nil, fmt.Errorf("expected node 'while_stmt', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 7 {
		return nil, fmt.Errorf("expected node with 7 leaf, got %d", len(node.GetChildren()))
	}

	condition, err := p.HandleExpression(node.GetChildren()[2])
	if err != nil {
		return nil, fmt.Errorf("failed to handle nested while expression: %w", err)
	}

	nested, statements := node.GetChildren()[5].GetChildren(), []Statement{}
	for _, child := range nested {
		switch child.GetName() {
		case "sl_comment", "ml_comment": // Comment nodes in the AST are just skipped
			continue
		default:
			stmt, err := p.HandleStatement(child)
			if err != nil {
				return WhileStmt{}, fmt.Errorf("failed to handle statement: %w", err)
			}
			statements = append(statements, stmt)
		}
	}

	return WhileStmt{Condition: condition, Block: statements}, nil
}

// Specialized function to convert a "return_stmt" node to a 'jack.ReturnStmt'.
func (p *Parser) HandleReturnStmt(node pc.Queryable) (Statement, error) {
	if node.GetName() != "return_stmt" {
		return nil, fmt.Errorf("expected node 'return_stmt', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		return nil, fmt.Errorf("expected node with 3 leaf, got %d", len(node.GetChildren()))
	}

	// The return value/expression can be omitted (for example if the return type is void)
	if node.GetChildren()[1].GetName() == "missing" {
		return ReturnStmt{Expr: nil}, nil
	}

	expr, err := p.HandleExpression(node.GetChildren()[1])
	if err != nil {
		return nil, fmt.Errorf("failed to handle nested expression: %w", err)
	}

	return ReturnStmt{Expr: expr}, nil
}

// Generalized function to dispatch and convert between multiple expression types returning a 'jack.Expression'.
func (p *Parser) HandleExpression(node pc.Queryable) (Expression, error) {
	switch node.GetName() {
	case "array_expr":
		expr, err := p.HandleArrayExpr(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'array' expression: %w", err)
		}
		return expr, nil

	case "unary_expr":
		expr, err := p.HandleUnaryExpr(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'unary' expression: %w", err)
		}
		return expr, nil

	case "binary_expr":
		expr, err := p.HandleBinaryExpr(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'binary' expression: %w", err)
		}
		return expr, nil

	case "funcall_expr":
		stmt, err := p.HandleFunCallExpr(node)
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'funcall' expression: %w", err)
		}
		return stmt, nil

	case "subexpr":
		stmt, err := p.HandleExpression(node.GetChildren()[1])
		if err != nil {
			return nil, fmt.Errorf("failed to handle 'nested' expression: %w", err)
		}
		return stmt, nil

	case "IDENT":
		return VarExpr{Var: node.GetValue()}, nil
	case "THIS":
		return VarExpr{Var: "this"}, nil

	case "INT":
		return LiteralExpr{Type: DataType{Main: Int}, Value: node.GetValue()}, nil
	case "CHAR":
		return LiteralExpr{Type: DataType{Main: Char}, Value: node.GetValue()}, nil
	case "TRUE", "FALSE":
		return LiteralExpr{Type: DataType{Main: Bool}, Value: node.GetValue()}, nil
	case "STRING":
		return LiteralExpr{Type: DataType{Main: String}, Value: strings.Trim(node.GetValue(), `"`)}, nil
	case "NULL":
		return LiteralExpr{Type: DataType{Main: Object}, Value: node.GetValue()}, nil

	default:
		return nil, fmt.Errorf("unrecognized node '%s' in expression", node.GetName())
	}
}

// Specialized function to convert a "array_expr" node to a 'jack.ArrayExpr'.
func (p *Parser) HandleArrayExpr(node pc.Queryable) (Expression, error) {
	if node.GetName() != "array_expr" {
		return nil, fmt.Errorf("expected node 'array_expr', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 4 {
		return nil, fmt.Errorf("expected node with 4 leaf, got %d", len(node.GetChildren()))
	}

	array := node.GetChildren()[0].GetValue()

	expr, err := p.HandleExpression(node.GetChildren()[2])
	if err != nil {
		return nil, fmt.Errorf("failed to handle nested array index expression: %w", err)
	}

	return ArrayExpr{Var: array, Index: expr}, nil
}

// Specialized function to convert a "unary_expr" node to a 'jack.UnaryExpr'.
func (p *Parser) HandleUnaryExpr(node pc.Queryable) (Expression, error) {
	if node.GetName() != "unary_expr" {
		return nil, fmt.Errorf("expected node 'unary_expr', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 2 {
		return nil, fmt.Errorf("expected node with 2 leaf, got %d", len(node.GetChildren()))
	}

	exprType := ExprType(strings.ToLower((node.GetChildren()[0].GetName())))

	rhs, err := p.HandleExpression(node.GetChildren()[1])
	if err != nil {
		return nil, fmt.Errorf("failed to handle left-hand side expression: %w", err)
	}

	return UnaryExpr{Type: exprType, Rhs: rhs}, nil
}

// Specialized function to convert a "binary_expr" node to a 'jack.BinaryExpr'.
func (p *Parser) HandleBinaryExpr(node pc.Queryable) (Expression, error) {
	if node.GetName() != "binary_expr" {
		return nil, fmt.Errorf("expected node 'binary_expr', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		return nil, fmt.Errorf("expected node with 3 leaf, got %d", len(node.GetChildren()))
	}
	lhs, err := p.HandleExpression(node.GetChildren()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to handle left-hand side expression: %w", err)
	}

	exprType := ExprType(strings.ToLower((node.GetChildren()[1].GetName())))

	rhs, err := p.HandleExpression(node.GetChildren()[2])
	if err != nil {
		return nil, fmt.Errorf("failed to handle right-hand side expression: %w", err)
	}

	return BinaryExpr{Type: exprType, Lhs: lhs, Rhs: rhs}, nil
}

// Specialized function to convert a "funcall_expr" node to a 'jack.FuncCallExpr'.
func (p *Parser) HandleFunCallExpr(node pc.Queryable) (Expression, error) {
	if node.GetName() != "funcall_expr" {
		return nil, fmt.Errorf("expected node 'funcall_expr', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 4 {
		return nil, fmt.Errorf("expected node with 4 leaf, got %d", len(node.GetChildren()))
	}

	nested := node.GetChildren()[0].GetChildren()
	external, class, method := len(nested) > 1, "", ""
	if external {
		class, method = nested[0].GetValue(), nested[1].GetValue()
	} else {
		class, method = "", nested[0].GetValue()
	}

	nested, arguments := node.GetChildren()[2].GetChildren(), []Expression{}
	for _, child := range nested {
		arg, err := p.HandleExpression(child)
		if err != nil {
			return nil, fmt.Errorf("failed to handle nested argument expression: %w", err)
		}
		arguments = append(arguments, arg)
	}

	return FuncCallExpr{IsExtCall: external, Var: class, FuncName: method, Arguments: arguments}, nil
}
