package hcl

// Token represents a sequence of bytes from some HCL code that has been
// tagged with a type and its range within the source file.
type (
	Token struct {
		Type  rune
		Bytes []byte
	}

	Tokens []Token
)
