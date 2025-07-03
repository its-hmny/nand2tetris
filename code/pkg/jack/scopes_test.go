package jack_test

import (
	"testing"

	"its-hmny.dev/nand2tetris/pkg/jack"
)

func TestClassScope(t *testing.T) {
	test := func(st jack.ScopeTable, lookup string, expectedVar jack.Variable, expectedOffset uint16, fail bool) {
		offset, variable, err := st.ResolveVariable(lookup)
		if err != nil && !fail {
			t.Fatalf("expected to find %s, got error: %v", lookup, err)
		}
		if variable != expectedVar {
			t.Errorf("expected to find variable '%s', got %+v", lookup, expectedVar)
		}
		if offset != expectedOffset {
			t.Errorf("expected to find offset %d for variable '%s', got '%d'", expectedOffset, lookup, offset)
		}
	}

	t.Run("Without variable shadowing", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass") // Push a new class scope before doing anything

		// Register a field variable and a static variable
		st.RegisterVariable(jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.String})
		st.RegisterVariable(jack.Variable{Name: "test_field_2", Type: jack.Field, DataType: jack.Char})
		st.RegisterVariable(jack.Variable{Name: "test_static_2", Type: jack.Static, DataType: jack.Bool})

		// All of these variables should be found and resolved correctly
		test(st, "test_field", jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Int}, 0, false)
		test(st, "test_static", jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.String}, 0, false)
		test(st, "test_field_2", jack.Variable{Name: "test_field_2", Type: jack.Field, DataType: jack.Char}, 1, false)
		test(st, "test_static_2", jack.Variable{Name: "test_static_2", Type: jack.Static, DataType: jack.Bool}, 1, false)

		// All of these variables should be found and resolved correctly
		test(st, "random1", jack.Variable{}, 0, true)
		test(st, "random2", jack.Variable{}, 0, true)
		test(st, "random3", jack.Variable{}, 0, true)
		test(st, "random4", jack.Variable{}, 0, true)
	})

	t.Run("With variable shadowing", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass") // Push a new class scope before doing anything

		// Register a field variable and a static variable
		st.RegisterVariable(jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.String})
		st.RegisterVariable(jack.Variable{Name: "test_class", Type: jack.Static, DataType: jack.Object, ClassName: "AnotherClass"})
		// These two variables should shadow the previous ones
		st.RegisterVariable(jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Char})
		st.RegisterVariable(jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.Bool})
		st.RegisterVariable(jack.Variable{Name: "test_class", Type: jack.Static, DataType: jack.Object, ClassName: "Class"})

		// All of these variables should be found and resolved correctly
		test(st, "test_field", jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Char}, 1, false)
		test(st, "test_static", jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.Bool}, 2, false)
		test(st, "test_class", jack.Variable{Name: "test_class", Type: jack.Static, DataType: jack.Object, ClassName: "Class"}, 3, false)

		// All of these variables should be found and resolved correctly
		test(st, "random1", jack.Variable{}, 0, true)
		test(st, "random2", jack.Variable{}, 0, true)
		test(st, "random3", jack.Variable{}, 0, true)
		test(st, "random4", jack.Variable{}, 0, true)
	})

	t.Run("With scope deallocation", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass") // Push a new class scope before doing anything

		// Register a field variable and a static variable
		st.RegisterVariable(jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test_field_2", Type: jack.Field, DataType: jack.Char})
		st.RegisterVariable(jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.String})
		st.RegisterVariable(jack.Variable{Name: "test_static_2", Type: jack.Static, DataType: jack.Bool})

		// All of these variables should be found and resolved correctly
		test(st, "test_field", jack.Variable{Name: "test_field", Type: jack.Field, DataType: jack.Int}, 0, false)
		test(st, "test_field_2", jack.Variable{Name: "test_field_2", Type: jack.Field, DataType: jack.Char}, 1, false)
		test(st, "test_static", jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.String}, 0, false)
		test(st, "test_static_2", jack.Variable{Name: "test_static_2", Type: jack.Static, DataType: jack.Bool}, 1, false)

		st.PopClassScope() // Deallocates the current class scope

		// All of these variables should not be found and resolved since the scope is deallocated
		test(st, "test_field", jack.Variable{}, 0, true)
		test(st, "test_field_2", jack.Variable{}, 0, true)
		// All of these variables should found and resolved correctly since they are static and span all scopes
		test(st, "test_static", jack.Variable{Name: "test_static", Type: jack.Static, DataType: jack.String}, 0, false)
		test(st, "test_static_2", jack.Variable{Name: "test_static_2", Type: jack.Static, DataType: jack.Bool}, 1, false)
	})
}

func TestSubroutineScope(t *testing.T) {
	test := func(st jack.ScopeTable, lookup string, expectedVar jack.Variable, expectedOffset uint16, fail bool) {
		offset, variable, err := st.ResolveVariable(lookup)
		if err != nil && !fail {
			t.Fatalf("expected to find %s, got error: %v", lookup, err)
		}
		if variable != expectedVar {
			t.Errorf("expected to find variable '%s', got %+v", lookup, expectedVar)
		}
		if offset != expectedOffset {
			t.Errorf("expected to find offset %d for variable '%s', got '%d'", expectedOffset, lookup, offset)
		}
	}

	t.Run("Without variable shadowing", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass")      // Push a new class scope before doing anything
		st.PushClassScope("TestSubroutine") // Push a new subroutine scope before doing anything

		// Register a field variable and a static variable
		st.RegisterVariable(jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.String})
		st.RegisterVariable(jack.Variable{Name: "test_local_2", Type: jack.Local, DataType: jack.Char})
		st.RegisterVariable(jack.Variable{Name: "test_parameter_2", Type: jack.Parameter, DataType: jack.Bool})

		// All of these variables should be found and resolved correctly
		test(st, "test_local", jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Int}, 0, false)
		test(st, "test_parameter", jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.String}, 0, false)
		test(st, "test_local_2", jack.Variable{Name: "test_local_2", Type: jack.Local, DataType: jack.Char}, 1, false)
		test(st, "test_parameter_2", jack.Variable{Name: "test_parameter_2", Type: jack.Parameter, DataType: jack.Bool}, 1, false)

		// All of these variables should be found and resolved correctly
		test(st, "random1", jack.Variable{}, 0, true)
		test(st, "random2", jack.Variable{}, 0, true)
		test(st, "random3", jack.Variable{}, 0, true)
		test(st, "random4", jack.Variable{}, 0, true)
	})

	t.Run("With variable shadowing (on method scope)", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass")      // Push a new class scope before doing anything
		st.PushClassScope("TestSubroutine") // Push a new subroutine scope before doing anything

		// Register a field variable and a static variable
		st.RegisterVariable(jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.String})
		st.RegisterVariable(jack.Variable{Name: "test_class", Type: jack.Parameter, DataType: jack.Object, ClassName: "AnotherClass"})
		// These two variables should shadow the previous ones
		st.RegisterVariable(jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Char})
		st.RegisterVariable(jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.Bool})
		st.RegisterVariable(jack.Variable{Name: "test_class", Type: jack.Parameter, DataType: jack.Object, ClassName: "Class"})

		// All of these variables should be found and resolved correctly
		test(st, "test_local", jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Char}, 1, false)
		test(st, "test_parameter", jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.Bool}, 2, false)
		test(st, "test_class", jack.Variable{Name: "test_class", Type: jack.Parameter, DataType: jack.Object, ClassName: "Class"}, 3, false)

		// All of these variables should be found and resolved correctly
		test(st, "random1", jack.Variable{}, 0, true)
		test(st, "random2", jack.Variable{}, 0, true)
		test(st, "random3", jack.Variable{}, 0, true)
		test(st, "random4", jack.Variable{}, 0, true)
	})

	t.Run("With scope deallocation", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass")      // Push a new class scope before doing anything
		st.PushClassScope("TestSubroutine") // Push a new subroutine scope before doing anything

		// Register a field variable and a static variable
		st.RegisterVariable(jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.String})

		// All of these variables should be found and resolved correctly
		test(st, "test_local", jack.Variable{Name: "test_local", Type: jack.Local, DataType: jack.Int}, 0, false)
		test(st, "test_parameter", jack.Variable{Name: "test_parameter", Type: jack.Parameter, DataType: jack.String}, 0, false)

		st.PopSubroutineScope() // Deallocates the current subroutine scope

		// All of these variables should not be found and resolved since the scope is deallocated
		test(st, "test_local", jack.Variable{}, 0, true)
		test(st, "test_parameter", jack.Variable{}, 0, true)
	})

	t.Run("With variable shadowing (on class scope)", func(t *testing.T) {
		st := jack.ScopeTable{}
		st.PushClassScope("TestClass") // Push a new class scope before doing anything

		// Register variables on the class scope
		st.RegisterVariable(jack.Variable{Name: "test1", Type: jack.Field, DataType: jack.Int})
		st.RegisterVariable(jack.Variable{Name: "test2", Type: jack.Static, DataType: jack.String})

		st.PushSubRoutineScope("TestSubroutine") // Push a new subroutine scope before doing anything

		// Register variables on the class scope
		st.RegisterVariable(jack.Variable{Name: "test1", Type: jack.Local, DataType: jack.Bool})
		st.RegisterVariable(jack.Variable{Name: "test2", Type: jack.Parameter, DataType: jack.Char})

		// All of these variables should be found and resolved correctly
		test(st, "test1", jack.Variable{Name: "test1", Type: jack.Local, DataType: jack.Bool}, 0, false)
		test(st, "test2", jack.Variable{Name: "test2", Type: jack.Parameter, DataType: jack.Char}, 0, false)

		st.PopSubroutineScope() // Push a new subroutine scope before doing anything

		// All of these variables should be found and resolved correctly
		test(st, "test1", jack.Variable{Name: "test1", Type: jack.Field, DataType: jack.Int}, 0, false)
		test(st, "test2", jack.Variable{Name: "test2", Type: jack.Static, DataType: jack.String}, 0, false)
	})
}

func TestScopeTracking(t *testing.T) {
	test := func(st jack.ScopeTable, expected string, fail bool) {
		scope := st.GetScope()
		if scope != expected && !fail {
			t.Errorf("expected to get scope %s, got %+v", expected, scope)
		}
	}

	t.Run("Basic scope tracking checks", func(t *testing.T) {
		st := jack.ScopeTable{}

		st.PushClassScope("TestClass") // Push a new class scope before doing anything
		test(st, "TestClass.Global", false)

		st.PushSubRoutineScope("TestSubroutine") // Push a new subroutine scope before doing anything
		test(st, "TestClass.TestSubroutine", false)

		st.PopSubroutineScope() // Deallocates the current subroutine scope
		test(st, "TestClass.Global", false)

		st.PopClassScope() // Deallocates the current class scope
		test(st, "Global", false)
	})
}
