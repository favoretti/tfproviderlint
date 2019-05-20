// Package schematypemapelem defines an Analyzer that checks for
// Schema of TypeMap missing Elem
package schematypemapelem

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/terraform-providers/terraform-provider-aws/linter/passes/commentignore"
	"github.com/terraform-providers/terraform-provider-aws/linter/passes/schemaschema"
)

const Doc = `check for Schema of TypeMap missing Elem

The schematypemapelem analyzer reports cases of TypeMap schemas missing Elem,
which currently passes Terraform schema validation, but breaks downstream tools
and may be required in the future.`

const analyzerName = "schematypemapelem"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		schemaschema.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemas := pass.ResultOf[schemaschema.Analyzer].([]*ast.CompositeLit)
	for _, schema := range schemas {
		if ignorer.ShouldIgnore(analyzerName, schema) {
			continue
		}

		var elemFound bool
		var typeMap bool

		for _, elt := range schema.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				name := v.Key.(*ast.Ident).Name

				if name == "Elem" {
					elemFound = true
					continue
				}

				if name != "Type" {
					continue
				}

				switch v := v.Value.(type) {
				default:
					continue
				case *ast.SelectorExpr:
					// Use AST over TypesInfo here as schema uses ValueType
					if v.Sel.Name != "TypeMap" {
						continue
					}

					switch t := pass.TypesInfo.TypeOf(v).(type) {
					default:
						continue
					case *types.Named:
						// HasSuffix here due to vendoring
						if !strings.HasSuffix(t.Obj().Pkg().Path(), "github.com/hashicorp/terraform/helper/schema") {
							continue
						}

						typeMap = true
					}
				}
			}
		}

		if typeMap && !elemFound {
			pass.Reportf(schema.Type.(*ast.SelectorExpr).Sel.Pos(), "schema of TypeMap should include Elem")
		}
	}

	return nil, nil
}