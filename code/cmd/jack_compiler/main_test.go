package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// This test checks the output of mmy Jack Compiler against the pre generated output of the same Jack Compiler
func TestJackCompiler(t *testing.T) {
	test := func(inputs []string, output string, stdlib bool) {
		options := map[string]string{"stdlib": fmt.Sprint(stdlib)}

		status := Handler(inputs, options)
		if status != 0 {
			t.Fatalf("Unexpected exit status code: expected 0 got: %d", status)
		}

		cmd := exec.Command("git", "diff", output)
		if err := cmd.Run(); err != nil {
			t.Fatalf("The diff between the generated code and the expected one do not match")
		}
	}
	t.Run("ArrayTest", func(t *testing.T) {
		base := "../../../projects/10 - Jack I: Syntax Analysis/01 - ArrayTest"
		test([]string{base}, base, true)
	})

	t.Run("ExpressionLessSquare", func(t *testing.T) {
		base := "../../../projects/10 - Jack I: Syntax Analysis/02 - ExpressionLessSquare"
		test([]string{base}, base, true)
	})

	t.Run("Square", func(t *testing.T) {
		base := "../../../projects/10 - Jack I: Syntax Analysis/03 - Square"
		test([]string{base}, base, true)
	})

	t.Run("Seven", func(t *testing.T) {
		base := "../../../projects/11 - Jack II: Code Generation/01 - Seven"
		test([]string{base}, base, true)
	})

	t.Run("Average", func(t *testing.T) {
		base := "../../../projects/11 - Jack II: Code Generation/02 - Average"
		test([]string{base}, base, true)
	})

	t.Run("ConvertToBin", func(t *testing.T) {
		base := "../../../projects/11 - Jack II: Code Generation/03 - ConvertToBin"
		test([]string{base}, base, true)
	})

	t.Run("ComplexArrays", func(t *testing.T) {
		base := "../../../projects/11 - Jack II: Code Generation/04 - ComplexArrays"
		test([]string{base}, base, true)
	})

	t.Run("Square", func(t *testing.T) {
		base := "../../../projects/11 - Jack II: Code Generation/05 - Square"
		test([]string{base}, base, true)
	})

	t.Run("Pong", func(t *testing.T) {
		base := "../../../projects/11 - Jack II: Code Generation/06 - Pong"
		test([]string{base}, base, true)
	})

}

// This test checks the output of mmy Jack Compiler against the built-in Jack Compiler from
// the Nand2Tetris course. It runs the compiler on various Jack programs and compares the
// generated VM code with the expected output  stored in the corresponding .diff files.
// The test ensures that the Jack Compiler produces the same VM code for every run and
// always has the same changes from the built-in compiler output.
func TestAgainstBuiltIn(t *testing.T) {
	test := func(inputs []string, output string, stdlib bool, test string) {
		options := map[string]string{"stdlib": fmt.Sprint(stdlib)}

		status := Handler(inputs, options)
		if status != 0 {
			t.Fatalf("Unexpected exit status code: expected 0 got: %d", status)
		}

		generated, err := exec.Command("git", "diff", output).Output()
		if err != nil {
			t.Fatalf("Failed to run 'git diff': %v", err)
		}
		expected, err := os.ReadFile(test)
		if err != nil {
			t.Fatalf("Failed to read generated diff: %v", err)
		}
		if string(generated) != string(expected) {
			t.Errorf("The expected diff and the generated one do not match")
		}
	}

	t.Run("HelloWorld", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/01 - HelloWorld"
		tester := fmt.Sprintf("%s/%s", base, "HelloWorld.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("Average", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/02 - Average"
		tester := fmt.Sprintf("%s/%s", base, "Average.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("Fraction", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/03 - Fraction"
		tester := fmt.Sprintf("%s/%s", base, "Fraction.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("List", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/04 - List"
		tester := fmt.Sprintf("%s/%s", base, "List.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("ConvertToBin", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/05 - ConvertToBin"
		tester := fmt.Sprintf("%s/%s", base, "ConvertToBin.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("ComplexArrays", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/06 - ComplexArrays"
		tester := fmt.Sprintf("%s/%s", base, "ComplexArrays.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("Square", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/07 - Square"
		tester := fmt.Sprintf("%s/%s", base, "Square.diff")
		test([]string{base}, base, true, tester)
	})

	t.Run("Pong", func(t *testing.T) {
		base := "../../../projects/09 - High-Level Language/08 - Pong"
		tester := fmt.Sprintf("%s/%s", base, "Pong.diff")
		test([]string{base}, base, true, tester)
	})
}
