package hack

// ----------------------------------------------------------------------------
// General information

// This section contains some general information about the Hack instruction set.
//
// We declare a shared 'Instruction' interface for both A and C instructions as well
// as defining some useful constants for runtime assertions during the codegen phase
// such as the 'MaxAddressableMemory' that defines the upper limit to Memory capacity.

// Just used to put together A and C instructions struct, use type switch to disambiguate.
type Instruction interface{}

const MaxAddressableMemory uint16 = (1 << 15) // Max memory address indexable for an A Instruction.

// ----------------------------------------------------------------------------
// A Instructions

// In memory representation of an A Instruction for the Hack architecture spec.
//
// The A instruction has only one functionality in the Hack computer, it instructs
// the CPU to load a specific memory address from the computer memory (this includes
// both the RAM as well as the memory mapped I/O such as Keyboard and Screen).
//
// The location can be expressed in multiple way:
// - A raw memory address (e.g. 1, 2, 3)
// - A user defined label (e.g. LOOP, ADD, TEMP)
// - A built-in symbols from the Hack architecture spec (e.g. SP, THIS, THAT)
type AInstruction struct {
	LocType LocationType // The type of the location identified by 'Name' field
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
// - A raw memory address (e.g. 1, 2, 3)
// - Specified through a user defined label (e.g. LOOP, ADD, TEMP)
// - Using one of the built-in symbols of the Hack architecture (e.g. SP, THIS, THAT)
type CInstruction struct {
	Comp string // The 'computation' bit-codes, defines the calculation that the CPU should perform
	Dest string // The 'destination' bit-codes, defines if/where the result should be saved
	Jump string // The 'jump' bit-codes, define on what premise the jump to another instruction should occur
}
