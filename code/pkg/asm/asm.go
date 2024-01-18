package asm

// ----------------------------------------------------------------------------
// General information

// This section contains some general information about the Asm language.
//
// We declare a shared 'Statement' interface for both A and C instructions as well as defining
// custom labels for specific code section (allowing arbitrary jumps) at runtime during code execution.
// This in turns enables iterations and conditionals both here and at the upper levels (VM, Compiler).

// Just used to put together label declaration, A inst and C inst in the same datatype.
type Statement interface{}

// ----------------------------------------------------------------------------
// Label Declarations

// In memory representation of a label declaration statement for the Assembler language.
//
// There's not much here to be honest, we just keep track of the user defined name to resolve
// future references to the same label (e.g. when referencing a label in an A Instruction).
// During the lowering phases this label will be mapped to their location in the program
// and a symbol table will be generated from it, the latter will be used in the codegen phase.
type LabelDecl struct {
	Name string // The symbol/ident chosen by the user for the label
}

// ----------------------------------------------------------------------------
// A Instructions

// In memory representation of an A Instruction for the Assembler language.
//
// The A instruction has only one functionality in the Hack computer, it instructs
// the CPU to load a specific memory address/location from the computer memory (this
// includes both the RAM and the memory mapped I/O). The location can be referenced
// either by an alias (labels) or by specifying the raw location.
// During the lowering phase each label will be assigned its type (Raw | BuiltIn | Label).
type AInstruction struct {
	Location string // A generic "payload" (the label/builtin/raw symbol)
}

// ----------------------------------------------------------------------------
// C Instructions

// In memory representation of an C Instruction for the Assembler language.
//
// The C instruction handles the computation side of the Hack computer, it instructs
// the CPU on what operation to execute and which register to use, also it allows to
// specify jump conditions to change the execution flow at runtime.
type CInstruction struct {
	Comp string // The 'computation' bit-codes, defines the calculation that the CPU should perform
	Dest string // The 'destination' bit-codes, defines if/where the result should be saved
	Jump string // The 'jump' bit-codes, define on what premise the jump to another instruction should occur
}
