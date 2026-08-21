package core

import (
	"context"
	"fmt"
	"testing"
)

func TestEvalBatchOrderAndValues(t *testing.T) {
	inputs := make([]map[string]any, 20)
	for i := range inputs {
		inputs[i] = map[string]any{"x": i}
	}

	results, err := EvalBatch(context.Background(), "x * 2", inputs, BatchConfig{Concurrency: 4})
	if err != nil {
		t.Fatalf("EvalBatch: %v", err)
	}
	if len(results) != len(inputs) {
		t.Fatalf("got %d results, want %d", len(results), len(inputs))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("input %d: unexpected error: %v", i, r.Err)
			continue
		}
		got, ok := r.Value.(int64)
		if !ok {
			t.Errorf("input %d: value type %T, want int64", i, r.Value)
			continue
		}
		if want := int64(i * 2); got != want {
			t.Errorf("input %d: got %d, want %d", i, got, want)
		}
	}
}

func TestEvalBatchErrorIsolation(t *testing.T) {
	script := `
		assert(x != 0, "x is zero")
		100 / x
	`
	inputs := []map[string]any{{"x": 5}, {"x": 0}, {"x": 4}}

	results, err := EvalBatch(context.Background(), script, inputs, BatchConfig{})
	if err != nil {
		t.Fatalf("EvalBatch: %v", err)
	}
	if results[0].Err != nil || results[0].Value.(int64) != 20 {
		t.Errorf("input 0: got value=%v err=%v, want 20/nil", results[0].Value, results[0].Err)
	}
	if results[1].Err == nil {
		t.Errorf("input 1: expected an error for x==0, got value=%v", results[1].Value)
	}
	if results[2].Err != nil || results[2].Value.(int64) != 25 {
		t.Errorf("input 2: got value=%v err=%v, want 25/nil", results[2].Value, results[2].Err)
	}
}

// TestEvalBatchConcurrentModules exercises real per-worker module work under high
// concurrency. Run with -race to catch shared-state bugs.
func TestEvalBatchConcurrentModules(t *testing.T) {
	script := `len(csv.parse(data))` // header row is dropped, so this is row count

	inputs := make([]map[string]any, 50)
	for i := range inputs {
		csvText := "name,age\n"
		for r := 0; r <= i; r++ { // i+1 data rows
			csvText += fmt.Sprintf("row%d,%d\n", r, r)
		}
		inputs[i] = map[string]any{"data": csvText}
	}

	results, err := EvalBatch(context.Background(), script, inputs, BatchConfig{Concurrency: 8}, WithCSV())
	if err != nil {
		t.Fatalf("EvalBatch: %v", err)
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("input %d: unexpected error: %v", i, r.Err)
			continue
		}
		if want := int64(i + 1); r.Value.(int64) != want {
			t.Errorf("input %d: got %v rows, want %d", i, r.Value, want)
		}
	}
}

// TestEvalBatchConcurrencyParity confirms parallelism does not change outcomes.
func TestEvalBatchConcurrencyParity(t *testing.T) {
	inputs := make([]map[string]any, 30)
	for i := range inputs {
		inputs[i] = map[string]any{"x": i}
	}
	script := "x * x"

	serial, err := EvalBatch(context.Background(), script, inputs, BatchConfig{Concurrency: 1})
	if err != nil {
		t.Fatalf("serial EvalBatch: %v", err)
	}
	parallel, err := EvalBatch(context.Background(), script, inputs, BatchConfig{})
	if err != nil {
		t.Fatalf("parallel EvalBatch: %v", err)
	}
	for i := range inputs {
		if serial[i].Value != parallel[i].Value || serial[i].Err != nil || parallel[i].Err != nil {
			t.Errorf("input %d: serial=%v parallel=%v", i, serial[i].Value, parallel[i].Value)
		}
	}
}

func TestEvalBatchEmpty(t *testing.T) {
	results, err := EvalBatch(context.Background(), "x", nil, BatchConfig{})
	if err != nil {
		t.Fatalf("EvalBatch: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}
