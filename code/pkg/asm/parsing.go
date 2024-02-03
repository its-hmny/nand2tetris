package asm

import (
	"fmt"
	"io"
	"os"

	pc "github.com/prataprc/goparsec"
)

// ----------------------------------------------------------------------------
// Parser Combinator(s)

// This section defines the Parser Combinator for every token & instruction of the Asm language.
//
// Each parser combinator either manages a instruction (A Inst, C Inst, Label Decl) or some pieces
// of it: namely tokens and identifiers. Also we manage comments inside the codebase that can
// either present themselves at the beginning of the line or in the middle.

// Top level object, will generate the traversable AST based on the input plus the PCs below.
var ast = pc.NewAST("assembler", 0)

var (
	// Parser combinator for an entire Assembler program (a sequence of comments and instructions)
	pProgram = ast.ManyUntil("program", nil, ast.OrdChoice("item", nil, pComment, pInstruction), pc.End())

	// Parser combinator for a generic Assembler instruction (either C, A or Label declaration)
	pInstruction = ast.OrdChoice("instruction", nil, pAInst, pCInst, pLabelDecl)
	// Parser combinator for comments in Assembler program
	pComment = ast.And("comment", nil, pc.Atom("//", "//"), pc.Token(`(?m).*$`, "COMMENT"))

	// Parser combinator for A Instructions
	pAInst = ast.And("a-inst", nil, pc.Atom("@", "@"), pLabel)
	// Parser combinator for new label declaration
	pLabelDecl = ast.And("label-decl", nil, pc.Atom("(", "("), pLabel, pc.Atom(")", ")"))
	// Parser combinator for C Instructions
	pCInst = ast.And("c-inst", nil,
		ast.Maybe("maybe-assign", nil, ast.And("assign", nil, pDest, pc.Atom("=", "="))),
		pComp, // 'comp' should always be provided
		ast.Maybe("maybe-goto", nil, ast.And("goto", nil, pc.Atom(";", ";"), pJump)),
	)
)

var (
	// Generic label parser (A Instruction + Label declaration)
	// NOTE: A label can be any sequence of letters, digits, and symbols (_, ., $, :).
	// NOTE: A label cannot begin with a leading digit (a symbol is indeed allowed).
	pLabel = ast.OrdChoice("label", nil, pc.Int(), pc.Token(`[A-Za-z_.$:][0-9a-zA-Z_.$:]*`, "SYMBOL"))

	// Generic destination parser (C Instruction subsection)
	// NOTE: The order of the Atom is reversed w.r.t. the one provided in the translation table cause
	// if not the single destination section will match before in the PC (BFS Search algorithm)
	pDest = ast.OrdChoice("dest", nil,
		pc.Atom("AM", "AM"), pc.Atom("AD", "AD"), pc.Atom("MD", "MD"),
		pc.Atom("D", "D"), pc.Atom("A", "A"), pc.Atom("M", "M"),
	)

	// Generic computation parser (C Instruction subsection)
	// NOTE: The order of the Atom is reversed w.r.t. the one provided in the translation table cause
	// if not the 'Constant and identifiers' part will match before the order (BFS Search algorithm)
	pComp = ast.OrdChoice("comp", nil,
		// - Bitwise register with register operations
		pc.Atom("D&A", "D&A"), pc.Atom("D&M", "D&M"),
		pc.Atom("D|A", "D|A"), pc.Atom("D|M", "D|M"),
		// - Register with register operations
		pc.Atom("D+A", "D+A"), pc.Atom("D+M", "D+M"),
		pc.Atom("D-A", "D-A"), pc.Atom("D-M", "D-M"),
		pc.Atom("A-D", "A-D"), pc.Atom("M-D", "M-D"),
		// - Increment and decrement operations
		pc.Atom("D+1", "D+1"), pc.Atom("A+1", "A+1"), pc.Atom("M+1", "M+1"),
		pc.Atom("D-1", "D-1"), pc.Atom("A-1", "A-1"), pc.Atom("M-1", "M-1"),
		// - Binary and numerical negations
		pc.Atom("!D", "!D"), pc.Atom("!A", "!A"), pc.Atom("!M", "!M"),
		pc.Atom("-D", "-D"), pc.Atom("-A", "-A"), pc.Atom("-M", "-M"),
		// - Constants and identities
		pc.Atom("0", "0"), pc.Atom("1", "1"), pc.Atom("-1", "-1"),
		pc.Atom("D", "D"), pc.Atom("A", "A"), pc.Atom("M", "M"),
	)

	// Generic jump parser (C Instruction subsection)
	pJump = ast.OrdChoice("jump", nil,
		pc.Atom("JNE", "JNE"), pc.Atom("JEQ", "JEQ"),
		pc.Atom("JGT", "JGT"), pc.Atom("JGE", "JGE"),
		pc.Atom("JLT", "JLT"), pc.Atom("JLE", "JLE"),
		pc.Atom("JMP", "JMP"),
	)
)

// ----------------------------------------------------------------------------
// Asm Parser

// This section defines the Parser for the nand2tetris Asm language.
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

// Scans the textual input stream coming from the 'reader' method and returns a traversable
// Abstract Syntax Tree (AST) that can be eventually used down the line for other compilation
// steps such as lowering, optimizations (dead branch removal, loop unrolling and so on...)
func (p *Parser) Parse() (pc.Queryable, bool) {
	content, err := io.ReadAll(p.reader)
	if err != nil {
		return nil, false
	}

	// Feature flag: Enable 'goparsec' library's debug logs
	if os.Getenv("PARSEC_DEBUG") != "" {
		ast.SetDebug()
	}

	// We generate the traversable Abstract Syntax Tree from the source content
	root, _ := ast.Parsewith(pProgram, pc.NewScanner(content))

	// Feature flag: Enables export of the AST as Dot file (debug.ast.fot)
	if os.Getenv("EXPORT_AST") != "" {
		file, _ := os.Create(fmt.Sprintf("%s/debug.ast.dot", os.Getenv("DEBUG_FOLDER")))
		defer file.Close()

		file.Write([]byte(ast.Dotstring("\"Assembler AST\"")))
	}

	// Feature flag: Enables pretty printing of the AST on the console
	if os.Getenv("PRINT_AST") != "" {
		ast.Prettyprint()
	}
	// TODO (hmny): This hardcoding to true should be changed
	return root, true // Success is based on the reaching of 'EOF'
}
