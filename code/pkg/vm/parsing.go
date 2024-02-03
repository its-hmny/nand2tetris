package vm

import (
	"fmt"
	"io"
	"os"

	pc "github.com/prataprc/goparsec"
)

// ----------------------------------------------------------------------------
// Parser Combinator(s)

// This section defines the Parser Combinator for every token & instruction of the Vm language.
//
// Each parser combinator either manages an operation (MemoryOp, ArithmeticOp, ...) or some pieces
// of it: namely tokens and identifiers. Also we manage comments inside the codebase that can
// either present themselves at the beginning of the line or in the middle.

// Top level object, will generate the traversable AST based on the input plus the PCs below.
var ast = pc.NewAST("virtual_machine", 0)

var (
	// Parser combinator for a VM module/class, in the nand2tetris VM there's a Java like
	// behavior where a program is composed of multiple '.vm' file ('.class' in Java) where
	// each contains the bytecode for the specific module/class (a separate translation unit).
	pModule = ast.ManyUntil("module", nil, ast.OrdChoice("node", nil, pComment, pOperation), pc.End())

	// Parser combinator for comments in Assembler program
	pComment = ast.And("comment", nil, pc.Atom("//", "//"), pc.Token(`(?m).*$`, "COMMENT"))
	// Parser combinator for a generic VM operation (MemoryOp, ArithmeticOp, ...)
	pOperation = ast.OrdChoice("operation", nil,
		// Stack operation and label/function declaration
		pMemoryOp, pArithmeticOp, pLabelDecl, pFuncDecl,
		// Jump operation of multiple sorts
		pReturnOp, pFunCallOp, pUncondJumpOp, pCondJumpOp,
	)

	// Memory operation, compliant with the following syntax: "{push|pop} {segment} {index}"
	pMemoryOp = ast.And("memory_op", nil, pMemOpType, pSegment, pc.Int())
	// Arithmetic operation, could either be binary or unary (modifies only the Stack Pointer)
	pArithmeticOp = ast.And("arithmetic_op", nil, pArithOpType)

	// Label declaration, compliant with the following syntax: "label {symbol}"s
	pLabelDecl = ast.And("label_decl", nil, pc.Atom("label", "LABEL"), pLabel)
	// Function declaration, compliant with the following syntax: "function {name} {n_args}"
	pFuncDecl = ast.And("func_decl", nil, pc.Atom("function", "FUNC"), pFuncName, pc.Int())

	// Unconditional jump operation, compliant with the following syntax: "goto {symbol}"
	pUncondJumpOp = ast.And("goto_op", nil, pc.Atom("goto", "GOTO"), pLabel)
	// Conditional jump operation, compliant with the following syntax: "if goto {symbol}"
	pCondJumpOp = ast.And("if-goto_op", nil, pc.Atom("if-goto", "IF-GOTO"), pLabel)

	// Return operation, compliant with the following syntax: "return"
	pReturnOp = ast.And("returns", nil, pc.Atom("return", "RETURN"))
	// Function calling operation, compliant with the following syntax: "call {name} {n_args}"
	pFunCallOp = ast.And("func_call", nil, pc.Atom("call", "CALL"), pFuncName, pc.Int())
)

var (
	pLabel    = pc.Ident() // Label names follow the same pattern as identifiers in other languages
	pFuncName = pc.Ident() // Function names follow the same pattern as identifiers in other languages

	// Available memory operation type (only push and pop since it's stack based)
	pMemOpType = ast.OrdChoice("mem_op_type", nil, pc.Atom("push", "PUSH"), pc.Atom("pop", "POP"))
	// Available heap segments (they act as registers and are used alongside the stack)
	pSegment = ast.OrdChoice("mem_segment", nil,
		pc.Atom("argument", "ARGUMENT"), pc.Atom("local", "LOCAL"),
		pc.Atom("static", "STATIC"), pc.Atom("constant", "CONSTANT"),
		pc.Atom("this", "THIS"), pc.Atom("that", "THAT"),
		pc.Atom("temp", "TEMP"), pc.Atom("pointer", "POINTER"),
	)

	// Available arithmetic operation types (more functionality will be provided in the next phases)
	pArithOpType = ast.OrdChoice("operations", nil,
		// Comparison operations available on the VM bytecode
		pc.Atom("eq", "EQ"), pc.Atom("gt", "GT"), pc.Atom("lt", "LT"),
		// Arithmetic operations available on the VM bytecode
		pc.Atom("add", "ADD"), pc.Atom("sub", "SUB"), pc.Atom("neg", "NEG"),
		// Bit-a-bit operations available on the VM bytecode
		pc.Atom("not", "NOT"), pc.Atom("and", "AND"), pc.Atom("or", "OR"),
	)
)

// ----------------------------------------------------------------------------
// Vm Parser

// This section defines the Parser for the nand2tetris Vm language.
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
	root, _ := ast.Parsewith(pModule, pc.NewScanner(content))

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
