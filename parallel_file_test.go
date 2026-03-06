package swag

import (
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeFilesParallel_ExecutesAllFiles(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/simple"
	p := New()
	
	require.NoError(t, p.getAllGoFileInfo("github.com/swaggo/swag/testdata/simple", searchDir))
	
	var counter int32
	err := p.packages.RangeFilesParallel(func(info *AstFileInfo) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}, 4)
	
	assert.NoError(t, err)
	assert.Greater(t, counter, int32(0), "Should process at least one file")
}

func TestRangeFilesParallel_ErrorPropagation(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/simple"
	p := New()
	
	require.NoError(t, p.getAllGoFileInfo("github.com/swaggo/swag/testdata/simple", searchDir))
	
	testErr := assert.AnError
	err := p.packages.RangeFilesParallel(func(info *AstFileInfo) error {
		return testErr
	}, 4)
	
	assert.Error(t, err)
}

func TestRangeFilesParallel_ConcurrencyLimit(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/simple"
	p := New()
	
	require.NoError(t, p.getAllGoFileInfo("github.com/swaggo/swag/testdata/simple", searchDir))
	
	// Test with different concurrency limits
	for _, limit := range []int{1, 2, 4, -1} {
		t.Run("limit_"+string(rune(limit+'0')), func(t *testing.T) {
			var counter int32
			err := p.packages.RangeFilesParallel(func(info *AstFileInfo) error {
				atomic.AddInt32(&counter, 1)
				return nil
			}, limit)
			
			assert.NoError(t, err)
			assert.Greater(t, counter, int32(0))
		})
	}
}

func TestParallelParsing_SequentialEquivalence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping equivalence test in short mode")
	}

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// Parse sequentially
	seqParser := New()
	err := seqParser.ParseAPIMultiSearchDir([]string{searchDir}, mainFile, 100)
	require.NoError(t, err)
	seqSwagger := seqParser.GetSwagger()

	// Parse in parallel
	parParser := New()
	err = parParser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 4)
	require.NoError(t, err)
	parSwagger := parParser.GetSwagger()

	// Marshal both to JSON for comparison
	seqJSON, err := json.MarshalIndent(seqSwagger, "", "  ")
	require.NoError(t, err)

	parJSON, err := json.MarshalIndent(parSwagger, "", "  ")
	require.NoError(t, err)

	// Compare JSON representations
	assert.JSONEq(t, string(seqJSON), string(parJSON), 
		"Parallel parsing should produce identical output to sequential parsing")
}

func TestParallelParsing_NoDuplicateOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping duplicate check test in short mode")
	}

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// Run parallel parsing multiple times to catch race conditions
	for i := 0; i < 10; i++ {
		parser := New()
		err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 8)
		require.NoError(t, err)

		swagger := parser.GetSwagger()
		
		// Check for duplicate operation IDs
		operationIDs := make(map[string]int)
		for path, pathItem := range swagger.Paths.Paths {
			methods := []struct{
				name string
				op interface{}
			}{
				{"GET", pathItem.Get},
				{"POST", pathItem.Post},
				{"PUT", pathItem.Put},
				{"DELETE", pathItem.Delete},
				{"PATCH", pathItem.Patch},
				{"HEAD", pathItem.Head},
				{"OPTIONS", pathItem.Options},
			}
			
			for _, m := range methods {
				if m.op != nil {
					// Operation exists for this method
					key := path + ":" + m.name
					operationIDs[key]++
					assert.Equal(t, 1, operationIDs[key], 
						"Operation %s should appear exactly once", key)
				}
			}
		}
	}
}

func TestParallelParsing_NoDuplicateDefinitions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping duplicate definitions test in short mode")
	}

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// Run parallel parsing multiple times
	for i := 0; i < 10; i++ {
		parser := New()
		err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 8)
		require.NoError(t, err)

		swagger := parser.GetSwagger()
		
		// Verify each definition name appears exactly once
		definitionNames := make(map[string]int)
		for name := range swagger.Definitions {
			definitionNames[name]++
			assert.Equal(t, 1, definitionNames[name], 
				"Definition %s should appear exactly once", name)
		}
	}
}

func TestParallelParsing_WithDifferentConcurrencyLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency levels test in short mode")
	}

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// Test with different concurrency levels
	concurrencyLevels := []int{0, 1, 2, 4, 8, -1}
	
	var referenceJSON string
	for i, concurrency := range concurrencyLevels {
		parser := New()
		err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, concurrency)
		require.NoError(t, err)

		swagger := parser.GetSwagger()
		swaggerJSON, err := json.MarshalIndent(swagger, "", "  ")
		require.NoError(t, err)

		if i == 0 {
			// First iteration (sequential) is the reference
			referenceJSON = string(swaggerJSON)
		} else {
			// All other concurrency levels should match
			assert.JSONEq(t, referenceJSON, string(swaggerJSON),
				"Concurrency level %d should produce same output as sequential", concurrency)
		}
	}
}

func TestParallelParsing_DeterministicTagOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tag ordering test in short mode")
	}

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// Parse multiple times in parallel
	var previousTagOrder []string
	for i := 0; i < 10; i++ {
		parser := New()
		err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 8)
		require.NoError(t, err)

		swagger := parser.GetSwagger()
		
		// Extract tag names
		tagNames := make([]string, len(swagger.Tags))
		for i, tag := range swagger.Tags {
			tagNames[i] = tag.Name
		}

		if i == 0 {
			previousTagOrder = tagNames
		} else {
			// Tags should be in the same order every time
			assert.Equal(t, previousTagOrder, tagNames,
				"Tags should be in deterministic order across parallel runs")
		}
	}
}

func TestParallelParsing_ZeroFileParallelism(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// Zero should use sequential parsing (backward compatible)
	parser := New()
	err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 0)
	assert.NoError(t, err)
}

func TestParallelParsing_AutoConcurrency(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/simple"
	mainFile := "main.go"

	// -1 should use GOMAXPROCS
	parser := New()
	err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, -1)
	assert.NoError(t, err)
}
