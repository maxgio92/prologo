package prologo_test

import (
	"fmt"

	"github.com/maxgio92/prologo"
)

func ExampleDetectPrologues() {
	// x86-64 machine code: nop; push rbp; mov rbp, rsp
	code := []byte{0x90, 0x55, 0x48, 0x89, 0xe5}
	prologues := prologo.DetectPrologues(code, 0x1000)
	for _, p := range prologues {
		fmt.Printf("[%s] 0x%x: %s\n", p.Type, p.Address, p.Instructions)
	}
	// Output:
	// [classic] 0x1001: push rbp; mov rbp, rsp
}
