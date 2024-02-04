package vm

// ----------------------------------------------------------------------------
// General information

// This section contains some general information about the VM intermediate language.
//
// We declare a shared 'Operation' interface for every macro operation available for the
// language and we define some other useful top-level struct such as Program and Module.
// Is important to note that a VM program can be composed of multiple translation units
// that can be also referenced as file or modules or also classes.

// A VM Program is just a set of multiple modules/files, in the VM spec each Jack class is
// translated to its own .vm file (just like Java .class file) that can be handled as its
// own translation unit during the compilation or lowering phases.
type Program []Module

// A VM Module is just a linear list of VM operations/instructions
type Module []Operation

// Used to put together all operation in the VM language (Memory, Arithmetic, ... ops).
type Operation interface{}

// ----------------------------------------------------------------------------
// Memory Op

// In memory representation of a Memory operation for the VM language.
//
// In the VM intermediate language there are only two possible memory operation on the stack.
// We could either push a new value taken from the specified segment location on the stack's
// top or take the stack's top and saves its value at the specified segment location.
type MemoryOp struct {
	Operation OperationType // The type of operation, either 'push' or 'pop'
	Segment   SegmentType   // The named memory segment to use (this, that, temp, ...)
	Offset    uint16        // The specific location/offset inside of the memory segment
}

type OperationType string // Enum to manage the operation allowed for a MemoryOp

const (
	Push OperationType = "push"
	Pop  OperationType = "pop"
)

type SegmentType string // Enum to manage the segment accessible for a MemoryOp

const (
	Temp     SegmentType = "temp"     // Real segment used to store intermediate computations
	Constant SegmentType = "constant" // Virtual segment used to access numeric constant

	Local    SegmentType = "local"    // Real segment used to store local function variables
	Static   SegmentType = "static"   // Real segment used to store shared/static variables
	Argument SegmentType = "argument" // Real segment used to store function's argument

	This    SegmentType = "this"    // Virtual segment used to point to a specific memory location
	That    SegmentType = "that"    // Virtual segment used to point to a specific memory location
	Pointer SegmentType = "pointer" // Real segment w/ 2 location used to set the 'this' and 'that' pointers
)

// ----------------------------------------------------------------------------
// Arithmetic Op

// In memory representation of a Arithmetic operation for the VM language.
//
// In the VM intermediate language there are just a handful of operation available.
// In particular each operation acts directly on the top of the stack, of course we have both unary
// and binary operation, the specific management of each op will be handled in the codegen phase.
type ArithmeticOp struct{ Operation ArithOpType }

type ArithOpType string // Enum to manage the operation allowed for an ArithmeticOp

const (
	Eq ArithOpType = "eq" // Comparison operations
	Gt ArithOpType = "gt"
	Lt ArithOpType = "lt"

	Add ArithOpType = "add" // Arithmetic operations
	Sub ArithOpType = "sub"
	Neg ArithOpType = "neg"

	Not ArithOpType = "not" // Bitwise operations
	And ArithOpType = "and"
	Or  ArithOpType = "or"
)
