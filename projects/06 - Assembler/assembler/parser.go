package assembler

import (
	"encoding/json"
	"fmt"
	"io"

	pc "github.com/prataprc/goparsec"
	"its-hmny.dev/n2t-assembler/hack"
)

func handleLabel(data []pc.ParsecNode) pc.ParsecNode {
	location, err := data[0].(*pc.Terminal)
	if !err {
		fmt.Println("ERR: Unable to cast 'pc.ParsecNode' to 'pc.Terminal'")
	}

	return pc.NewTerminal("Label", location.Value, location.Position)
}

func handleDest(data []pc.ParsecNode) pc.ParsecNode {
	dest, err := data[0].(*pc.Terminal)
	if !err {
		fmt.Println("ERR: Unable to cast 'pc.ParsecNode' to 'pc.Terminal'")
	}

	return pc.NewTerminal("Dest", dest.Value, dest.Position)
}

func handleComp(data []pc.ParsecNode) pc.ParsecNode {
	comp, err := data[0].(*pc.Terminal)
	if !err {
		fmt.Println("ERR: Unable to cast 'pc.ParsecNode' to 'pc.Terminal'")
	}

	return pc.NewTerminal("Comp", comp.Value, comp.Position)
}

func handleJump(data []pc.ParsecNode) pc.ParsecNode {
	jump, err := data[0].(*pc.Terminal)
	if !err {
		fmt.Println("ERR: Unable to cast 'pc.ParsecNode' to 'pc.Terminal'")
	}

	return pc.NewTerminal("Comp", jump.Value, jump.Position)
}

func handleAInst(data []pc.ParsecNode) pc.ParsecNode {
	at, location := data[0].(*pc.Terminal), data[1].(*pc.Terminal)
	if at.Name != "@" || location.Name != "Label" {
		fmt.Println("ERR: Unable to parse malformed A Instruction")
	}

	inst := hack.AInstruction{LocType: hack.Label, LocName: location.Value}
	fmt.Printf("Found A instruction: %+v\n", inst)

	return pc.NewTerminal("AInst", location.Value, location.Position)
}

func handleCInst(data []pc.ParsecNode) pc.ParsecNode {
	dest, comp, jump := data[0].(*pc.Terminal), data[2].(*pc.Terminal), data[4].(*pc.Terminal)
	if comp.Name == "" {
		fmt.Println("ERR: Unable to parse malformed C Instruction")
	}

	inst := hack.CInstruction{Comp: comp.Value, Dest: dest.Value, Jump: jump.Value}
	fmt.Printf("Found C instruction: %+v\n", inst)

	return &pc.Terminal{} // TODO (hmny): No idea what to put here
}

func handleLabelDecl(data []pc.ParsecNode) pc.ParsecNode {
	location := data[1].(*pc.Terminal)
	if location.Name != "Label" {
		fmt.Println("ERR: Unable to parse malformed Label Declaration")
	}

	fmt.Printf("Found Label declaration: %s\n", location.Value)

	return pc.NewTerminal("Label", location.Value, location.Position)
}

var (
	// Generic label parser (A Instruction + Label declaration)
	// TODO (hmny): pc.Ident() may be too wide for Assembler label declaration
	pLabel = pc.OrdChoice(handleLabel, pc.Int(), pc.Ident())

	// TODO(hmny): Dest in C instruction is optional, must be handled
	// Generic destination parser (C Instruction subsection)
	pDest = pc.OrdChoice(handleDest, pc.Atom("D", "D"), pc.Atom("A", "A"), pc.Atom("M", "M"))

	// Generic computation parser (C Instruction subsection)
	// NOTE: The order of the Atom is reversed w.r.t. the one provided in the translation table cause
	// if not the 'Constant and identifiers' part will match before the order (BFS Search algorithm)
	pComp = pc.OrdChoice(handleComp,
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
	pJump = pc.OrdChoice(handleJump,
		pc.Atom("JGT", "JGT"), pc.Atom("JEQ", "JEQ"), pc.Atom("JGE", "JGE"),
		pc.Atom("JLT", "JLT"), pc.Atom("JLE", "JLE"), pc.Atom("JMP", "JMP"),
	)
)

var (
	// A Instruction declaration parser
	pAInst = pc.And(handleAInst, pc.Atom("@", "@"), pLabel)

	// C Instruction declaration parser
	pCInst = pc.And(handleCInst, pDest, pc.Atom("=", "="), pComp, pc.Atom(";", ";"), pJump)

	// New Label declaration parser
	pLabelDecl = pc.And(handleLabelDecl, pc.Atom("(", "("), pLabel, pc.Atom(")", ")"))

	// Assembler program parser
	pProgram = pc.ManyUntil(nil, pc.OrdChoice(nil, pAInst, pCInst, pLabelDecl), pc.End())
)

type Parser struct {
	Reader io.Reader
}

func (p *Parser) Parse() (bool, error) {
	content, err := io.ReadAll(p.Reader)
	if err != nil {
		return false, fmt.Errorf("unable to read content: %s", err)
	}

	parsed, _ := pProgram(pc.NewScanner(content))

	// Debug only, in future I'll try to create an AST from it with something
	// like this: 'pc.NewAST("asm", 10_000_000).Parsewith(pProgram, scanner)'
	json, _ := json.MarshalIndent(parsed, "", "  ")
	fmt.Printf("%T: %s\n", parsed, json)

	return true, nil

}

