package jack

import (
	"errors"
	"io"
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
	return Class{}, errors.New("Parser.Parse() not implemented yet")
}
