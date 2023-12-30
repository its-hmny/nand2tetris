package hack

// ======================================================================================
//  								 		General
// ======================================================================================

// In order to determine if we're working with an A or C instruction we can type
// switch on the interface value, sadly since A and C instructions are so different
// there's not really any shareable method to put inside the interface.
type Instruction interface{}

const (
	MaxAddressAllowed uint16 = (1 << 15)
)

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
)

// ======================================================================================
//  									A Instructions
// ======================================================================================

type LocationType int // Enumeration for all the different type of location (built-in, label, raw)

const (
	Raw     LocationType = 0 // Raw address literal (e.g. @2345, @8989)
	Label   LocationType = 1 // User-defined location w/ a user given name (e.g. @MAIN, @LOOP)
	BuiltIn LocationType = 2 // Predefined  associations by the Hack specs (@SCREEN, @KBD, @R1)
)

// In memory representation of an A Instruction for the Hack computer specification
type AInstruction struct {
	LocType LocationType // The subtype of the location identified by 'Name'
	LocName string       // A generic payload (the label/builtin name or the raw address)
}

// ======================================================================================
//										C Instructions
// ======================================================================================

// In memory representation of a C instruction for the Hack computer specification
type CInstruction struct {
	// TODO (hmny): Populate this struct w/ the required fields
}




