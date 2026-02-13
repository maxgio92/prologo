package prologo

import (
	"debug/elf"
	"fmt"
	"io"

	"golang.org/x/arch/x86/x86asm"
)

// DetectPrologues analyzes raw machine code bytes and returns detected function
// prologues. baseAddr is the virtual address corresponding to the start of code.
// This function performs no I/O and works with any binary format.
func DetectPrologues(code []byte, baseAddr uint64) []Prologue {
	var result []Prologue

	offset := 0
	addr := baseAddr
	var prevInsn *x86asm.Inst

	for offset < len(code) {
		inst, err := x86asm.Decode(code[offset:], 64)
		if err != nil {
			offset++
			addr++
			prevInsn = nil
			continue
		}

		// Pattern 1: Classic frame pointer setup - push rbp; mov rbp, rsp
		if prevInsn != nil &&
			prevInsn.Op == x86asm.PUSH && prevInsn.Args[0] == x86asm.RBP &&
			inst.Op == x86asm.MOV && inst.Args[0] == x86asm.RBP && inst.Args[1] == x86asm.RSP {
			result = append(result, Prologue{
				Address:      addr - uint64(prevInsn.Len),
				Type:         PrologueClassic,
				Instructions: "push rbp; mov rbp, rsp",
			})
		}

		// Pattern 2: No-frame-pointer function - sub rsp, imm
		if inst.Op == x86asm.SUB && inst.Args[0] == x86asm.RSP {
			if imm, ok := inst.Args[1].(x86asm.Imm); ok && imm > 0 {
				if prevInsn == nil || prevInsn.Op == x86asm.RET {
					result = append(result, Prologue{
						Address:      addr,
						Type:         PrologueNoFramePointer,
						Instructions: fmt.Sprintf("sub rsp, 0x%x", imm),
					})
				}
			}
		}

		// Pattern 3: Push rbp as first instruction
		if inst.Op == x86asm.PUSH && inst.Args[0] == x86asm.RBP {
			if prevInsn == nil || prevInsn.Op == x86asm.RET {
				result = append(result, Prologue{
					Address:      addr,
					Type:         ProloguePushOnly,
					Instructions: "push rbp",
				})
			}
		}

		// Pattern 4: Stack allocation with lea - lea rsp, [rsp-imm]
		if inst.Op == x86asm.LEA && inst.Args[0] == x86asm.RSP {
			if prevInsn == nil || prevInsn.Op == x86asm.RET {
				result = append(result, Prologue{
					Address:      addr,
					Type:         PrologueLEABased,
					Instructions: "lea rsp, [rsp-offset]",
				})
			}
		}

		prevInsn = &inst
		offset += inst.Len
		addr += uint64(inst.Len)
	}

	return result
}

// DetectProloguesFromELF parses an ELF binary from the given reader, extracts
// the .text section, and returns detected function prologues.
func DetectProloguesFromELF(r io.ReaderAt) ([]Prologue, error) {
	f, err := elf.NewFile(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ELF file: %w", err)
	}
	defer f.Close()

	textSec := f.Section(".text")
	if textSec == nil {
		return nil, fmt.Errorf("no .text section found")
	}

	code, err := textSec.Data()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read .text section: %w", err)
	}

	return DetectPrologues(code, textSec.Addr), nil
}
