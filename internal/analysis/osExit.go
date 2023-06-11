package analysis

import "golang.org/x/tools/go/analysis"

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "check for os.Exit calls from main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}
