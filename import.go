package fynerisor

import (
	"context"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// newImportBuiltin creates the import() function that is a no-op at runtime.
// The import() function is only used as a marker for static analysis via
// AnalyzeRequirements(). The actual importing is done by the Go application
// before script execution via Window.ImportScript().
//
// This design allows:
// - Scripts to declare dependencies explicitly
// - Static analysis to extract imports before execution
// - Go applications to control import resolution and security
// - Import order to be preserved
//
// Example script:
//
//	import(["utils.risor", "helpers.risor"])
//	require(["v0.2", "@http"])
//
//	// Use imported functions
//	let result = utilFunction()
//
// Example Go application:
//
//	reqs, _ := fynerisor.AnalyzeRequirements(script)
//	fw := fynerisor.NewApp("My App")
//	for _, imp := range reqs.Imports {
//	    fw.ImportScript(imp)  // Load each import
//	}
//	fw.LoadScript(script)
//	fw.Execute()
func newImportBuiltin() *object.Builtin {
	return object.NewBuiltin("import", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		// import() is a no-op at runtime
		// It's only used for static analysis via AnalyzeRequirements()
		// The actual importing is done by the Go application before execution
		return object.Nil, nil
	})
}
