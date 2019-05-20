// Package resourcedatasetkey defines an Analyzer that checks for
// ResourceData.Set() calls using * dereferences
package resourcedatasetkey

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/terraform-providers/terraform-provider-aws/linter/passes/commentignore"
	"github.com/terraform-providers/terraform-provider-aws/linter/passes/resourcedataset"
)

const Doc = `check for Resource having Exists functions

The resourcedatasetkey analyzer reports a complex key argument for a Set()
call. It is preferred to explicitly use a string literal as the key argument.`

const analyzerName = "resourcedatasetkey"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		resourcedataset.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	sets := pass.ResultOf[resourcedataset.Analyzer].([]*ast.CallExpr)
	for _, set := range sets {
		if ignorer.ShouldIgnore(analyzerName, set) {
			continue
		}

		if len(set.Args) < 2 {
			continue
		}

		switch v := set.Args[0].(type) {
		default:
			pass.Reportf(v.Pos(), "%s: ResourceData.Set() key argument should be string literal", analyzerName)
		case *ast.BasicLit:
			continue
		}
	}

	return nil, nil
}