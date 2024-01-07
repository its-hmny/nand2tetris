package assembler

import (
	"io"
	"os"

	pc "github.com/prataprc/goparsec"
)

// Top level object, will generate the traversable AST based on the input plus the PCs below.
var ast = pc.NewAST("assembler", 0)

var (
	// Parser combinator for an entire Assembler program
	pProgram = ast.ManyUntil("program", nil, ast.OrdChoice("item", nil, pComment, pInst), pc.End())
	// Parser combinator for a generic Assembler instruction (either C, A or Label declaration)
	pInst = ast.OrdChoice("inst", nil, pAInst, pCInst, pLabelDecl)
	// Parser combinator for comments in Assembler program
	pComment = ast.And("comment", nil, pc.Atom("//", "//"), pc.Token(`(?m).*$`, "COMMENT"))

	// Parser combinator for A Instructions
	pAInst = ast.And("a-inst", nil, pc.Atom("@", "@"), pLabel)
	// Parser combinator for new label declaration
	pLabelDecl = ast.And("label-dcl", nil, pc.Atom("(", "("), pLabel, pc.Atom(")", ")"))
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
	pDest = ast.OrdChoice("dest", nil, pc.Atom("D", "D"), pc.Atom("A", "A"), pc.Atom("M", "M"))

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
		pc.Atom("JGT", "JGT"), pc.Atom("JEQ", "JEQ"), pc.Atom("JGE", "JGE"),
		pc.Atom("JLT", "JLT"), pc.Atom("JLE", "JLE"), pc.Atom("JMP", "JMP"),
	)
)

type Parser struct{}

func NewParser() Parser { return Parser{} }

func (p *Parser) Parse(r io.Reader) (pc.Queryable, bool) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, false
	}

	// Feature flag: Enable 'goparsec' library's debug logs
	if os.Getenv("PARSEC_DEBUG") != "" {
		ast.SetDebug()
	}

	// We generate the traversable Abstract Syntax Tree from the source content
	root, scanner := ast.Parsewith(pProgram, pc.NewScanner(content))

	// Feature flag: Enables export of the AST as Dot file (debug.ast.fot)
	if os.Getenv("EXPORT_AST") != "" {
		file, _ := os.Create("debug/debug.ast.dot")
		defer file.Close()

		file.Write([]byte(ast.Dotstring("\"Assembler AST\"")))
	}

	// Feature flag: Enables pretty printing of the AST on the console
	if os.Getenv("PRINT_AST") != "" {
		ast.Prettyprint()
	}

	return root, scanner.Endof() // Success is based on the reaching of 'EOF'
}
