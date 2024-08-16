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
		pc.Atom("class", "CLASS"), pIdent, pLBrace,
		// ast.Kleene("fields_decls", nil, nil),     // TODO (hmny): Add fields parser
		ast.Kleene("method_decls", nil, pMethod), // TODO (hmny): Add method parser
		pRBrace,
	)

	pMethod = ast.And("method_decl", nil,
		pc.Atom("function", "FUNC"), pDataType, pIdent,
		pLParen, ast.Kleene("arguments", nil, pc.And(nil, pDataType, pIdent)), pRParen,
		pLBrace, ast.Kleene("statements", nil, pStatement), pRBrace,
	)
)

var (
	pStatement = ast.OrdChoice("statement", nil, pDoStmt, pReturnStmt)

	pDoStmt = ast.And("do_stmt", nil, pc.Atom("do", "DO"),
		// TODO (hmny): For now all the full qualifier is jammed into a single Ident
		// ! - ext call to another class method (e.g. 'do X.ExtMethod();')
		// ! - local call to same class method (e.g. 'do InternalMethod();')
		pIdent, pLParen,
		ast.Kleene("args", nil, pExpr, pc.Kleene(nil, pc.And(nil, pComma, pExpr))),
		pRParen, pSemi,
	)

	pReturnStmt = ast.And("return_stmt", nil,
		pc.Atom("return", "RETURN"), pc.Maybe(nil, pExpr), pSemi,
	)
)

var (
	// TODO (hmny): 'pc.String()' doesn't seem to work properly, will need my own
	pExpr = ast.OrdChoice("expression", nil, pc.Int(), pc.Float())
)

var (
	// Generic Identifier parser (for label and function declaration)
	// NOTE: An ident can be any sequence of letters, digits, and symbols (_, ., $, :).
	// NOTE: An ident cannot begin with a leading digit (a symbol is indeed allowed).
	pIdent = pc.Token(`[A-Za-z_.$:][0-9a-zA-Z_.$:]*`, "IDENT")

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
// It uses parser combinators to obtain the AST from the source code (the latter can be provided)
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
