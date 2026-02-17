// Package resurgo identifies functions in executable binaries through static
// analysis using instruction-level disassembly. It provides two complementary
// approaches:
//
// # Prologue Detection
//
// Detects function entry points by recognizing common prologue patterns including
// classic frame pointer setup (push rbp; mov rbp, rsp), no-frame-pointer functions
// (sub rsp, imm), push-only prologues, and LEA-based stack allocation.
//
// Use [DetectPrologues] to analyze raw bytes directly, or
// [DetectProloguesFromELF] to extract and analyze the .text section of an ELF binary.
//
// # Call Site Analysis
//
// Identifies functions by detecting CALL and JMP instructions and extracting their
// target addresses. This approach works on heavily optimized code where prologues
// may be omitted and is compiler-agnostic. It provides confidence scoring based on
// instruction type (call vs jump) and addressing mode (direct vs indirect).
//
// Use [DetectCallSites] to analyze raw bytes, or [DetectCallSitesFromELF]
// for ELF binaries. Results are filtered to only include targets within the
// .text section.
//
// # Combined Analysis
//
// For highest-confidence function detection, use [DetectFunctions] which combines
// both prologue and call site analysis. Functions detected by both methods
// receive the highest confidence rating. This is particularly effective for
// recovering functions in stripped binaries or heavily optimized code.
//
// # Confidence Scoring
//
// The confidence level indicates the reliability of a detection:
//   - High: Direct CALL instructions or prologue + called/jumped to
//   - Medium: Unconditional JMP or prologue-only
//   - Low: Conditional jumps (usually intra-function branches)
//   - None: Register-indirect (cannot be statically resolved)
package resurgo
