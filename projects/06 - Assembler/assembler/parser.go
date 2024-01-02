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

	return &pc.Terminal{Name: "Label", Value: location.Value, Position: location.Position}
}
func handleAInst(data []pc.ParsecNode) pc.ParsecNode {
	at, location := data[0].(*pc.Terminal), data[1].(*pc.Terminal)
	if at.Name != "@" || location.Name != "Label" {
		fmt.Println("ERR: Unable to parse malformed A Instruction")
	}

	inst := hack.AInstruction{LocType: hack.Label, LocName: location.Value}
	fmt.Printf("Found A instruction: %+v\n", inst)

	return pc.Terminal{Name: "AInst", Value: location.Value, Position: location.Position}
}

var (
	// Generic label parser (A Instruction + Label declaration)
	pLabel = pc.OrdChoice(handleLabel, pc.Int(), pc.Ident())

)

var (
	// A Instruction declaration parser
	pAInst = pc.And(handleAInst, pc.Atom("@", "@"), pLabel)

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

