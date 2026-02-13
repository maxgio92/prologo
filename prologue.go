package prologo

// PrologueType represents the type of function prologue.
type PrologueType string

// Recognized function prologue patterns.
const (
	PrologueClassic   PrologueType = "classic"
	PrologueNoFramePointer PrologueType = "no-frame-pointer"
	ProloguePushOnly  PrologueType = "push-only"
	PrologueLEABased  PrologueType = "lea-based"
)

// Prologue represents a detected function prologue.
type Prologue struct {
	Address      uint64       `json:"address"`
	Type         PrologueType `json:"type"`
	Instructions string       `json:"instructions"`
}
