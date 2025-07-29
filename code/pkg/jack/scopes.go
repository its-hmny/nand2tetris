package jack

import (
	"fmt"
	"strings"

	"its-hmny.dev/nand2tetris/pkg/utils"
)

type Scope struct {
	name    string
	entries utils.Stack[Variable]
}

type ScopeTable struct {
	static utils.Stack[Variable]

	local     Scope
	field     Scope
	parameter Scope
}

func NewScopeTable() *ScopeTable {
	return &ScopeTable{
		static:    utils.Stack[Variable]{},
		local:     Scope{},
		field:     Scope{},
		parameter: Scope{},
	}
}

func (st *ScopeTable) PushClassScope(class string) {
	newScope := fmt.Sprintf("%s.Global", class)
	st.field = Scope{name: newScope, entries: utils.Stack[Variable]{}}
}

func (st *ScopeTable) PopClassScope() { st.field = Scope{} }

func (st *ScopeTable) PushSubRoutineScope(method string) {
	newScope := strings.ReplaceAll(st.GetScope(), "Global", method)
	st.local = Scope{name: newScope, entries: utils.Stack[Variable]{}}
	st.parameter = Scope{name: newScope, entries: utils.Stack[Variable]{}}
}

func (st *ScopeTable) PopSubroutineScope() { st.local, st.parameter = Scope{}, Scope{} }

func (st *ScopeTable) GetScope() string {
	if st.local.name != "" && st.parameter.name != "" {
		return st.local.name
	}

	if st.field.name != "" {
		return st.field.name
	}

	return "Global"
}

func (st *ScopeTable) RegisterVariable(new Variable) {
	switch new.VarType {
	case Local:
		st.local.entries.Push(new)
	case Field:
		st.field.entries.Push(new)
	case Parameter:
		st.parameter.entries.Push(new)
	case Static:
		st.static.Push(new)
	}
}

func (st *ScopeTable) ResolveVariable(name string) (uint16, Variable, error) {
	scopes := []utils.Stack[Variable]{st.local.entries, st.parameter.entries, st.field.entries, st.static}

	for _, scope := range scopes {
		for idx, entry := range scope.Iterator() {
			if entry.Name == name {
				return uint16(idx), entry, nil
			}
		}
	}

	return 0, Variable{}, fmt.Errorf("variable '%s' undeclared, not found in any scope", name)
}
