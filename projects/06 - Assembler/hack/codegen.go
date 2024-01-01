package hack

import (
	"fmt"
	"log"
	"strconv"
)

// ----------------------------------------------------------------------------
// Code Generator

// Takes some a set of 'hack.Instruction' and spits out their binary counterparts.
//
// In order to resolve user defined labels in A instructions, during initialization of
// of the Code Generator a Symbol Table should be provided.
type CodeGenerator struct {
	Program []Instruction     // The set of instructions to convert in Hack binary format
	Table   map[string]uint16 // Mapping to resolve user-defined labels to their underlying address
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
			log.Fatal("TODO: Yet to be implemented")
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
		address, found = cg.Table[inst.LocName]
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
