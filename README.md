# prologo

A Go library for detecting function prologues in ELF binaries by disassembling and analyzing instruction patterns.

## Overview

**prologo** is a Go library that disassembles the `.text` section of ELF executables and identifies common function prologue patterns used by compilers.

## Features

- **ELF Binary Support**: Parses and analyzes ELF format executables
- **Multiple Prologue Detection**: Recognizes various function entry patterns
- **Pattern Classification**: Labels detected prologues by type
- **Fast Analysis**: Efficient Go-based disassembly

## Supported Architectures

- **x86_64** (AMD64) - Full support

## Supported Function Prologues

prologo detects the following x86_64 function prologue patterns:

### 1. Classic Frame Pointer Setup
```asm
push rbp        ; Save caller's frame pointer
mov rbp, rsp    ; Set up new frame pointer
```

**Detection Label**: `classic`

**Description**: Traditional prologue that establishes a complete stack frame with base pointer.

- Used by: Non-optimized builds (`-O0`), code with `-fno-omit-frame-pointer`

---

### 2. No-Frame-Pointer Function
```asm
sub rsp, 0x20   ; Allocate stack space directly
```

**Detection Label**: `no-frame-pointer`

**Description**: Optimized prologue that skips frame pointer setup entirely.

- Used by: Optimized builds (`-O2`, `-O3`), `-fomit-frame-pointer` flag

---

### 3. Push-Only Prologue
```asm
push rbp        ; Save frame pointer only
```

**Detection Label**: `push-only`

**Description**: Minimal prologue that saves RBP but doesn't establish a frame chain.

---

### 4. LEA-Based Stack Allocation
```asm
lea rsp, [rsp-0x20]   ; Allocate using LEA instead of SUB
```

**Detection Label**: `lea-based`

**Description**: Alternative stack allocation using LEA instruction. LEA doesn't modify CPU flags (unlike SUB).

---

## Usage

Import prologo in your Go project:

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/maxgio92/prologo"
)

func main() {
    f, err := os.Open("./myapp")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    prologues, err := prologo.DetectProloguesFromELF(f)
    if err != nil {
        log.Fatal(err)
    }

    counts := make(map[prologo.PrologueType]int)
    for _, p := range prologues {
        fmt.Printf("[%s] 0x%x: %s\n", p.Type, p.Address, p.Instructions)
        counts[p.Type]++
    }

    fmt.Println()
    for typ, n := range counts {
        fmt.Printf("  %s: %d\n", typ, n)
    }
    fmt.Printf("Total: %d\n", len(prologues))
}
```

#### Example output

```
[classic] 0x401000: push rbp; mov rbp, rsp
[classic] 0x401020: push rbp; mov rbp, rsp
[no-frame-pointer] 0x401040: sub rsp, 0x20
[push-only] 0x401060: push rbp
...

  classic: 1547
  no-frame-pointer: 1
  push-only: 2
Total: 1550
```

#### API Reference

**Functions:**

```go
// Core detection — works on raw machine code bytes, no I/O.
func DetectPrologues(code []byte, baseAddr uint64) ([]Prologue, error)

// Convenience wrapper — parses ELF from the reader, extracts .text, calls DetectPrologues.
func DetectProloguesFromELF(r io.ReaderAt) ([]Prologue, error)
```

**Types:**

```go
type PrologueType string

const (
    PrologueClassic   PrologueType = "classic"
    PrologueNoFramePointer PrologueType = "no-frame-pointer"
    ProloguePushOnly  PrologueType = "push-only"
    PrologueLEABased  PrologueType = "lea-based"
)

type Prologue struct {
    Address      uint64       `json:"address"`
    Type         PrologueType `json:"type"`
    Instructions string       `json:"instructions"`
}
```

`DetectPrologues` accepts raw bytes and a base virtual address, making it format-agnostic (works with ELF, PE, Mach-O, raw dumps).

`DetectProloguesFromELF` accepts an `io.ReaderAt` (e.g. `*os.File`) and handles ELF parsing internally.

## Implementation Details

### Architecture

```
┌─────────────────┐
│   ELF Binary    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  ELF Parser     │ ← debug/elf package
│  (.text section)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Disassembler   │ ← golang.org/x/arch/x86/x86asm
│  (x86_64 decode)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Pattern Matcher │ ← Prologue detection logic
│  (1-2 inst seq) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  []Prologue     │
│ (addr + type)   │
└─────────────────┘
```

### Limitations

- **No Symbol Information**: Works on stripped binaries but reports addresses only
- **Heuristic-Based**: May have false positives in data sections or inline data
- **Linear Disassembly**: Doesn't handle indirect jumps or computed addresses
- **No CET Support**: ENDBR64 detection not yet implemented

## Dependencies

- **Go 1.21+**
- [`golang.org/x/arch/x86/x86asm`](https://pkg.go.dev/golang.org/x/arch/x86/x86asm) - x86 disassembler
- `debug/elf` (standard library) - ELF parser

## References

- [System V AMD64 ABI](https://refspecs.linuxbase.org/elf/x86_64-abi-0.99.pdf)
- [Intel 64 and IA-32 Architectures Software Developer Manuals](https://www.intel.com/content/www/us/en/developer/articles/technical/intel-sdm.html)
- [Go x86 Assembler](https://pkg.go.dev/golang.org/x/arch/x86/x86asm)
- [ELF Format Specification](https://refspecs.linuxfoundation.org/elf/elf.pdf)

---

**Author**: Massimiliano Giovagnoli ([@maxgio92](https://github.com/maxgio92))
