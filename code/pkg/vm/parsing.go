package vm

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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
		// Stack operation + label and jump operations
		pMemoryOp, pArithmeticOp, pLabelDecl, pGotoOp,
		// Function related operations and statements
		pFuncDecl, pFunCallOp, pReturnOp,
	)

	// Memory operation, compliant with the following syntax: "{push|pop} {segment} {index}"
	pMemoryOp = ast.And("memory_op", nil, pMemOpType, pSegment, pc.Int())
	// Arithmetic operation, could either be binary or unary (modifies only the Stack Pointer)
	pArithmeticOp = ast.And("arithmetic_op", nil, pArithOpType)

	// Label declaration, compliant with the following syntax: "label {symbol}"s
	pLabelDecl = ast.And("label_decl", nil, pc.Atom("label", "LABEL"), pIdent)
	// Jump operation, compliant with the following syntax: "{if-goto|goto} {symbol}"
	pGotoOp = ast.And("goto_op", nil, pJumpType, pIdent)

	// Function declaration, compliant with the following syntax: "function {name} {n_args}"
	pFuncDecl = ast.And("func_decl", nil, pc.Atom("function", "FUNC"), pIdent, pc.Int())
	// Function call operation, compliant with the following syntax: "call {name} {n_args}"
	pFunCallOp = ast.And("func_call", nil, pc.Atom("call", "CALL"), pIdent, pc.Int())
	// Return operation, compliant with the following syntax: "return"
	pReturnOp = ast.And("return_op", nil, pc.Atom("return", "RETURN"))
)

var (
	// Generic Identifier parser (for label and function declaration)
	// NOTE: An ident can be any sequence of letters, digits, and symbols (_, ., $, :).
	// NOTE: An ident cannot begin with a leading digit (a symbol is indeed allowed).
	pIdent = pc.Token(`[A-Za-z_.$:][0-9a-zA-Z_.$:]*`, "IDENT")

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

	// Jump types can either be conditional (if-goto) or unconditional (goto).
	pJumpType = ast.OrdChoice("jump_type", nil, pc.Atom("goto", "GOTO"), pc.Atom("if-goto", "IF-GOTO"))
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

// Parser entrypoint divides the 2 phases of the parsing pipeline
// Text --> AST: This step is done using PCs and returns a generic traversable AST
// AST --> IR: This step is done by traversing the AST and extracting the 'vm.Module'
func (p *Parser) Parse() (Module, error) {
	content, err := io.ReadAll(p.reader)
	if err != nil {
		return nil, fmt.Errorf("cannot read from 'io.Reader': %s", err)
	}

	ast, success := p.FromSource(content)
	if !success {
		return nil, fmt.Errorf("failed to parse AST from input content")
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
	root, _ := ast.Parsewith(pModule, pc.NewScanner(source))

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

// This function takes the root node of the raw parsed AST and does a DFS on it parsing
// one by one each subtree and retuning a 'vm.Module' that can be used as in-memory and
// type-safe AST not dependent on the parsing library used.
func (p *Parser) FromAST(root pc.Queryable) (Module, error) {
	module := []Operation{}

	if root.GetName() != "module" {
		return nil, fmt.Errorf("expected node 'program', found %s", root.GetName())
	}

	for _, child := range root.GetChildren() {
		switch child.GetName() {
		case "memory_op": // Memory operation subtree, appends 'vm.MemoryOp' to 'modules'
			op, err := p.HandleMemoryOp(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "arithmetic_op": // Arithmetic operation subtree, appends 'vm.ArithmeticOp' to 'modules'
			op, err := p.HandleArithmeticOp(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "label_decl": // Label declaration subtree, appends 'vm.LabelDeclaration' to 'modules'
			op, err := p.HandleLabelDecl(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "goto_op": // Goto operation subtree, appends 'vm.GotoOp' to 'modules'
			op, err := p.HandleGotoOp(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "func_decl": // Function declaration subtree, appends 'vm.FuncDecl' to 'modules'
			op, err := p.HandleFuncDecl(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "return_op": // Return operation subtree, appends 'vm.ReturnOp' to 'modules'
			op, err := p.HandleReturnOp(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "func_call": // Function call operation subtree, appends 'vm.FuncCallOp' to 'modules'
			op, err := p.HandleFuncCall(child)
			if op == nil || err != nil {
				return nil, err
			}
			module = append(module, op)

		case "comment": // Comment nodes in the AST are just skipped
			continue

		default: // Error case, unrecognized subtree in the AST
			return nil, fmt.Errorf("unrecognized node '%s'", child.GetName())
		}
	}

	return module, nil
}

// Specialized function to convert a "memory_op" node to a 'vm.MemoryOp'.
func (Parser) HandleMemoryOp(node pc.Queryable) (Operation, error) {
	if node.GetName() != "memory_op" {
		return nil, fmt.Errorf("expected node 'memory_op', got %s", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		return nil, fmt.Errorf("expected node with 3 leaf, got %d", len(node.GetChildren()))
	}

	operation := OperationType(node.GetChildren()[0].GetValue())
	segment := SegmentType(node.GetChildren()[1].GetValue())
	offset, err := strconv.ParseUint(node.GetChildren()[2].GetValue(), 10, 16)
	if err != nil {
		log.Fatalf("failed to parse 'offset' in MemoryOp, got '%s'", node.GetChildren()[2].GetValue())
	}

	return MemoryOp{Operation: operation, Segment: segment, Offset: uint16(offset)}, nil
}

// Specialized function to convert a "arithmetic_op" node to a 'vm.ArithmeticOp'.
func (Parser) HandleArithmeticOp(node pc.Queryable) (Operation, error) {
	if node.GetName() != "arithmetic_op" {
		log.Fatalf("expected node 'arithmetic_op', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 1 {
		log.Fatalf("expected node 'arithmetic_op' with 1 leaf, got %d", len(node.GetChildren()))
	}

	return ArithmeticOp{Operation: ArithOpType(node.GetChildren()[0].GetValue())}, nil
}

// Specialized function to convert a "label_decl" node to a 'vm.LabelDeclaration'.
func (Parser) HandleLabelDecl(node pc.Queryable) (Operation, error) {
	if node.GetName() != "label_decl" {
		log.Fatalf("expected node 'label_decl', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 2 {
		log.Fatalf("expected node 'label_decl' with 2 leaf, got %d", len(node.GetChildren()))
	}

	return LabelDeclaration{Name: node.GetChildren()[1].GetValue()}, nil
}

// Specialized function to convert a "goto_op" node to a 'vm.GotoOp'.
func (Parser) HandleGotoOp(node pc.Queryable) (Operation, error) {
	if node.GetName() != "goto_op" {
		log.Fatalf("expected node 'goto_op', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 2 {
		log.Fatalf("expected node 'goto_op' with 2 leaf, got %d", len(node.GetChildren()))
	}

	jump := JumpType(node.GetChildren()[0].GetValue())
	label := node.GetChildren()[1].GetValue()

	return GotoOp{Jump: jump, Label: label}, nil
}

// Specialized function to convert a "func_decl" node to a 'vm.FuncDecl'.
func (Parser) HandleFuncDecl(node pc.Queryable) (Operation, error) {
	if node.GetName() != "func_decl" {
		log.Fatalf("expected node 'func_decl', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		log.Fatalf("expected node 'func_decl' with 3 leaf, got %d", len(node.GetChildren()))
	}

	name := node.GetChildren()[1].GetValue()
	args, err := strconv.ParseUint(node.GetChildren()[2].GetValue(), 10, 8)
	if err != nil {
		log.Fatalf("failed to parse 'args' in FuncDecl, got '%s'", node.GetChildren()[2].GetValue())
	}

	return FuncDecl{Name: name, ArgsNum: uint8(args)}, nil
}

// Specialized function to convert a "return_op" node to a 'vm.ReturnOp'.
func (Parser) HandleReturnOp(node pc.Queryable) (Operation, error) {
	if node.GetName() != "return_op" {
		log.Fatalf("expected node 'return_op', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 1 {
		log.Fatalf("expected node 'return_op' with 1 leaf, got %d", len(node.GetChildren()))
	}

	return ReturnOp{}, nil
}

// Specialized function to convert a "func_call" node to a 'vm.FuncCallOp'.
func (Parser) HandleFuncCall(node pc.Queryable) (Operation, error) {
	if node.GetName() != "func_call" {
		log.Fatalf("expected node 'func_call', got %s ", node.GetName())
	}
	if len(node.GetChildren()) != 3 {
		log.Fatalf("expected node 'func_call' with 3 leaf, got %d", len(node.GetChildren()))
	}

	name := node.GetChildren()[1].GetValue()
	args, err := strconv.ParseUint(node.GetChildren()[2].GetValue(), 10, 8)
	if err != nil {
		log.Fatalf("failed to parse 'args' in FuncCallOp, got '%s'", node.GetChildren()[2].GetValue())
	}

	return FuncCallOp{Name: name, ArgsNum: uint8(args)}, nil
}
