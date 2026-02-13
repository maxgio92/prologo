# prologo

A Go library for static function recovery from executable binaries.

It works with raw bytes from any binary format as well as parsing ELF files.

## Features

- **Prologue-Based Detection**: Recognizes common function entry patterns by instruction analysis
- **Format-Agnostic Core**: Works on raw machine code bytes from any binary format
- **ELF Convenience Wrapper**: Built-in support for parsing ELF executables
- **Pattern Classification**: Labels detected prologues by type (classic, no-frame-pointer, push-only, lea-based)

## Supported architectures

- **x86_64** (AMD64)

## Function Metadata

### Prologues

Prologues are one of the metadata that prologo recovers about functions. They are detected by recognizing common instruction patterns at function entry points.

#### 1. Classic Frame Pointer Setup (`classic`)

```asm
push rbp        ; Save caller's frame pointer
mov rbp, rsp    ; Set up new frame pointer
```
Traditional prologue that establishes a complete stack frame with base pointer.
Used by: Non-optimized builds (`-O0`), code with `-fno-omit-frame-pointer`

#### 2. No-Frame-Pointer Function (`no-frame-pointer`)

```asm
sub rsp, 0x20   ; Allocate stack space directly
```
Optimized prologue that skips frame pointer setup entirely.
Used by: Optimized builds (`-O2`, `-O3`), `-fomit-frame-pointer` flag

#### 3. Push-Only Prologue (`push-only`)

```asm
push rbp        ; Save frame pointer only
```
Minimal prologue that saves RBP but doesn't establish a frame chain.

#### 4. LEA-Based Stack Allocation (`lea-based`)

```asm
lea rsp, [rsp-0x20]   ; Allocate using LEA instead of SUB
```
Alternative stack allocation using LEA instruction. LEA doesn't modify CPU flags (unlike SUB).

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

    for _, p := range prologues {
        fmt.Printf("[%s] 0x%x: %s\n", p.Type, p.Address, p.Instructions)
    }
}
```

#### Example output

```
[classic] 0x401000: push rbp; mov rbp, rsp
[classic] 0x401020: push rbp; mov rbp, rsp
[no-frame-pointer] 0x401040: sub rsp, 0x20
[push-only] 0x401060: push rbp
```

## API Reference

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
│   (ASM decode)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Pattern Matcher │ ← Prologue detection logic
│   (insns seq)   │
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

