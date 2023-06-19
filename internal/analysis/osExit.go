package analysis

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "check for os.Exit calls from main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name != "main" {
					// not in main package
					return false
				}
			case *ast.FuncDecl:
				if x.Name.Name != "main" {
					// not in main func
					return false
				}
			case *ast.SelectorExpr:
				if x.X != nil && x.Sel != nil {
					ident, ok := x.X.(*ast.Ident)
					if ok && ident.Name == "os" && x.Sel.Name == "Exit" {
						pass.Reportf(ident.NamePos, "os.Exit called from main func")
					}
				}
			}

			return true
		})
	}
	return nil, nil
}
