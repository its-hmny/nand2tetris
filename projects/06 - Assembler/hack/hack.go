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

var (
	OpTable = map[string]uint16{
		"0": 0b0101010, "1": 0b011111, "-1": 0b0111010,
		"D": 0b0001100, "A": 0b0110000, "M": 0b1110000,
		"!D": 0b0001101, "!A": 0b0110001, "!M": 0b1110001,
		"-D": 0b0001111, "-A": 0b0110011, "-M": 0b1110011,
		"D+1": 0b0011111, "A+1": 0110111, "M+1": 0b1110111,
		"D-1": 0b0001110, "A-1": 0b0110010, "M-1": 0b1110010,
		"D+A": 0b0000010, "D+M": 0b1000010,
		"D-A": 0b0010011, "D-M": 0b1010011, "A-D": 0b0000111, "M-D": 0b1000111,
		"D&A": 0b0000000, "D&M": 0b100000, "D|A": 0b0010101, "D|M": 0b1010101,
	}

	DestTable = map[string]uint16{
		"M": 0b001, "D": 0b010, "A": 0b100,
		"MD": 0b011, "AM": 0b101, "AD": 0b110, "AMD": 0b111,
	}

	JumpTable = map[string]uint16{
		"JGT": 0b001, "JEQ": 0b010, "JGE": 0b011,
		"JLT": 0b100, "JNE": 0b101, "JLE": 0b110, "JMP": 0b111,
	}
)

// In memory representation of a C instruction for the Hack computer specification
type CInstruction struct {
	Operation   string // The calculation bit-code to be performed by the ALU
	Destination string // The destination registry where the result will be saved
	Jump        string // The jump directive to interrupt the linear flow execution
}
