// staticlint tool contains a list of go/analysis checks.
//
// The tool includes the following checks:
//
// 1. All go/analysis passes checks (https://pkg.go.dev/golang.org/x/tools/go/analysis/passes);
//
// 2. All SA staticcheck checks (https://staticcheck.io/docs/checks/#SA);
//
// 3. All QF quickfix checks (https://staticcheck.io/docs/checks/#QF);
//
// 4. Analyzer to check for unchecked errors in code (https://github.com/kisielk/errcheck);
//
// 5. Analyzer to detect magic numbers (https://github.com/tommy-muehle/go-mnd);
//
// 6. Analyzer to check for calling os.Exit in main functions.
//
// For more details run:
//
//	staticlint -help
//
// The following example perform no os.Exit calls analysis for given project:
//
//	staticlint -noosexit <project path>
package main
