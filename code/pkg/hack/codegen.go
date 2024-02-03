package hack

import (
	"fmt"
	"strconv"
)

// ----------------------------------------------------------------------------
// Translation tables

// This section contains the translation tables cornerstone of the codegen phase.
//
// This table provides a simple yet effective way to resolve the everything built-in and
// in the Hack specification. Notably we have a the following tables defined:
//	- 'BuiltInTable': Specifies how to translate BuiltIn labels in A instructions to their address
//  - 'CompTable': Specifies how to translate the 'Comp' opcode in C instructions
//  - 'DestTable': Specifies how to translate the 'Dest' opcode in C instructions
//  - 'JumpTable': Specifies how to translate the 'Jump' opcode in C instructions

var (
	BuiltInTable = map[string]uint16{
		// Virtual Machine specific aliases (see project 7)
		"SP": 0, "LCL": 1, "ARG": 2, "THIS": 3, "THAT": 4,
		// Named general purpose registers
		"R0": 0, "R1": 1, "R2": 2, "R3": 3, "R4": 4, "R5": 5,
		"R6": 6, "R7": 7, "R8": 8, "R9": 9, "R10": 10, "R11": 11,
		"R12": 12, "R13": 13, "R14": 14, "R15": 15,
		// Memory mapped I/O locations
		"SCREEN": 16384, "KBD": 24576,
	}

	CompTable = map[string]uint16{
		// - Constants and identities
		"0": 0b0101010, "1": 0b0111111, "-1": 0b0111010,
		"D": 0b0001100, "A": 0b0110000, "M": 0b1110000,
		// - Binary and numerical negations
		"!D": 0b0001101, "!A": 0b0110001, "!M": 0b1110001,
		"-D": 0b0001111, "-A": 0b0110011, "-M": 0b1110011,
		// - Increment and decrement operations
		"D+1": 0b0011111, "A+1": 0b0110111, "M+1": 0b1110111,
		"D-1": 0b0001110, "A-1": 0b0110010, "M-1": 0b1110010,
		// - Register with register operations
		"D+A": 0b0000010, "D+M": 0b1000010,
		"D-A": 0b0010011, "D-M": 0b1010011,
		"A-D": 0b0000111, "M-D": 0b1000111,
		// - Bitwise register with register operations
		"D&A": 0b0000000, "D&M": 0b1000000,
		"D|A": 0b0010101, "D|M": 0b1010101,
	}

	DestTable = map[string]uint16{
		"": 0b000, "M": 0b001, "D": 0b010, "A": 0b100,
		"MD": 0b011, "AM": 0b101, "AD": 0b110, "AMD": 0b111,
	}

	JumpTable = map[string]uint16{
		"": 0b000, "JGT": 0b001, "JEQ": 0b010, "JGE": 0b011,
		"JLT": 0b100, "JNE": 0b101, "JLE": 0b110, "JMP": 0b111,
	}
)

// ----------------------------------------------------------------------------
// Code Generator

// Takes some a set of 'hack.Instruction' and spits out their binary counterparts.
//
// In order to resolve user defined labels in A instructions, during initialization of
// of the Code Generator a Symbol Table should be provided.
type CodeGenerator struct {
	program    Program     // The set of instructions to convert in Hack binary format
	table      SymbolTable // Mapping to resolve user-defined labels to their underlying address
	nVarOffset uint16      // Internal offset to allocate memory for new variables
}

// Initializes and returns to the caller a brand new 'CodeGenerator' struct.
// Requires both a non-nil Program 'p' (what we want to translate) as well as
// an optionally nullable Symbol Table 'st' used to resolve user defined labels.
func NewCodeGenerator(p Program, st SymbolTable) CodeGenerator {
	return CodeGenerator{program: p, table: st}
}

// Translates each instruction in the 'Program' to the Hack binary format.
//
// Each instruction will pass through the following step: evaluation, validation and then conversion
// to its binary representation (stored inside a uint16) so that it can be further elaborated by the
// function caller (e.g. dumping .hack code to a file, runtime interpretation, ...).
func (cg *CodeGenerator) Generate() ([]string, error) {
	hack := make([]string, 0, len(cg.program))

	for _, instruction := range cg.program {
		var generated string = ""
		var err error = nil

		switch tInstruction := instruction.(type) {
		case AInstruction:
			generated, err = cg.GenerateAInst(tInstruction)
		case CInstruction:
			generated, err = cg.GenerateCInst(tInstruction)
		}

		if err != nil {
			return nil, err
		}
		hack = append(hack, generated)
	}

	return hack, nil
}

// Specialized function to convert an A Instruction to the Hack format.
//
// As part of the conversion (for both built-in and user-defined labels) there's a lookup
// on their respective symbol tables in order to determine the 'real' location address.
// For location not resolved or resolved to an Out-of-Bound address an error is returned.
func (cg *CodeGenerator) GenerateAInst(inst AInstruction) (string, error) {
	found, address := false, uint16(0)

	switch inst.LocType {
	case Raw: // Simply translate the raw address from 'string' to 'int'
		num, err := strconv.ParseInt(inst.LocName, 10, 16)
		address, found = uint16(num), err == nil
	case Label: // Lookup the label name in the provided SymbolTable
		address, found = cg.table[inst.LocName]
		// If not found we treat it as a new variable
		if !found {
			// Assign a new memory location starting from 16 onwards
			address, found = 16+cg.nVarOffset, true
			// And update the SymbolTable so that future references
			// gets resolved/points to the same locations in RAM
			cg.table[inst.LocName] = address
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
func (cg *CodeGenerator) GenerateCInst(inst CInstruction) (string, error) {
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
