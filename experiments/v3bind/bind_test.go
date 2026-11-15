package v3bind

import (
	"context"
	"strings"
	"sync"
	"testing"

	risor "github.com/deepnoodle-ai/risor/go/v3"
)

type State struct {
	Count int      `json:"count"`
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

func TestCompileAndRun(t *testing.T) {
	set, err := Compile[State](map[string]string{
		"count":    "state.count",
		"title":    `"Count: " + string(state.count)`,
		"hasItems": "len(state.items) > 0",
	}, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Names(); len(got) != 3 || got[0] != "count" || got[1] != "hasItems" || got[2] != "title" {
		t.Fatalf("Names = %v", got)
	}

	out, err := set.Run(context.Background(), State{
		Count: 2,
		Name:  "Ada",
		Items: []string{"a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n, ok := risor.AsInt(out["count"]); !ok || n != 2 {
		t.Fatalf("count = %#v", out["count"])
	}
	if out["title"] != "Count: 2" {
		t.Fatalf("title = %#v", out["title"])
	}
	if out["hasItems"] != true {
		t.Fatalf("hasItems = %#v", out["hasItems"])
	}
}

func TestCheckErrorNamesTheField(t *testing.T) {
	_, err := Compile[State](map[string]string{
		"bad": "state.nope",
	}, Config{})
	if err == nil || !strings.Contains(err.Error(), "no field") {
		t.Fatalf("got %v, want a no-field check error", err)
	}
}

func TestLimits(t *testing.T) {
	set, err := Compile[State](map[string]string{
		"big": "range(10000).sum()",
	}, Config{Steps: 10})
	if err != nil {
		t.Fatal(err)
	}
	_, err = set.Run(context.Background(), State{})
	if err == nil || !strings.Contains(err.Error(), "steps-exceeded") {
		t.Fatalf("got %v, want steps-exceeded", err)
	}
}

func TestProgramsRunConcurrently(t *testing.T) {
	set, err := Compile[State](map[string]string{
		"count": "state.count",
	}, Config{})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			out, err := set.Run(context.Background(), State{Count: n})
			if err != nil {
				t.Errorf("run %d: %v", n, err)
				return
			}
			if got, _ := risor.AsInt(out["count"]); got != int64(n) {
				t.Errorf("run %d: count = %v", n, out["count"])
			}
		}(i)
	}
	wg.Wait()
}
