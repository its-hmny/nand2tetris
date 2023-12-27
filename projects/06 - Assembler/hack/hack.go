package hack

// ======================================================================================
//  								 		General
// ======================================================================================

// In order to determine if we're working with an A or C instruction we can type
// switch on the interface value, sadly since A and C instructions are so different
// there's not really any shareable method to put inside the interface.
type Instruction interface{}

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




