package main

import (
	"fmt"
	"os"
	"testing"
)

func TestHackAssembler(t *testing.T) {
	test := func(input string, output string, compare string) {
		status := Handler([]string{input, output}, nil)
		if status != 0 {
			t.Fatalf("Unexpected exit status code: expected 0 got: %d", status)
		}

		compiledContent, err := os.ReadFile(output)
		if err != nil {
			t.Fatalf("Error reading output file %s: %v", output, err)
		}

		expectedContent, err := os.ReadFile(compare)
		if err != nil {
			t.Fatalf("Error reading compare file %s: %v", compare, err)
		}

		if string(compiledContent) != string(expectedContent) {
			t.Fatal("Output and compare file contents do not match")
		}
	}

	t.Run("Add.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/01 - Add"
		input := fmt.Sprintf("%s/%s", base, "Add.asm")
		output := fmt.Sprintf("%s/%s", base, "Add.hack")
		compare := fmt.Sprintf("%s/%s", base, "Add.cmp")
		test(input, output, compare)
	})

	t.Run("Max.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/02 - Max"
		input := fmt.Sprintf("%s/%s", base, "Max.asm")
		output := fmt.Sprintf("%s/%s", base, "Max.hack")
		compare := fmt.Sprintf("%s/%s", base, "Max.cmp")
		test(input, output, compare)
	})

	t.Run("MaxL.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/02 - Max"
		input := fmt.Sprintf("%s/%s", base, "MaxL.asm")
		output := fmt.Sprintf("%s/%s", base, "MaxL.hack")
		compare := fmt.Sprintf("%s/%s", base, "MaxL.cmp")
		test(input, output, compare)
	})

	t.Run("Rect.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/03 - Rect"
		input := fmt.Sprintf("%s/%s", base, "Rect.asm")
		output := fmt.Sprintf("%s/%s", base, "Rect.hack")
		compare := fmt.Sprintf("%s/%s", base, "Rect.cmp")
		test(input, output, compare)
	})

	t.Run("RectL.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/03 - Rect"
		input := fmt.Sprintf("%s/%s", base, "RectL.asm")
		output := fmt.Sprintf("%s/%s", base, "RectL.hack")
		compare := fmt.Sprintf("%s/%s", base, "RectL.cmp")
		test(input, output, compare)
	})

	t.Run("Pong.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/04 - Pong"
		input := fmt.Sprintf("%s/%s", base, "Pong.asm")
		output := fmt.Sprintf("%s/%s", base, "Pong.hack")
		compare := fmt.Sprintf("%s/%s", base, "Pong.cmp")
		test(input, output, compare)
	})

	t.Run("PongL.asm", func(t *testing.T) {
		base := "../../../projects/06 - Assembler/04 - Pong"
		input := fmt.Sprintf("%s/%s", base, "PongL.asm")
		output := fmt.Sprintf("%s/%s", base, "PongL.hack")
		compare := fmt.Sprintf("%s/%s", base, "PongL.cmp")
		test(input, output, compare)
	})
}
