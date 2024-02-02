package hclwrite

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TokensForExpression(expr hclsyntax.Expression) Tokens {
	tks := tokensForExpression(expr, false)

	ret := make(Tokens, len(tks))
	for i := range tks {
		ret[i] = &Token{
			Type:  tks[i].Type,
			Bytes: tks[i].Bytes,
		}
	}
	return ret
}

func tokensForExpression(expr hclsyntax.Expression, quoted bool) (tks hclsyntax.Tokens) {
	switch e := expr.(type) {
	case *hclsyntax.AnonSymbolExpr:
		return TokensForAnonSymbolExpr(e)
	case *hclsyntax.BinaryOpExpr:
		return TokensForBinaryOpExpr(e)
	case *hclsyntax.ConditionalExpr:
		return TokensForConditionalExpr(e)
	case *hclsyntax.ForExpr:
		return TokensForForExpr(e)
	case *hclsyntax.FunctionCallExpr:
		return TokensForFunctionCallExpr(e)
	case *hclsyntax.IndexExpr:
		return TokensForIndexExpr(e)
	case *hclsyntax.LiteralValueExpr:
		return TokensForLiteralValueExpr(e, quoted)
	case *hclsyntax.ObjectConsExpr:
		return TokensForObjectConsExpr(e)
	case *hclsyntax.ObjectConsKeyExpr:
		return TokensForObjectConsKeyExpr(e)
	case *hclsyntax.RelativeTraversalExpr:
		return TokensForRelativeTraversalExpr(e)
	case *hclsyntax.ScopeTraversalExpr:
		return TokensForScopeTraversalExpr(e)
	case *hclsyntax.SplatExpr:
		return TokensForSplatExpr(e)
	case *hclsyntax.TemplateExpr:
		return TokensForTemplateExpr(e, quoted)
	case *hclsyntax.TemplateJoinExpr:
		return TokensForTemplateJoinExpr(e)
	case *hclsyntax.TemplateWrapExpr:
		return TokensForTemplateWrapExpr(e)
	case *hclsyntax.TupleConsExpr:
		return TokensForTupleConsExpr(e)
	case *hclsyntax.UnaryOpExpr:
		return TokensForUnaryOpExpr(e)
	case *hclsyntax.ParenthesesExpr:
		return tokensForExpression(e.Expression, quoted)
	}

	return
}

func TokensForAnonSymbolExpr(e *hclsyntax.AnonSymbolExpr) (tks hclsyntax.Tokens) {
	return
}

func TokensForBinaryOpExpr(e *hclsyntax.BinaryOpExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, tokensForExpression(e.LHS, false)...)

	switch e.Op {
	case hclsyntax.OpLogicalOr:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenOr,
			Bytes: []byte("||"),
		})
	case hclsyntax.OpLogicalAnd:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenAnd,
			Bytes: []byte("&&"),
		})
	case hclsyntax.OpEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenEqualOp,
			Bytes: []byte("=="),
		})
	case hclsyntax.OpNotEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenNotEqual,
			Bytes: []byte("!="),
		})
	case hclsyntax.OpGreaterThan:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenGreaterThan,
			Bytes: []byte{'>'},
		})
	case hclsyntax.OpGreaterThanOrEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenGreaterThanEq,
			Bytes: []byte(">="),
		})
	case hclsyntax.OpLessThan:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenLessThan,
			Bytes: []byte{'<'},
		})
	case hclsyntax.OpLessThanOrEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenLessThanEq,
			Bytes: []byte("<="),
		})
	case hclsyntax.OpAdd:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenPlus,
			Bytes: []byte{'+'},
		})
	case hclsyntax.OpSubtract:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenMinus,
			Bytes: []byte{'-'},
		})
	case hclsyntax.OpMultiply:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenStar,
			Bytes: []byte{'*'},
		})
	case hclsyntax.OpDivide:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenSlash,
			Bytes: []byte{'/'},
		})
	case hclsyntax.OpModulo:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenPercent,
			Bytes: []byte{'%'},
		})
	}

	tks = append(tks, tokensForExpression(e.RHS, false)...)

	return
}

func TokensForConditionalExpr(e *hclsyntax.ConditionalExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, tokensForExpression(e.Condition, false)...)

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenQuestion,
		Bytes: []byte{'?'},
	})

	tks = append(tks, tokensForExpression(e.TrueResult, false)...)

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenColon,
		Bytes: []byte{':'},
	})

	tks = append(tks, tokensForExpression(e.FalseResult, false)...)

	return
}

func TokensForForExpr(e *hclsyntax.ForExpr) (tks hclsyntax.Tokens) {
	if e.KeyExpr == nil {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte{'['},
		})
	} else {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenOBrace,
			Bytes: []byte{'{'},
		})
	}

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte{'\n'},
	})

	tks = append(tks,
		hclsyntax.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("for"),
		},
		hclsyntax.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(e.KeyVar),
		})

	if e.ValVar != "" {
		tks = append(tks,
			hclsyntax.Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte{','},
			},
			hclsyntax.Token{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(e.ValVar),
			})
	}

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenIdent,
		Bytes: []byte("in"),
	})

	tks = append(tks, tokensForExpression(e.CollExpr, false)...)

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenColon,
		Bytes: []byte{':'},
	})

	if e.KeyExpr != nil {
		tks = append(tks, tokensForExpression(e.KeyExpr, false)...)

		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenFatArrow,
			Bytes: []byte("=>"),
		})
	}

	tks = append(tks, tokensForExpression(e.ValExpr, false)...)

	if e.Group {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenEllipsis,
			Bytes: []byte("..."),
		})
	}

	if e.CondExpr != nil {
		tks = append(tks,
			hclsyntax.Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte{'\n'},
			},
			hclsyntax.Token{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte("if"),
			})

		tks = append(tks, tokensForExpression(e.CondExpr, false)...)
	}

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenNewline,
		Bytes: []byte{'\n'},
	})

	if e.KeyExpr == nil {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte{']'},
		})
	} else {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenCBrace,
			Bytes: []byte{'}'},
		})
	}

	return
}

func TokensForFunctionCallExpr(e *hclsyntax.FunctionCallExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenIdent,
		Bytes: []byte(e.Name),
	})

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenOParen,
		Bytes: []byte{'('},
	})

	for i, arg := range e.Args {
		if i > 0 {
			tks = append(tks, hclsyntax.Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte{','},
			})
		}

		tks = append(tks, tokensForExpression(arg, false)...)

		if e.ExpandFinal && i == len(e.Args)-1 {
			tks = append(tks, hclsyntax.Token{
				Type:  hclsyntax.TokenEllipsis,
				Bytes: []byte("..."),
			})
		}
	}

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenCParen,
		Bytes: []byte{')'},
	})

	return
}

func TokensForIndexExpr(e *hclsyntax.IndexExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, tokensForExpression(e.Collection, true)...)

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenOBrack,
		Bytes: []byte{'['},
	})

	tks = append(tks, tokensForExpression(e.Key, false)...)

	tks = append(tks,
		hclsyntax.Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte{']'},
		})

	return
}

func TokensForLiteralValueExpr(e *hclsyntax.LiteralValueExpr, quoted bool) (tks hclsyntax.Tokens) {
	ss := TokensForValue(e.Val)

	for i, s := range ss {
		if quoted && e.Val.Type() == cty.String && (i == 0 || i == len(ss)-1) {
			continue
		}

		tks = append(tks, hclsyntax.Token{
			Type:  s.Type,
			Bytes: s.Bytes,
		})
	}

	return
}

func TokensForObjectConsExpr(e *hclsyntax.ObjectConsExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenOBrace,
		Bytes: []byte{'{'},
	})

	for i := range e.Items {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})

		tks = append(tks, tokensForExpression(e.Items[i].KeyExpr, false)...)

		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		})

		tks = append(tks, tokensForExpression(e.Items[i].ValueExpr, false)...)
	}

	if len(e.Items) > 0 {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenCBrace,
		Bytes: []byte{'}'},
	})

	return
}

func TokensForObjectConsKeyExpr(e *hclsyntax.ObjectConsKeyExpr) (tks hclsyntax.Tokens) {
	return tokensForExpression(e.Wrapped, false)
}

func TokensForRelativeTraversalExpr(e *hclsyntax.RelativeTraversalExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, tokensForExpression(e.Source, true)...)

	for _, s := range TokensForTraversal(e.Traversal) {
		tks = append(tks, hclsyntax.Token{
			Type:  s.Type,
			Bytes: s.Bytes,
		})
	}

	return
}

func TokensForScopeTraversalExpr(e *hclsyntax.ScopeTraversalExpr) (tks hclsyntax.Tokens) {
	for _, s := range TokensForTraversal(e.Traversal) {
		tks = append(tks, hclsyntax.Token{
			Type:  s.Type,
			Bytes: s.Bytes,
		})
	}

	return
}

func TokensForSplatExpr(e *hclsyntax.SplatExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, tokensForExpression(e.Source, true)...)

	tks = append(tks,
		hclsyntax.Token{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte{'['},
		},
		hclsyntax.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte{'*'},
		},
		hclsyntax.Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte{']'},
		})

	return
}

func TokensForTemplateExpr(e *hclsyntax.TemplateExpr, quoted bool) (tks hclsyntax.Tokens) {
	if !quoted {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenOQuote,
			Bytes: []byte{'"'},
		})
	}

	for i := range e.Parts {
		switch t := e.Parts[i].(type) {
		case *hclsyntax.LiteralValueExpr:
			tks = append(tks, TokensForLiteralValueExpr(t, true)...)
		case *hclsyntax.ScopeTraversalExpr:
			tks = append(tks, hclsyntax.Token{
				Type:  hclsyntax.TokenTemplateInterp,
				Bytes: []byte("${"),
			})
			tks = append(tks, TokensForScopeTraversalExpr(t)...)
			tks = append(tks, hclsyntax.Token{
				Type:  hclsyntax.TokenTemplateSeqEnd,
				Bytes: []byte{'}'},
			})
		case *hclsyntax.ConditionalExpr:
			tks = append(tks,
				hclsyntax.Token{
					Type:  hclsyntax.TokenTemplateControl,
					Bytes: []byte("%{"),
				},
				hclsyntax.Token{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("if"),
				})
			tks = append(tks, tokensForExpression(t.Condition, true)...)
			tks = append(tks, hclsyntax.Token{
				Type:  hclsyntax.TokenTemplateSeqEnd,
				Bytes: []byte{'}'},
			})
			tks = append(tks, tokensForExpression(t.TrueResult, true)...)

			if t.FalseResult != nil {
				ss := tokensForExpression(t.FalseResult, true)
				if len(ss) != 0 {
					tks = append(tks,
						hclsyntax.Token{
							Type:  hclsyntax.TokenTemplateControl,
							Bytes: []byte("%{"),
						},
						hclsyntax.Token{
							Type:  hclsyntax.TokenStringLit,
							Bytes: []byte("else"),
						},
						hclsyntax.Token{
							Type:  hclsyntax.TokenTemplateSeqEnd,
							Bytes: []byte{'}'},
						})
					tks = append(tks, ss...)
				}
			}

			tks = append(tks,
				hclsyntax.Token{
					Type:  hclsyntax.TokenTemplateControl,
					Bytes: []byte("%{"),
				},
				hclsyntax.Token{
					Type:  hclsyntax.TokenStringLit,
					Bytes: []byte("endif"),
				},
				hclsyntax.Token{
					Type:  hclsyntax.TokenTemplateSeqEnd,
					Bytes: []byte{'}'},
				})
		}
	}

	if !quoted {
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenCQuote,
			Bytes: []byte{'"'},
		})
	}

	return
}

func TokensForTemplateJoinExpr(e *hclsyntax.TemplateJoinExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenOBrack,
		Bytes: []byte("["),
	})

	tks = append(tks, tokensForExpression(e.Tuple, false)...)

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenCBrack,
		Bytes: []byte("]"),
	})

	return
}

func TokensForTemplateWrapExpr(e *hclsyntax.TemplateWrapExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenTemplateInterp,
		Bytes: []byte("${"),
	})

	tks = append(tks, tokensForExpression(e.Wrapped, false)...)

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenTemplateSeqEnd,
		Bytes: []byte("}"),
	})

	return
}

func TokensForTupleConsExpr(e *hclsyntax.TupleConsExpr) (tks hclsyntax.Tokens) {
	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenOBrack,
		Bytes: []byte{'['},
	})

	for i := range e.Exprs {
		if i > 0 {
			tks = append(tks, hclsyntax.Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte{','},
			})
		}

		tks = append(tks, tokensForExpression(e.Exprs[i], false)...)
	}

	tks = append(tks, hclsyntax.Token{
		Type:  hclsyntax.TokenCBrack,
		Bytes: []byte{']'},
	})

	return
}

func TokensForUnaryOpExpr(e *hclsyntax.UnaryOpExpr) (tks hclsyntax.Tokens) {
	switch e.Op {
	case hclsyntax.OpLogicalNot:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenBang,
			Bytes: []byte{'!'},
		})
	case hclsyntax.OpEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		})
	case hclsyntax.OpGreaterThan:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenGreaterThan,
			Bytes: []byte{'>'},
		})
	case hclsyntax.OpGreaterThanOrEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenGreaterThanEq,
			Bytes: []byte(">="),
		})
	case hclsyntax.OpLessThan:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenLessThan,
			Bytes: []byte{'<'},
		})
	case hclsyntax.OpLessThanOrEqual:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenLessThanEq,
			Bytes: []byte("<="),
		})
	case hclsyntax.OpNegate:
		tks = append(tks, hclsyntax.Token{
			Type:  hclsyntax.TokenMinus,
			Bytes: []byte{'-'},
		})
	}

	tks = append(tks, tokensForExpression(e.Val, false)...)

	return
}
