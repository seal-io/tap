package tap

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/configs"
)

type (
	TerraformResources = map[string]*configs.Resource
)

// Operate operates the patch on the given Terraform resources.
func Operate(tfRess TerraformResources, patch *Patch, pathSyntax string) (TerraformResources, error) {
	if patch == nil {
		return tfRess, nil
	}

	for _, op := range patch.Operations {
		po, err := getPathOperator(op.Path, pathSyntax)
		if err != nil {
			return nil, fmt.Errorf("error getting path operator: %w", err)
		}

		for rn, r := range tfRess {
			var nr *configs.Resource

			switch op.Mode {
			default:
				return nil, fmt.Errorf("unknown operation mode: %s", op.Mode)
			case "add":
				nr, err = po.Add(r, op.Value)
			case "replace":
				nr, err = po.Replace(r, op.Value)
			case "remove":
				nr, err = po.Remove(r)
			case "set":
				nr, err = po.Set(r, op.Value)
			}

			if err != nil && !patch.ContinueOnError {
				return nil, fmt.Errorf("error %s on resource %s: %w", op.Mode, rn, err)
			}

			if nr != nil {
				tfRess[rn] = nr
			}
		}
	}

	return tfRess, nil
}

type PathOperator interface {
	// Add adds Value at the path if not found in the given configs.Resource.
	Add(*configs.Resource, Value) (*configs.Resource, error)
	// Replace replaces the value at the path if found in the given configs.Resource.
	Replace(*configs.Resource, Value) (*configs.Resource, error)
	// Remove removes the value at the path if found in the given configs.Resource.
	Remove(*configs.Resource) (*configs.Resource, error)
	// Set sets the value at the path of the given configs.Resource.
	Set(*configs.Resource, Value) (*configs.Resource, error)
}

func getPathOperator(path, pathSyntax string) (PathOperator, error) {
	switch pathSyntax {
	default:
		return nil, fmt.Errorf("unknown path syntax: %s", pathSyntax)
	case "json_pointer":
		return NewJSONPointerPathOperator(path)
	}
}

func toHCLSyntaxExpression(expression hcl.Expression) hclsyntax.Expression {
	return expression.(hclsyntax.Expression)
}

func toHCLSyntaxBody(body hcl.Body) *hclsyntax.Body {
	return body.(*hclsyntax.Body)
}
