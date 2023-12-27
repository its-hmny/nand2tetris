package main

import (
	"fmt"

	"its-hmny.dev/n2t-assembler/hack"
)

func main() {
	fmt.Println("Hello world from N2T Hack Assembler")

	test1 := hack.AInstruction{LocType: hack.Label, LocName: "Test"}
	fmt.Printf("%T: %+v\n", test1, test1)

	test2 := hack.CInstruction{}
	fmt.Printf("%T: %+v\n", test2, test2)
}
