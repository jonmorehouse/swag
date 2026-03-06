package swag

import (
	"runtime"
	"testing"

	"github.com/go-openapi/spec"
)

// BenchmarkParsing_Sequential benchmarks sequential file parsing
func BenchmarkParsing_Sequential(b *testing.B) {
	searchDir := "testdata/simple"
	mainFile := "main.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := New()
		if err := parser.ParseAPIMultiSearchDir([]string{searchDir}, mainFile, 100); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsing_Parallel_Auto benchmarks parallel parsing with auto concurrency
func BenchmarkParsing_Parallel_Auto(b *testing.B) {
	searchDir := "testdata/simple"
	mainFile := "main.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := New()
		if err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, -1); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsing_Parallel_2Workers benchmarks parallel parsing with 2 workers
func BenchmarkParsing_Parallel_2Workers(b *testing.B) {
	searchDir := "testdata/simple"
	mainFile := "main.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := New()
		if err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 2); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsing_Parallel_4Workers benchmarks parallel parsing with 4 workers
func BenchmarkParsing_Parallel_4Workers(b *testing.B) {
	searchDir := "testdata/simple"
	mainFile := "main.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := New()
		if err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 4); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsing_Parallel_8Workers benchmarks parallel parsing with 8 workers
func BenchmarkParsing_Parallel_8Workers(b *testing.B) {
	searchDir := "testdata/simple"
	mainFile := "main.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := New()
		if err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 8); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsing_Parallel_16Workers benchmarks parallel parsing with 16 workers
func BenchmarkParsing_Parallel_16Workers(b *testing.B) {
	searchDir := "testdata/simple"
	mainFile := "main.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := New()
		if err := parser.ParseAPIMultiSearchDirWithConcurrency([]string{searchDir}, mainFile, 100, 16); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRangeFiles_Sequential benchmarks sequential file iteration
func BenchmarkRangeFiles_Sequential(b *testing.B) {
	searchDir := "testdata/simple"
	parser := New()
	
	if err := parser.getAllGoFileInfo("github.com/swaggo/swag/testdata/simple", searchDir); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := parser.packages.RangeFiles(func(info *AstFileInfo) error {
			return nil
		}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRangeFiles_Parallel benchmarks parallel file iteration
func BenchmarkRangeFiles_Parallel(b *testing.B) {
	searchDir := "testdata/simple"
	parser := New()
	
	if err := parser.getAllGoFileInfo("github.com/swaggo/swag/testdata/simple", searchDir); err != nil {
		b.Fatal(err)
	}

	concurrency := runtime.GOMAXPROCS(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := parser.packages.RangeFilesParallel(func(info *AstFileInfo) error {
			return nil
		}, concurrency); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMutexContention benchmarks lock contention in processRouterOperation
func BenchmarkMutexContention(b *testing.B) {
	parser := New()
	parser.swagger.Paths = &spec.Paths{
		Paths: make(map[string]spec.PathItem),
	}

	operation := &Operation{
		Operation: spec.Operation{
			OperationProps: spec.OperationProps{
				ID: "testOp",
			},
		},
		RouterProperties: []RouteProperties{
			{
				Path:       "/test",
				HTTPMethod: "GET",
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := processRouterOperation(parser, operation); err != nil {
				b.Fatal(err)
			}
		}
	})
}