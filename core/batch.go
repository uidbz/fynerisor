package core

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"
)

// BatchResult is the outcome of running the batch script for a single input.
// Results returned by EvalBatch are index-aligned with the inputs slice.
type BatchResult struct {
	Value any   // Script result, converted to a Go value. Nil when Err is set.
	Err   error // Per-input error (compilation is shared, so this is a run/panic error). Nil on success.
}

// BatchConfig tunes how EvalBatch executes.
type BatchConfig struct {
	// Concurrency bounds how many inputs run at once. Values <= 0 default to
	// runtime.GOMAXPROCS(0).
	Concurrency int
}

// EvalBatch compiles script once and runs it over each input concurrently.
//
// Each input runs on its own fresh Context and VM, so there is no shared mutable
// interpreter state between workers: this is the "compile once, run many" pattern
// (an immutable *bytecode.Code shared across independent VMs). Each entry in inputs
// is exposed to the script as global variables, e.g. an input of {"file": "a.csv"}
// makes the global `file` available to that run.
//
// Results are index-aligned with inputs. The returned error is non-nil only for
// setup failures such as compilation; per-input failures are reported in the
// corresponding BatchResult.Err so one bad input does not fail the whole batch.
//
// EvalBatch is headless: it has no GUI dependency and does not touch the Fyne UI
// thread. Do not use it to drive widgets.
//
// Concurrency-safety notes:
//   - Stateless module globals (csv, json, strings, ...) are safe: each worker
//     Context constructs its own module objects from opts.
//   - Values passed via WithGlobal(name, value) are shared by pointer across all
//     workers. If such a value is stateful (a DB handle, a mutable map), the
//     caller is responsible for making it concurrency-safe.
//   - The script itself should not rely on cross-input shared state.
func EvalBatch(
	ctx context.Context,
	script string,
	inputs []map[string]any,
	cfg BatchConfig,
	opts ...Option,
) ([]BatchResult, error) {
	results := make([]BatchResult, len(inputs))
	if len(inputs) == 0 {
		return results, nil
	}

	// Union of all input keys, so every worker env carries the same key set and
	// satisfies the compiled bytecode's global-name validation regardless of any
	// per-input key differences.
	inputKeys := make(map[string]struct{})
	for _, in := range inputs {
		for k := range in {
			inputKeys[k] = struct{}{}
		}
	}

	// Compile once. A probe Context supplies the base global names (csv, print,
	// ...); input keys are added as placeholders. Only the names matter at
	// compile time, not the values.
	probe := NewContext(opts...)
	templateEnv := make(map[string]any, len(probe.env)+len(inputKeys))
	for k, v := range probe.env {
		templateEnv[k] = v
	}
	for k := range inputKeys {
		if _, ok := templateEnv[k]; !ok {
			templateEnv[k] = nil
		}
	}

	code, err := risor.Compile(ctx, script, risor.WithEnv(templateEnv))
	if err != nil {
		return nil, fmt.Errorf("batch: compilation failed: %w", err)
	}

	// runOne executes the shared compiled code for a single input on its own
	// fresh, isolated Context+VM and records the outcome at results[idx].
	runOne := func(idx int, input map[string]any) {
		defer func() {
			if r := recover(); r != nil {
				results[idx] = BatchResult{Err: fmt.Errorf("batch: panic running input %d: %v", idx, r)}
			}
		}()

		// Fresh, isolated Context per worker: its own env, module cache and
		// import stack, so concurrent import() calls cannot race.
		worker := NewContext(opts...)

		// Worker env = base globals + full input-key union (backfilled nil) +
		// this input's values. Keeps every run env a superset of the compiled
		// globals.
		env := make(map[string]any, len(worker.env)+len(inputKeys))
		for k, v := range worker.env {
			env[k] = v
		}
		for k := range inputKeys {
			env[k] = nil
		}
		for k, v := range input {
			env[k] = v
		}

		runOpts := append([]risor.Option{risor.WithEnv(env)}, worker.userGlobals...)
		value, runErr := risor.Run(ctx, code, runOpts...)
		results[idx] = BatchResult{Value: value, Err: runErr}
	}

	// Warm up on the calling goroutine before fanning out. The first run
	// initializes risor's lazily-created package globals (e.g. the default type
	// registry); doing it once up front means the concurrent workers below only
	// read that shared state, never race to create it.
	runOne(0, inputs[0])
	if len(inputs) == 1 {
		return results, nil
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.GOMAXPROCS(0)
	}
	if concurrency > len(inputs)-1 {
		concurrency = len(inputs) - 1
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := 1; i < len(inputs); i++ {
		// Stop dispatching new work if the context is cancelled.
		if err := ctx.Err(); err != nil {
			results[i] = BatchResult{Err: err}
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, input map[string]any) {
			defer wg.Done()
			defer func() { <-sem }()
			runOne(idx, input)
		}(i, inputs[i])
	}

	wg.Wait()
	return results, nil
}
