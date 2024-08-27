package jack

import (
	"errors"
	"fmt"
	"io"
	"os"

	pc "github.com/prataprc/goparsec"
)

var ast = pc.NewAST("jack_program", 0)

var (
	pClass = ast.And("class_decl", nil,
		ast.Kleene("file_header", nil, pComment),
		pc.Atom("class", "CLASS"), pIdent, pLBrace,
		// TODO (hmny): Add fields parser with support for single and multiline comments
		// ?  ast.Kleene("fields_or_comments", nil, ast.OrdChoice("items", nil, pField, pComment)),
		ast.Kleene("methods_or_comments", nil, ast.OrdChoice("items", nil, pMethod, pComment)),
		pRBrace,
	)

	pMethod = ast.And("method_decl", nil,
		// Func keyword, return type and function/method name
		pc.Atom("function", "FUNC"), pDataType, pIdent,
		// '(', comma separated argument type(s) and name(s), ')'
		pLParen, ast.Kleene("arguments", nil, ast.And("argument", nil, pDataType, pIdent), pComma), pRParen,
		// '{', statement and or comments (s), '}'
		pLBrace, ast.Kleene("statements_or_comments", nil, ast.OrdChoice("item", nil, pStatement, pComment)), pRBrace,
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
	pStatement = ast.And("statement", nil, ast.OrdChoice("item", nil, pDoStmt, pReturnStmt), pSemi)

	pDoStmt = ast.And("do_stmt", nil,
		// Support both external method call and local method call syntax:
		// - 'External': call to another class method (e.g. 'do X.ExtMethod()')
		// - 'Local': call to same class/instance method (e.g. 'do InternalMethod()')
		pc.Atom("do", "DO"), ast.Many("qualifiers", nil, pIdent, pDot),
		// '(', comma separated argument passing w/ expression to be eval'd, ')'
		pLParen, ast.Kleene("args", nil, pExpr, pComma), pRParen,
	)

	pReturnStmt = ast.And("return_stmt", nil, pc.Atom("return", "RETURN"), pc.Maybe(nil, pExpr))
)

var (
	// ! The order of this PCs is important: by putting Int() before Float() we'll not be able to parse a float
	// !completely because the integer part will be picked up by the Int() PC before given back control to PExpr.
	pExpr    = ast.OrdChoice("expression", nil, pLiteral)
	pLiteral = ast.OrdChoice("literal", nil,
		// Numeric literals (int and float) as well as string literals
		pc.Float(), pc.Int(), pc.Token(`"(?:\\.|[^"\\])*"`, "STRING"),
		// also we cover in this way boolean literal declaration (true | false) and null
		pc.Token("true", "TRUE"), pc.Token("false", "FALSE"),
		pc.Token("null", "NULL"), // TODO (hmny): Should we also add char literal PC
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
	pLBrace = pc.Atom("{", "LBRACE")
	pRBrace = pc.Atom("}", "RBRACE")
	pLParen = pc.Atom("(", "RPAREN")
	pRParen = pc.Atom(")", "RPAREN")

	// Available memory operation type (only push and pop since it's stack based)
	pDataType = ast.OrdChoice("data_type", nil,
		pc.Atom("int", "INT"), pc.Atom("char", "CHAR"), pc.Atom("bool", "BOOL"),
		pc.Atom("null", "NULL"), pc.Atom("void", "VOID"), pIdent,
	)
)

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

	// ! return p.FromAST(ast)

	fmt.Println(ast) // TODO Remove
	return Class{}, errors.New("Parser.Parse not implemented yet")
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

		file.Write([]byte(ast.Dotstring("\"VM AST\"")))
	}

	// Feature flag: Enables pretty printing of the AST on the console
	if os.Getenv("PRINT_AST") != "" {
		ast.Prettyprint()
	}
	// TODO (hmny): This hardcoding to true should be changed
	return root, true // Success is based on the reaching of 'EOF'
}
