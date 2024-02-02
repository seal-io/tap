package hclsyntax

import "github.com/hashicorp/hcl/v2"

func (t Token) AsHCLToken() hcl.Token {
	return hcl.Token{
		Type:  rune(t.Type),
		Bytes: t.Bytes,
	}
}

func (ts Tokens) AsHCLTokens() hcl.Tokens {
	ret := make(hcl.Tokens, len(ts))
	for i, t := range ts {
		ret[i] = t.AsHCLToken()
	}
	return ret
}
