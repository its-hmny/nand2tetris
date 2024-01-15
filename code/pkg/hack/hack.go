package hack

// ----------------------------------------------------------------------------
// General information

// This section contains some general information about the Hack instruction set.
//
// We declare a shared 'Instruction' interface for both A and C instructions as well
// as defining some useful constants for runtime assertions during the codegen phase
// such as the 'MaxAddressableMemory' that defines the upper limit to Memory capacity
// and the 'BuiltInTable' that defines the hardcoded symbols for some well-known registers.

// Just used to put together A and C instructions struct, use type switch to disambiguate.
type Instruction interface{}

const MaxAddressableMemory uint16 = (1 << 15) // Max memory location available from an A instruction.

var BuiltInTable = map[string]uint16{ // Predefined/Built-In symbols by the Hack instruction set.
	// Virtual Machine specific aliases (see project 7)
	"SP": 0, "LCL": 1, "ARG": 2, "THIS": 3, "THAT": 4,
	// Named general purpose registers
	"R0": 0, "R1": 1, "R2": 2, "R3": 3, "R4": 4, "R5": 5,
	"R6": 6, "R7": 7, "R8": 8, "R9": 9, "R10": 10, "R11": 11,
	"R12": 12, "R13": 13, "R14": 14, "R15": 15,
	// Memory mapped I/O locations
	"SCREEN": 16384, "KBD": 24576,
}

// Also defined here we can find three direct translation tables, these are purposefully
// built to simplify the conversion from textual to binary of C instructions during the
// codegen phase. As one can see there's a mapping table for each part of a C instruction.
var (
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
// A Instructions

// In memory representation of an A Instruction for the Hack architecture spec.
//
// The A instruction has only one functionality in the Hack computer, it instructs
// the CPU to load a specific memory address/location from the computer memory (this
// includes both the RAM and the memory mapped I/O).
//
// The location can be expressed in multiple way:
// - A raw memory address (e;g; 1, 2, 3)
// - Specified through a user defined label (e;g; LOOP, ADD, TEMP)
// - Using one of the built-in symbols of the Hack architecture (e;g; SP, THIS, THAT)
type AInstruction struct {
	LocType LocationType // The subtype of the location identified by 'Name'
	LocName string       // A generic "payload" (the label/builtin/raw symbol)
}

type LocationType uint8 // Enumeration for all the different type of location (built-in, label, raw)

const (
	Raw     LocationType = 0 // Raw address literal (e.g. @2345, @8989)
	Label   LocationType = 1 // User-defined location w/ a user given name (e.g. @MAIN, @LOOP)
	BuiltIn LocationType = 2 // Predefined  associations by the Hack specs (@SCREEN, @KBD, @R1)
)

// ----------------------------------------------------------------------------
// C Instructions

// In memory representation of an C Instruction for the Hack architecture spec.
//
// The C instruction handles the computation side of the Hack computer, it instructs
// the CPU on what operation to execute and which register to use, also it allows to
// specify jump conditions to change the execution flow at runtime.
//
// The location can be expressed in multiple way:
// - A raw memory address (e;g; 1, 2, 3)
// - Specified through a user defined label (e;g; LOOP, ADD, TEMP)
// - Using one of the built-in symbols of the Hack architecture (e;g; SP, THIS, THAT)
type CInstruction struct {
	Comp string // The 'computation' bit-codes, defines the calculation that the CPU should perform
	Dest string // The 'destination' bit-codes, defines if/where the result should be saved
	Jump string // The 'jump' bit-codes, define on what premise the jump to another instruction should occur
}
