package hclwrite

import (
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// AppendHCLBody appends an existing hcl.Body (which must not be already attached
// to a body) to the end of the receiving body,
//
// If the given hcl.Body is not *hclsyntax.Body, this method is a no-op.
func (b *Body) AppendHCLBody(body hcl.Body) {
	if body == nil {
		return
	}

	bd, ok := body.(*hclsyntax.Body)
	if !ok {
		return
	}

	attrs := make([]*hclsyntax.Attribute, 0, len(bd.Attributes))
	{
		for k := range bd.Attributes {
			attrs = append(attrs, bd.Attributes[k])
		}

		sort.Slice(attrs, func(i, j int) bool {
			if attrs[i].SrcRange.Filename == attrs[j].SrcRange.Filename {
				return attrs[i].SrcRange.Start.Line < attrs[j].SrcRange.Start.Line
			}

			return attrs[i].SrcRange.Filename < attrs[j].SrcRange.Filename
		})
	}
	for i := range attrs {
		b.SetAttributeRaw(attrs[i].Name, TokensForExpression(attrs[i].Expr))
	}

	for i := range bd.Blocks {
		wb := b.AppendNewBlock(bd.Blocks[i].Type, bd.Blocks[i].Labels).Body()
		wb.AppendHCLBody(bd.Blocks[i].Body)
	}
}
