package main

import (
	"fmt"

	"its-hmny.dev/n2t-assembler/hack"
)

func main() {
	fmt.Println("Hello world from N2T Hack Assembler")

	test1 := hack.AInstruction{LocationType: hack.LocTypeLabel, LocationName: "Test"}
	fmt.Printf("AInstruction(%d): %+v\n", test1.Type(), test1)

	test2 := hack.CInstruction{}
	fmt.Printf("AInstruction(%d): %+v\n", test2.Type(), test2)
}
