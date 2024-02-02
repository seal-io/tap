package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func fromHCLTokens(tokens hcl.Tokens, skipNewline bool) hclwrite.Tokens {
	r := make(hclwrite.Tokens, 0, len(tokens))

	for i := range tokens {
		if skipNewline && i == len(tokens)-1 && hclsyntax.TokenType(tokens[i].Type) == hclsyntax.TokenNewline {
			continue
		}

		r = append(r, &hclwrite.Token{
			Type:  hclsyntax.TokenType(tokens[i].Type),
			Bytes: tokens[i].Bytes,
		})
	}

	return r
}
