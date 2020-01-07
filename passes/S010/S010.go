// Package S010 defines an Analyzer that checks for
// Schema with only Computed enabled and ValidateFunc configured
package S010

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/schemaschema"
)

const Doc = `check for Schema with only Computed enabled and ValidateFunc configured

The S010 analyzer reports cases of schemas which only enables Computed
and configures ValidateFunc, which will fail provider schema validation.`

const analyzerName = "S010"

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

		computed := terraformtype.HelperSchemaTypeSchemaComputed(schema)

		if computed == nil || !*computed {
			continue
		}

		optional := terraformtype.HelperSchemaTypeSchemaOptional(schema)

		if optional != nil && *optional {
			continue
		}

		required := terraformtype.HelperSchemaTypeSchemaRequired(schema)

		if required != nil && *required {
			continue
		}

		validateFunc := terraformtype.HelperSchemaTypeSchemaValidateFunc(schema)

		if validateFunc == nil {
			continue
		}

		switch t := schema.Type.(type) {
		default:
			pass.Reportf(schema.Lbrace, "%s: schema should not only enable Computed and configure ValidateFunc", analyzerName)
		case *ast.SelectorExpr:
			pass.Reportf(t.Sel.Pos(), "%s: schema should not only enable Computed and configure ValidateFunc", analyzerName)
		}
	}

	return nil, nil
}
