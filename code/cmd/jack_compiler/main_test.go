package main

import (
	"fmt"
	"testing"
)

func TestJackCompiler(t *testing.T) {
	test := func(inputs []string, stdlib bool) {
		options := map[string]string{"stdlib": fmt.Sprint(stdlib)}

		status := Handler(inputs, options)
		if status != 0 {
			t.Fatalf("Unexpected exit status code: expected 0 got: %d", status)
		}
	}

	t.Run("01 - HelloWorld", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/01 - HelloWorld"
		test([]string{input}, true)
	})

	t.Run("02 - Average", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/02 - Average"
		test([]string{input}, true)
	})

	t.Run("03 - Fraction", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/03 - Fraction"
		test([]string{input}, true)
	})

	t.Run("04 - List", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/04 - List"
		test([]string{input}, true)
	})

	t.Run("05 - ConvertToBin", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/05 - ConvertToBin"
		test([]string{input}, true)
	})

	t.Run("06 - ComplexArrays", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/06 - ComplexArrays"
		test([]string{input}, true)
	})

	t.Run("07 - Square", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/07 - Square"
		test([]string{input}, true)
	})

	t.Run("08 - Pong", func(t *testing.T) {
		input := "../../../projects/09 - High-Level Language/08 - Pong"
		test([]string{input}, true)
	})
}
