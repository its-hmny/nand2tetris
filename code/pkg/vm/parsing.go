package vm

import (
	"fmt"
	"io"
	"os"

	pc "github.com/prataprc/goparsec"
)

var ast = pc.NewAST("virtual_machine", 0)

var (
	// Parser combinator for a VM module/class, in the nand2tetris VM there's a Java like
	// behavior where a program is composed of multiple '.vm' file ('.class' in Java) where
	// each contains the bytecode for the specific module/class (a separate translation unit).
	pModule = ast.ManyUntil("module", nil, ast.OrdChoice("node", nil, pComment, pStatement), pc.End())

	// Parser combinator for comments in Assembler program
	pComment = ast.And("comment", nil, pc.Atom("//", "//"), pc.Token(`(?m).*$`, "COMMENT"))
	// Parser combinator for a generic VM statement (MemoryOp, ArithmeticOp, ...)
	pStatement = ast.OrdChoice("statement", nil,
		pMemoryOp, pArithmeticOp,
		pUncondJumpOp, pCondJumpOp,
		pLabelDecl, pFuncDecl,
		pReturnOp, pFunCallOp,
	)

	// Memory operation, compliant with the following syntax: "{push|pop} {segment} {index}"
	pMemoryOp = ast.And("memory_op", nil, pMemOpType, pSegment, pc.Int())
	// Arithmetic operation, could either be binary or unary (modifies only the Stack Pointer)
	pArithmeticOp = ast.And("arithmetic_op", nil, ast.OrdChoice("operations", nil,
		// Comparison operations available on the VM bytecode
		pc.Atom("eq", "EQ"), pc.Atom("gt", "GT"), pc.Atom("lt", "LT"),
		// Arithmetic operations available on the VM bytecode
		pc.Atom("add", "ADD"), pc.Atom("sub", "SUB"), pc.Atom("neg", "NEG"),
		// Bit-a-bit operations available on the VM bytecode
		pc.Atom("not", "NOT"), pc.Atom("and", "AND"), pc.Atom("or", "OR"),
	))

	// Unconditional jump operation, compliant with the following syntax: "goto {symbol}"
	pUncondJumpOp = ast.And("goto_op", nil, pc.Atom("goto", "GOTO"), pLabel)
	// Conditional jump operation, compliant with the following syntax: "if goto {symbol}"
	pCondJumpOp = ast.And("if-goto_op", nil, pc.Atom("if-goto", "IF-GOTO"), pLabel)

	// Label declaration, compliant with the following syntax: "label {symbol}"s
	pLabelDecl = ast.And("label_decl", nil, pc.Atom("label", "LABEL"), pLabel)
	// Function declaration, compliant with the following syntax: "function {name} {n_args}"
	pFuncDecl = ast.And("func_decl", nil, pc.Atom("function", "FUNC"), pFuncName, pc.Int())

	// Return operation, compliant with the following syntax: "return"
	pReturnOp = ast.And("returns", nil, pc.Atom("return", "RETURN"))
	// Function calling operation, compliant with the following syntax: "call {name} {n_args}"
	pFunCallOp = ast.And("func_call", nil, pc.Atom("call", "CALL"), pFuncName, pc.Int())
)

var (
	pLabel     = pc.Ident()
	pFuncName  = pc.Ident()
	pMemOpType = ast.OrdChoice("mem_op_type", nil, pc.Atom("push", "PUSH"), pc.Atom("pop", "POP"))
	pSegment   = ast.OrdChoice("mem_segment", nil,
		pc.Atom("argument", "ARGUMENT"), pc.Atom("local", "LOCAL"),
		pc.Atom("static", "STATIC"), pc.Atom("constant", "CONSTANT"),
		pc.Atom("this", "THIS"), pc.Atom("that", "THAT"),
		pc.Atom("temp", "TEMP"), pc.Atom("pointer", "POINTER"),
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
