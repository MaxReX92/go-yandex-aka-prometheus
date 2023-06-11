package analysis

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_Analyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./...")
}
