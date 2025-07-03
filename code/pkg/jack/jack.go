package jack

import "its-hmny.dev/nand2tetris/pkg/utils"

// ----------------------------------------------------------------------------
// General information

// This section contains some general information about the Jack programming language.
//
// A program is basically a container of classes (the  only top-level object allowed)
// and the program is started by locating the Main class and executing its 'main' method.
// Other than classes the other 4 main constructs are:
// - Variables: to declare containers of value (also used for class' fields)
// - Subroutines: to declare containers of instruction (also used for class' methods)
// - Statements: to perform a side effect, conditional jump or other program flow changes
// - Expressions: to perform a calculation that produces a result (arithmetic ops and so on...)

// A Jack Program is just a set of multiple classes, in the Jack spec each class is translated
// to its own .vm file (just like Java .class file) so the class is to be considered the top-level
// entity of the program and is mapped to a role equal to module or namespace in other languages.
type Program map[string]Class

// ----------------------------------------------------------------------------
// Classes

// A Class is a list of Fields that contains the state and Subroutines to change said state.
//
// Both Fields and Subroutines comes in a static variant (resp. static 'Variable' or function Subroutine) where
// the instance of the class is not scoped to the single object instantiation but to the program as a whole
type Class struct {
	Name        string                               // The class name or id, will also identify the instantiated object type
	Fields      utils.OrderedMap[string, Variable]   // The variable (static ors not) associated to the class or object instance
	Subroutines utils.OrderedMap[string, Subroutine] // The subroutines (static or not) associated to the class or object instance
}

// ----------------------------------------------------------------------------
// Subroutines

// A Subroutine is somewhat like a math function: it takes a series and inputs and returns and output.
//
// As part of its computation (statement evaluation) it may change the state of some variables in the
// program either by direct manipulation of the class' fields (static or nor) or by just returning values
// that will influence the program flow once returned to the caller.
type Subroutine struct {
	Name string         // Name/id, w/ the class id will identify universally the subroutine
	Type SubroutineType //Function type, used to determine the codegen strategy during compilation phase

	Return    DataType            // The type of value returned by the procedure ('void' for no value)
	Arguments map[string]Variable // The set of arguments to be provided and used during the execution

	Statements []Statement // The list of statements to be executed, a representation of the func program flow
}

type SubroutineType string // Enum to manage the different type allowed for a Subroutine

const (
	Method      SubroutineType = "method"
	Function    SubroutineType = "function"
	Constructor SubroutineType = "constructor"
)

// ----------------------------------------------------------------------------
// Statements

// A statement produces a side effect in the program flow wether by changing a var or jumping to another inst.
//
// We declare a shared 'Statement' interface for every macro operation available for
// the Jack language, then  we define one after the other all the specific statements
// w/ their internal logic and required data to perform it (or compile it).
type Statement interface{}

type DoStmt struct { // Unconditional jump, will call another subroutine and ignore its return value
	FuncCall FuncCallExpr //The function to be called
}

type VarStmt struct { // Variable declaration construct, will allocate a new var w/o a given value
	Vars []Variable // The name or identifiers of the new local variables
}

type LetStmt struct { // Variable assignment construct, will allocate a new var w/ a given value
	Lhs Expression // The expression to be assigned the value (only VarExpr and ArrayExpr are allowed)
	Rhs Expression // The expression to be evaluated and assigned to the LHS counterpart (all Expression are allowed)
}

type ReturnStmt struct { // Unconditional jump, will go back to the caller and provide it an (optional) output
	Expr Expression // The expression to be eval'd, casted to a the return value of the func
}

type IfStmt struct { // Conditional jump construct, will have to fork the execution flow based on a condition
	Condition Expression  // The expression to be eval'd, casted to a bool value
	ThenBlock []Statement // The code block to be executed if the condition is met
	ElseBlock []Statement // The code block to be executed if the condition is not met
}

type WhileStmt struct { // Conditional iteration construct, will execute a block based on a condition
	Condition Expression  // The expression to be eval'd, casted to a bool value
	Block     []Statement // The code block to be executed if the condition is met
}

// ----------------------------------------------------------------------------
// Expressions

// Expression take one or two sub-expressions and create a new value that can be used further.
//
// We declare a shared 'Expression' interface for every macro operation available for
// the Jack language, then  we define one after the other all the specific expressions
// w/ their internal logic and required data to perform it (or compile it).
type Expression interface{}

type VarExpr struct { // Extracts the value contained in a variable
	Var string // The name or identifier of the variable we want the value of
}

type LiteralExpr struct { // Extracts the value of a constant (also called literal)
	Type  DataType // The literal type (string, int, char, ...)
	Value string   // The constant value to be produced
}

type ArrayExpr struct { // Extracts the value of a single cell/element for an array
	Var   string     // The name or identifier of the array we want the value from
	Index Expression // The index of the value we want to extract
}

type UnaryExpr struct { // Applies a transformation to 1 expression to produce a new value
	Type ExprType   //  Here only 'Minus' and 'BoolNot' are allowed
	Rhs  Expression // UnaryExpr do only apply to the expr on the Right Hand Side
}

type BinaryExpr struct { // Combines the value of 2 expression to produce a new value
	Type ExprType   // Here only 'BoolNot' is not allowed
	Lhs  Expression // The expression o the Left Hand Side (1st to be evaluated)
	Rhs  Expression // The expression o the Right Hand Side (2nd to be evaluated)
}

type FuncCallExpr struct { // Call another subroutine for a variable or inside the same class
	IsExtCall bool   // Manages call from outside the class, e.g. 'class.Method(x, y)'
	Var       string // The object instance that has the desired subroutine ("" if IsExtCall = false)
	FuncName  string // The name/id of the desired subroutine we want to execute

	Arguments []Expression // The arguments list to be passed (they are yet to be evaluated)
}

type ExprType string // Enum to manage the operation allowed for an ExprType

const (
	Plus     ExprType = "plus"
	Minus    ExprType = "minus" // Used both for subtraction (BinaryExpr) and arithmetic negation (UnaryExpr)
	Divide   ExprType = "divide"
	Multiply ExprType = "multiply"

	BoolOr  ExprType = "bool_or"
	BoolAnd ExprType = "bool_and"
	BoolNot ExprType = "bool_neg"

	Equal     ExprType = "equal"
	LessThan  ExprType = "less_than"
	GreatThan ExprType = "greater_than"
)

// ----------------------------------------------------------------------------
// Variables

// Variables are containers of value that can be read/written through expressions/statements.
//
// The declared 'Variable' struct accommodates multiple configurations at the same time such as
// - Static & instanced fields for classes
// - Local variables and parameters for subroutines
type Variable struct {
	Name      string   // The var name, acts as identifier in the scope it is declared
	Type      VarType  // The variable type helps determine the scope of the variable
	DataType  DataType // The data type defines how to read or cast the value contained by the variable
	ClassName string   // The additional and specific class type if (DataType = Object)
}

type VarType string // Enum to manage the operation allowed for an VarType

const (
	Local     VarType = "local"
	Field     VarType = "field"
	Static    VarType = "static"
	Parameter VarType = "parameter"
)

type DataType string // Enum to manage the operation allowed for an DataType

const (
	Int    DataType = "int"
	Bool   DataType = "bool"
	Char   DataType = "char"
	Null   DataType = "null"
	String DataType = "string"
	Void   DataType = "void"
	Object DataType = "object"
)
