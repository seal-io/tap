package tap

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/configs"
)

const (
	dynamicBlockType = "dynamic"
)

type (
	JSONPointerPathToken struct {
		Raw   string
		Value string
	}
	JSONPointerPathOperator []JSONPointerPathToken
)

func NewJSONPointerPathOperator(path string) (PathOperator, error) {
	tks := TokenizeJSONPointerPath(path)
	if len(tks) == 0 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	return JSONPointerPathOperator(tks), nil
}

func (op JSONPointerPathOperator) String() string {
	ss := make([]string, len(op))
	for i := range op {
		ss[i] = op[i].Raw
	}

	return "/" + strings.Join(ss, "/")
}

func (op JSONPointerPathOperator) Search(resource *configs.Resource) (target, parent any, err error) {
	target, ok := resource.Config.(*hclsyntax.Body)
	if !ok {
		return nil, nil, fmt.Errorf("invalid resource type: %T", resource.Config)
	}

	for i := 0; i < len(op)-1; i++ {
		seg := op[i].Value

		switch t := target.(type) {
		default:
			return nil, nil, fmt.Errorf("invalid target type: %s: %T", op[:i+1], target)
		case *hclsyntax.Body:
			if attr, ok := t.Attributes[seg]; ok {
				target = attr.Expr
				parent = t

				continue
			}

			var bds []*hclsyntax.Body

			for j := range t.Blocks {
				switch t.Blocks[j].Type {
				case dynamicBlockType:
					if t.Blocks[j].Labels[0] != seg {
						continue
					}

					bds = append(bds, t.Blocks[j].Body.Blocks[0].Body)
				case seg:
					bds = append(bds, t.Blocks[j].Body)
				}
			}

			if len(bds) == 0 {
				return nil, nil, fmt.Errorf("path not found: %s", op[:i+1])
			}

			target = bds
			parent = t
		case *hclsyntax.ObjectConsExpr:
			var idx int
			for ; idx < len(t.Items); idx++ {
				tr, err := hcl.AbsTraversalForExpr(t.Items[idx].KeyExpr)
				if err == nil && !tr.IsRelative() &&
					tr[0].(hcl.TraverseRoot).Name == seg {
					break
				}
			}

			if idx == len(t.Items) {
				return nil, nil, fmt.Errorf("path not found: %s", op[:i+1])
			}

			target = t.Items[idx].ValueExpr
			parent = t
		case *hclsyntax.TupleConsExpr:
			idx, err := strconv.ParseInt(seg, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid indexer path: %s", op[:i+1])
			}

			if idx == -1 {
				idx = int64(len(t.Exprs) - 1)
			}

			if len(t.Exprs)-1 < int(idx) {
				return nil, nil, fmt.Errorf("path not found: %s", op[:i+1])
			}

			target = t.Exprs[idx]
			parent = t
		case []*hclsyntax.Body:
			idx, err := strconv.ParseInt(seg, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid indexer path: %s", op[:i+1])
			}

			if idx == -1 {
				idx = int64(len(t) - 1)
			}

			if len(t)-1 < int(idx) {
				return nil, nil, fmt.Errorf("path not found: %s", op[:i+1])
			}

			target = t[idx]
		}
	}

	return target, parent, nil
}

func (op JSONPointerPathOperator) Add(resource *configs.Resource, value Value) (*configs.Resource, error) {
	// Search.
	target, _, err := op.Search(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to search target: %w", err)
	}

	// Add.
	seg := op[len(op)-1].Value

	switch t := target.(type) {
	default:
		return nil, fmt.Errorf("invalid target type: %s: %T", op, target)
	case *hclsyntax.Body:
		switch {
		case value.Attribute != nil:
			if t.Attributes == nil {
				t.Attributes = make(map[string]*hclsyntax.Attribute)
			}

			if t.Attributes[seg] != nil {
				return nil, fmt.Errorf("path already exists: %s", op)
			}

			t.Attributes[seg] = &hclsyntax.Attribute{
				Name:      seg,
				Expr:      toHCLSyntaxExpression(value.Attribute.Expr),
				SrcRange:  value.Attribute.Range,
				NameRange: value.Attribute.NameRange,
			}
		case value.Block != nil:
			for j := range t.Blocks {
				switch t.Blocks[j].Type {
				default:
					continue
				case dynamicBlockType:
					if t.Blocks[j].Labels[0] == seg {
						return nil, fmt.Errorf("path already exists: %s", op)
					}
				case seg:
					return nil, fmt.Errorf("path already exists: %s", op)
				}
			}

			t.Blocks = append(t.Blocks, &hclsyntax.Block{
				Type:        seg,
				Labels:      value.Block.Labels,
				Body:        toHCLSyntaxBody(value.Block.Body),
				TypeRange:   value.Block.TypeRange,
				LabelRanges: value.Block.LabelRanges,
			})
		}
	case *hclsyntax.ObjectConsExpr:
		if value.Attribute == nil {
			return nil, errors.New("want patch attribute but got patch block")
		}

		for i := range t.Items {
			tr, err := hcl.AbsTraversalForExpr(t.Items[i].KeyExpr)
			if err == nil && !tr.IsRelative() &&
				tr[0].(hcl.TraverseRoot).Name == seg {
				return nil, fmt.Errorf("path already exists: %s", op)
			}
		}

		t.Items = append(t.Items, hclsyntax.ObjectConsItem{
			KeyExpr: &hclsyntax.ObjectConsKeyExpr{
				Wrapped: &hclsyntax.ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: seg,
						},
					},
				},
			},
			ValueExpr: toHCLSyntaxExpression(value.Attribute.Expr),
		})
	case *hclsyntax.TupleConsExpr:
		if value.Attribute == nil {
			return nil, errors.New("want patch attribute but got patch block")
		}

		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx != -1 {
			return nil, fmt.Errorf("illegal indexer path: %s", op)
		}

		t.Exprs = append(t.Exprs, toHCLSyntaxExpression(value.Attribute.Expr))
	case []*hclsyntax.Body:
		if value.Block == nil {
			return nil, errors.New("want patch block but got patch attribute")
		}

		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t) - 1)
		}

		if len(t)-1 < int(idx) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		bd := toHCLSyntaxBody(value.Block.Body)

		for k := range bd.Attributes {
			if t[idx].Attributes == nil {
				t[idx].Attributes = make(map[string]*hclsyntax.Attribute)
			}

			if t[idx].Attributes[k] != nil {
				return nil, fmt.Errorf("attribute %s already exists: %s", k, op)
			}

			t[idx].Attributes[k] = bd.Attributes[k]
		}

		t[idx].Blocks = append(t[idx].Blocks, bd.Blocks...)
	}

	return resource, nil
}

func (op JSONPointerPathOperator) Replace(resource *configs.Resource, value Value) (*configs.Resource, error) {
	// Search.
	target, _, err := op.Search(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to search target: %w", err)
	}

	// Replace.
	seg := op[len(op)-1].Value

	switch t := target.(type) {
	default:
		return nil, fmt.Errorf("invalid target type: %s: %T", op, target)
	case *hclsyntax.Body:
		switch {
		case value.Attribute != nil:
			if t.Attributes == nil || t.Attributes[seg] == nil {
				return nil, fmt.Errorf("path not found: %s", op)
			}

			t.Attributes[seg].Expr = toHCLSyntaxExpression(value.Attribute.Expr)
		case value.Block != nil:
			blkIdxes := make([]int, 0, len(t.Blocks))

			for j := range t.Blocks {
				switch t.Blocks[j].Type {
				default:
					continue
				case dynamicBlockType:
					if t.Blocks[j].Labels[0] != seg {
						continue
					}
				case seg:
				}

				blkIdxes = append(blkIdxes, j)
			}

			if len(blkIdxes) == 0 {
				return nil, fmt.Errorf("path not found: %s", op)
			}

			for _, j := range blkIdxes {
				if t.Blocks[j].Type == dynamicBlockType {
					t.Blocks[j].Body.Blocks[0].Body = toHCLSyntaxBody(value.Block.Body)
					continue
				}

				t.Blocks[j].Body = toHCLSyntaxBody(value.Block.Body)
			}
		}
	case *hclsyntax.ObjectConsExpr:
		if value.Attribute == nil {
			return nil, errors.New("want patch attribute but got patch block")
		}

		idx := -1

		for i := range t.Items {
			tr, err := hcl.AbsTraversalForExpr(t.Items[i].KeyExpr)
			if err == nil && !tr.IsRelative() &&
				tr[0].(hcl.TraverseRoot).Name == seg {
				idx = i
			}
		}

		if idx == -1 {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		t.Items[idx].ValueExpr = toHCLSyntaxExpression(value.Attribute.Expr)
	case *hclsyntax.TupleConsExpr:
		if value.Attribute == nil {
			return nil, errors.New("want patch attribute but got patch block")
		}

		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t.Exprs) - 1)
		}

		if len(t.Exprs)-1 < int(idx) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		t.Exprs[idx] = toHCLSyntaxExpression(value.Attribute.Expr)
	case []*hclsyntax.Body:
		if value.Block == nil {
			return nil, errors.New("want patch block but got patch attribute")
		}

		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t) - 1)
		}

		if len(t)-1 < int(idx) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		bd := toHCLSyntaxBody(value.Block.Body)

		t[idx].Attributes = bd.Attributes
		t[idx].Blocks = bd.Blocks
	}

	return resource, nil
}

func (op JSONPointerPathOperator) Remove(resource *configs.Resource) (*configs.Resource, error) {
	// Search.
	target, parent, err := op.Search(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to search target: %w", err)
	}

	// Remove.
	seg := op[len(op)-1].Value

	switch t := target.(type) {
	default:
		return nil, fmt.Errorf("invalid target type: %s: %T", op, target)
	case *hclsyntax.Body:
		if t.Attributes[seg] != nil {
			delete(t.Attributes, seg)
			return resource, nil
		}

		blks := make([]*hclsyntax.Block, 0, len(t.Blocks))

		for idx := range t.Blocks {
			switch t.Blocks[idx].Type {
			case dynamicBlockType:
				if t.Blocks[idx].Labels[0] == seg {
					continue
				}
			case seg:
				continue
			}

			blks = append(blks, t.Blocks[idx])
		}

		t.Blocks = blks
	case *hclsyntax.ObjectConsExpr:
		var idx int
		for ; idx < len(t.Items); idx++ {
			tr, err := hcl.AbsTraversalForExpr(t.Items[idx].KeyExpr)
			if err == nil && !tr.IsRelative() &&
				tr[0].(hcl.TraverseRoot).Name == seg {
				break
			}
		}

		if idx == len(t.Items) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		t.Items = append(t.Items[:idx], t.Items[idx+1:]...)
	case *hclsyntax.TupleConsExpr:
		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t.Exprs) - 1)
		}

		if len(t.Exprs)-1 < int(idx) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		t.Exprs = append(t.Exprs[:idx], t.Exprs[idx+1:]...)
	case []*hclsyntax.Body:
		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t) - 1)
		}

		if len(t)-1 < int(idx) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		pSeg := op[len(op)-2].Value
		p := parent.(*hclsyntax.Body)
		blks := make(hclsyntax.Blocks, 0, len(p.Blocks))

		for j := range p.Blocks {
			switch p.Blocks[j].Type {
			case dynamicBlockType:
				if p.Blocks[j].Labels[0] != pSeg {
					blks = append(blks, p.Blocks[j])
					continue
				}

				if p.Blocks[j].Body.Blocks[0].Body != t[idx] {
					blks = append(blks, p.Blocks[j])
				}
			case pSeg:
				if p.Blocks[j].Body != t[idx] {
					blks = append(blks, p.Blocks[j])
				}
			}
		}

		p.Blocks = blks
	}

	return resource, nil
}

func (op JSONPointerPathOperator) Set(resource *configs.Resource, value Value) (*configs.Resource, error) {
	// Search.
	target, _, err := op.Search(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to search target: %w", err)
	}

	// Set.
	seg := op[len(op)-1].Value

	switch t := target.(type) {
	default:
		return nil, fmt.Errorf("invalid target type: %s: %T", op, target)
	case *hclsyntax.Body:
		switch {
		case value.Attribute != nil:
			if t.Attributes == nil {
				t.Attributes = make(map[string]*hclsyntax.Attribute)
			}

			if t.Attributes[seg] == nil {
				t.Attributes[seg] = &hclsyntax.Attribute{
					Name:      seg,
					Expr:      toHCLSyntaxExpression(value.Attribute.Expr),
					SrcRange:  value.Attribute.Range,
					NameRange: value.Attribute.NameRange,
				}

				break
			}

			t.Attributes[seg].Expr = toHCLSyntaxExpression(value.Attribute.Expr)
		case value.Block != nil:
			blkIdxes := make([]int, 0, len(t.Blocks))

			for j := range t.Blocks {
				switch t.Blocks[j].Type {
				default:
					continue
				case dynamicBlockType:
					if t.Blocks[j].Labels[0] != seg {
						continue
					}
				case seg:
				}

				blkIdxes = append(blkIdxes, j)
			}

			if len(blkIdxes) == 0 {
				t.Blocks = append(t.Blocks, &hclsyntax.Block{
					Type:        seg,
					Labels:      value.Block.Labels,
					Body:        toHCLSyntaxBody(value.Block.Body),
					TypeRange:   value.Block.TypeRange,
					LabelRanges: value.Block.LabelRanges,
				})

				break
			}

			for _, j := range blkIdxes {
				if t.Blocks[j].Type == dynamicBlockType {
					t.Blocks[j].Body.Blocks[0].Body = toHCLSyntaxBody(value.Block.Body)
					continue
				}

				t.Blocks[j].Body = toHCLSyntaxBody(value.Block.Body)
			}
		}
	case *hclsyntax.ObjectConsExpr:
		if value.Attribute == nil {
			return nil, errors.New("want patch attribute but got patch block")
		}

		idx := -1

		for i := range t.Items {
			tr, err := hcl.AbsTraversalForExpr(t.Items[i].KeyExpr)
			if err == nil && !tr.IsRelative() &&
				tr[0].(hcl.TraverseRoot).Name == seg {
				idx = i
			}
		}

		if idx == -1 {
			t.Items = append(t.Items, hclsyntax.ObjectConsItem{
				KeyExpr: &hclsyntax.ObjectConsKeyExpr{
					Wrapped: &hclsyntax.ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: seg,
							},
						},
					},
				},
				ValueExpr: toHCLSyntaxExpression(value.Attribute.Expr),
			})

			break
		}

		t.Items[idx].ValueExpr = toHCLSyntaxExpression(value.Attribute.Expr)
	case *hclsyntax.TupleConsExpr:
		if value.Attribute == nil {
			return nil, errors.New("want patch attribute but got patch block")
		}

		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t.Exprs) - 1)
		}

		if len(t.Exprs)-1 < int(idx) {
			t.Exprs = append(t.Exprs, toHCLSyntaxExpression(value.Attribute.Expr))

			break
		}

		t.Exprs[idx] = toHCLSyntaxExpression(value.Attribute.Expr)
	case []*hclsyntax.Body:
		if value.Block == nil {
			return nil, errors.New("want patch block but got patch attribute")
		}

		idx, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid indexer path: %s", op)
		}

		if idx == -1 {
			idx = int64(len(t) - 1)
		}

		if len(t)-1 < int(idx) {
			return nil, fmt.Errorf("path not found: %s", op)
		}

		bd := toHCLSyntaxBody(value.Block.Body)

		t[idx].Attributes = bd.Attributes
		t[idx].Blocks = bd.Blocks
	}

	return resource, nil
}

func TokenizeJSONPointerPath(path string) []JSONPointerPathToken {
	ps := strings.Split(path, "/")
	if len(ps) < 2 {
		return nil
	}

	ps = ps[1:]

	var (
		// http://tools.ietf.org/html/rfc6901#section-4
		dec = strings.NewReplacer("~1", "/", "~0", "~")
		tks = make([]JSONPointerPathToken, 0, len(ps))
	)

	for i := range ps {
		tks = append(tks, JSONPointerPathToken{
			Raw:   ps[i],
			Value: dec.Replace(ps[i]),
		})
	}

	return tks
}
