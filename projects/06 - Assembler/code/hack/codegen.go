package hack

import (
	"fmt"
	"strconv"
)

// ----------------------------------------------------------------------------
// Code Generator

// Takes some a set of 'hack.Instruction' and spits out their binary counterparts.
//
// In order to resolve user defined labels in A instructions, during initialization of
// of the Code Generator a Symbol Table should be provided.
type CodeGenerator struct {
	Program     []Instruction     // The set of instructions to convert in Hack binary format
	SymbolTable map[string]uint16 // Mapping to resolve user-defined labels to their underlying address
	nVarOffset  uint16            // Internal offset to allocate memory for new variables
}

// Initializes and returns to the caller a brand new 'CodeGenerator' struct.
func NewCodeGenerator(p []Instruction, st map[string]uint16) CodeGenerator {
	return CodeGenerator{Program: p, SymbolTable: st}
}

// Translate each instruction in the 'Program' to the Hack binary format.
//
// Each instruction will pass through the following step: evaluation, validation and then conversion
// to its binary representation (stored inside a uint16) so that it can be further elaborated by the
// function caller (e.g. dumping .hack code to a file, runtime interpretation, ...).
func (cg *CodeGenerator) Translate() ([]string, error) {
	hack := make([]string, 0, len(cg.Program))

	for _, instruction := range cg.Program {
		var hackInst string = ""
		var err error = nil

		switch tInstruction := instruction.(type) {
		case AInstruction:
			hackInst, err = cg.TranslateAInst(tInstruction)
		case CInstruction:
			hackInst, err = cg.TranslateCInst(tInstruction)
		}

		if err != nil {
			return nil, err
		}
		hack = append(hack, hackInst)
	}

	return hack, nil
}

// Specialized function to convert an A Instruction to the Hack format.
//
// As part of the conversion (for both built-in and user-defined labels) there's a lookup
// on their respective symbol tables in order to determine the 'real' location address.
// For location not resolved or resolved to an Out-of-Bound address an error is returned.
func (cg *CodeGenerator) TranslateAInst(inst AInstruction) (string, error) {
	found, address := false, uint16(0)

	switch inst.LocType {
	case Raw: // Simply translate the raw address from 'string' to 'int'
		num, err := strconv.ParseInt(inst.LocName, 10, 16)
		address, found = uint16(num), err == nil
	case Label: // Lookup the label name in the provided SymbolTable
		address, found = cg.SymbolTable[inst.LocName]
		// If not found we treat it as a new variable
		if !found {
			// Assign a new memory location starting from 16 onwards
			address, found = 16+cg.nVarOffset, true
			// And update the SymbolTable so that future references
			// gets resolved/points to the same locations in RAM
			cg.SymbolTable[inst.LocName] = address
			cg.nVarOffset++
		}
	case BuiltIn: // Lookup the registry name in the WellKnow table
		address, found = BuiltInTable[inst.LocName]
	}

	if !found {
		return "", fmt.Errorf("unable to resolve address for location '%s'", inst.LocName)
	}
	// An A instruction always has the first bit set to zero (the opcode bit) this also mean
	// that, since each instructions 16 bit there are only 15 bit to address the Hack computer
	// memory this in turn means that the an address over 2^15 is invalid and out of bound.
	if address > MaxAddressableMemory {
		return "", fmt.Errorf("location '%s resolved to an address not allowed", inst.LocName)
	}
	// So here we just need to convert the address to its 16 bit binary representation
	return fmt.Sprintf("%016b", address), nil
}

// Specialized function to convert a C Instruction to the Hack format.
//
// As part of the conversion (for both built-in and user-defined labels) there's a lookup
// on their respective symbol tables in order to determine the 'real' location address.
// For location not resolved or resolved to an Out-of-Bound address an error is returned.
func (cg *CodeGenerator) TranslateCInst(inst CInstruction) (string, error) {
	command := uint16(0b111 << 13) // Puts the initial '111' opcode at the start

	// Since the 'Comp' bit-codes are the only ones mandatory we check before translation
	// that the are provided, the check on their well-formed(ness) will come when querying
	// the translation mappings (this also applies to 'Dest' and 'Jump' bit-codes).
	if _, found := CompTable[inst.Comp]; inst.Comp == "" || !found {
		return "", fmt.Errorf("unable to translate C instruction, missing or invalid operation code")
	}

	// CInst.Comp: Command translation with bit-a-bit manipulation
	if opcode, found := CompTable[inst.Comp]; found {
		command |= opcode << 6
	} else {
		return "", fmt.Errorf("unable to translate C instruction, unknown 'comp' opcode '%s'", inst.Comp)
	}
	// CInst.Dest: Command translation with bit-a-bit manipulation
	if opcode, found := DestTable[inst.Dest]; found {
		command |= opcode << 3
	} else {
		return "", fmt.Errorf("unable to translate C instruction, unknown 'dest' opcode '%s'", inst.Dest)
	}
	// CInst.Jump: Command translation with bit-a-bit manipulation
	if opcode, found := JumpTable[inst.Jump]; found {
		command |= opcode
	} else {
		return "", fmt.Errorf("unable to translate C instruction, unknown 'jump' opcode '%s'", inst.Jump)
	}

	return fmt.Sprintf("%016b", command), nil
}
