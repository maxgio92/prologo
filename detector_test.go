package prologo_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/maxgio92/prologo"
)

const (
	demoAppSource = "testdata/demo-app.go"
	demoAppBinary = "demo-app"
)

func TestDetectPrologues(t *testing.T) {
	tests := []struct {
		name      string
		code      []byte
		baseAddr  uint64
		wantCount int
		wantType  prologo.PrologueType
		wantAddr  uint64
	}{
		{
			// nop; push rbp; mov rbp, rsp
			// The leading nop ensures push rbp is not at start-of-input,
			// so only the classic pattern fires.
			name:      string(prologo.PrologueClassic),
			code:      []byte{0x90, 0x55, 0x48, 0x89, 0xe5},
			baseAddr:  0,
			wantCount: 1,
			wantType:  prologo.PrologueClassic,
			wantAddr:  1,
		},
		{
			// sub rsp, 0x20 at start of code (no preceding instruction)
			name:      string(prologo.PrologueNoFramePointer),
			code:      []byte{0x48, 0x83, 0xec, 0x20},
			baseAddr:  0,
			wantCount: 1,
			wantType:  prologo.PrologueNoFramePointer,
			wantAddr:  0,
		},
		{
			// push rbp; nop — push rbp at start, not followed by mov rbp, rsp
			name:      string(prologo.ProloguePushOnly),
			code:      []byte{0x55, 0x90},
			baseAddr:  0,
			wantCount: 1,
			wantType:  prologo.ProloguePushOnly,
			wantAddr:  0,
		},
		{
			name:      "EmptyNil",
			code:      nil,
			wantCount: 0,
		},
		{
			name:      "EmptySlice",
			code:      []byte{},
			wantCount: 0,
		},
		{
			// Garbage bytes that should not match any prologue pattern.
			name:      "InvalidBytes",
			code:      []byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prologues := prologo.DetectPrologues(tt.code, tt.baseAddr)

			if len(prologues) != tt.wantCount {
				t.Fatalf("expected %d prologue(s), got %d: %+v", tt.wantCount, len(prologues), prologues)
			}
			if tt.wantCount == 0 {
				return
			}
			if prologues[0].Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, prologues[0].Type)
			}
			if prologues[0].Address != tt.wantAddr {
				t.Errorf("expected address 0x%x, got 0x%x", tt.wantAddr, prologues[0].Address)
			}
		})
	}
}

func TestDetectProloguesFromELF(t *testing.T) {
	tests := []struct {
		name      string
		buildArgs []string                        // extra args after "go build -o <path>"
		minCounts map[prologo.PrologueType]int // minimum prologues per type
	}{
		{
			name:      "optimized",
			buildArgs: nil,
			minCounts: map[prologo.PrologueType]int{
				prologo.PrologueClassic:   1,
				prologo.PrologueNoFramePointer: 1,
			},
		},
		{
			name:      "unoptimized",
			buildArgs: []string{"-gcflags=all=-N -l"},
			minCounts: map[prologo.PrologueType]int{
				prologo.PrologueClassic: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binPath := filepath.Join(t.TempDir(), demoAppBinary)
			args := append([]string{"build", "-o", binPath}, tt.buildArgs...)
			args = append(args, demoAppSource)

			cmd := exec.Command("go", args...)
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("failed to compile demo-app: %v\n%s", err, out)
			}

			f, err := os.Open(binPath)
			if err != nil {
				t.Fatalf("failed to open compiled binary: %v", err)
			}
			defer f.Close()

			prologues, err := prologo.DetectProloguesFromELF(f)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(prologues) == 0 {
				t.Fatal("expected at least one prologue, got none")
			}

			counts := make(map[prologo.PrologueType]int)
			for _, p := range prologues {
				counts[p.Type]++
			}
			t.Logf("total prologues: %d, by type: %v", len(prologues), counts)

			for typ, min := range tt.minCounts {
				if counts[typ] < min {
					t.Errorf("expected at least %d %s prologue(s), got %d", min, typ, counts[typ])
				}
			}
		})
	}
}

func TestDetectProloguesFromELF_InvalidReader(t *testing.T) {
	r := bytes.NewReader([]byte{0x00, 0x01, 0x02, 0x03})
	_, err := prologo.DetectProloguesFromELF(r)
	if err == nil {
		t.Fatal("expected error for invalid ELF data, got nil")
	}
}
