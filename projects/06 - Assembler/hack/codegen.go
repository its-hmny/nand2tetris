package hack

import (
	"fmt"
	"log"
	"strconv"
)

type CodeGenerator struct {
	Program []Instruction     // The set of parsed instructions to convert in Hack format
	Table   map[string]uint16 // Used to resolve user-defined labels w/ their actual address
}

func (cg *CodeGenerator) Dump() ([]string, error) {
	code := make([]string, 0, len(cg.Program))

	for _, inst := range cg.Program {
		var hackInst string = ""
		var err error = nil

		switch typedInst := inst.(type) {
		case AInstruction:
			hackInst, err = cg.TranslateAInst(typedInst)
		case CInstruction:
			log.Fatal("TODO: Yet to be implemented")
		}

		if hackInst == "" || err != nil {
			return nil, err
		}

		code = append(code, hackInst)
	}

	return code, nil
}

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
	if address > MaxAddressAllowed {
		return "", fmt.Errorf("location '%s resolved to an address not allowed", inst.LocName)
	}
	// So here we just need to convert the address to its 16 bit binary representation
	return fmt.Sprintf("%016b", address), nil
}
