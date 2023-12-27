package hack

// ======================================================================================
//  								 		General
// ======================================================================================

type InstructionType int

const (
	A InstructionType = 0 // A instruction manage only the fetch from a memory location
	C InstructionType = 1 // C instruction manage arithmetic as well as jump operations
)

type Instruction interface {
	Type() InstructionType // To determine if we're working with an A or C instruction
}

// ======================================================================================
//  									A Instructions
// ======================================================================================

type LocType int // Enumeration for all the different type of location (built-in, label, raw)

const (
	LocTypeRaw     LocType = 0 // Raw address literal (e.g. @2345, @8989)
	LocTypeLabel   LocType = 1 // User-defined location w/ a user given name (e.g. @MAIN, @LOOP)
	LocTypeBuiltIn LocType = 2 // Predefined  associations by the Hack specs (@SCREEN, @KBD, @R1)
)

// In memory representation of an A Instruction for the Hack computer specification
type AInstruction struct {
	LocationType LocType // The subtype of the location identified by 'Name'
	LocationName string  // A generic payload (the label/builtin name or the raw address)
}

func (AInstruction) Type() InstructionType { return A }

// ======================================================================================
//										C Instructions
// ======================================================================================

// In memory representation of a C instruction for the Hack computer specification
type CInstruction struct {
	// TODO (hmny): Populate this struct w/ the required fields
}

func (CInstruction) Type() InstructionType { return C }

/*
