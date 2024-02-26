package main

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestVMTranslator(t *testing.T) {
	test := func(input string, output string, test string) {
		status := Handler([]string{input}, map[string]string{"output": output})
		if status != 0 {
			t.Fatalf("Unexpected exit status code: expected 0 got: %d", status)
		}

		cmd := exec.Command("../../../tools/CPUEmulator.sh", test)
		if err := cmd.Run(); err != nil {
			t.Fatalf("Error while running the '%s' test file: %s", test, err)
		}
	}

	t.Run("SimpleAdd.vm", func(t *testing.T) {
		base := "../../../projects/07 - VM I: Stack Arithmetic/01 - SimpleAdd"
		input := fmt.Sprintf("%s/%s", base, "SimpleAdd.vm")
		output := fmt.Sprintf("%s/%s", base, "SimpleAdd.asm")
		tester := fmt.Sprintf("%s/%s", base, "SimpleAdd.tst")
		test(input, output, tester)
	})

	t.Run("StackTest.vm", func(t *testing.T) {
		base := "../../../projects/07 - VM I: Stack Arithmetic/02 - StackTest"
		input := fmt.Sprintf("%s/%s", base, "StackTest.vm")
		output := fmt.Sprintf("%s/%s", base, "StackTest.asm")
		tester := fmt.Sprintf("%s/%s", base, "StackTest.tst")
		test(input, output, tester)
	})

	t.Run("BasicTest.vm", func(t *testing.T) {
		base := "../../../projects/07 - VM I: Stack Arithmetic/03 - BasicTest"
		input := fmt.Sprintf("%s/%s", base, "BasicTest.vm")
		output := fmt.Sprintf("%s/%s", base, "BasicTest.asm")
		tester := fmt.Sprintf("%s/%s", base, "BasicTest.tst")
		test(input, output, tester)
	})

	t.Run("PointerTest.vm", func(t *testing.T) {
		base := "../../../projects/07 - VM I: Stack Arithmetic/04 - PointerTest"
		input := fmt.Sprintf("%s/%s", base, "PointerTest.vm")
		output := fmt.Sprintf("%s/%s", base, "PointerTest.asm")
		tester := fmt.Sprintf("%s/%s", base, "PointerTest.tst")
		test(input, output, tester)
	})

	t.Run("StaticTest.vm", func(t *testing.T) {
		base := "../../../projects/07 - VM I: Stack Arithmetic/05 - StaticTest"
		input := fmt.Sprintf("%s/%s", base, "StaticTest.vm")
		output := fmt.Sprintf("%s/%s", base, "StaticTest.asm")
		tester := fmt.Sprintf("%s/%s", base, "StaticTest.tst")
		test(input, output, tester)
	})

	t.Run("BasicLoop.vm", func(t *testing.T) {
		base := "../../../projects/08 - VM II: Program Flow/01 - BasicLoop"
		input := fmt.Sprintf("%s/%s", base, "BasicLoop.vm")
		output := fmt.Sprintf("%s/%s", base, "BasicLoop.asm")
		tester := fmt.Sprintf("%s/%s", base, "BasicLoop.tst")
		test(input, output, tester)
	})

	t.Run("FibonacciSeries.vm", func(t *testing.T) {
		base := "../../../projects/08 - VM II: Program Flow/02 - FibonacciSeries"
		input := fmt.Sprintf("%s/%s", base, "FibonacciSeries.vm")
		output := fmt.Sprintf("%s/%s", base, "FibonacciSeries.asm")
		tester := fmt.Sprintf("%s/%s", base, "FibonacciSeries.tst")
		test(input, output, tester)
	})

	t.Run("SimpleFunction.vm", func(t *testing.T) {
		base := "../../../projects/08 - VM II: Program Flow/03 - SimpleFunction"
		input := fmt.Sprintf("%s/%s", base, "SimpleFunction.vm")
		output := fmt.Sprintf("%s/%s", base, "SimpleFunction.asm")
		tester := fmt.Sprintf("%s/%s", base, "SimpleFunction.tst")
		test(input, output, tester)
	})
}
