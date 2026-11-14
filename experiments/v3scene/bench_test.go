package v3scene

import (
	"context"
	"os"
	"testing"
)

func BenchmarkCounterStep(b *testing.B) {
	app := newTestApp(b, ModeActions,
		"testdata/counter.scene.json",
		"testdata/counter/update.risor")
	stepTB(b, app, map[string]any{"type": "init"})
	event := map[string]any{"type": "tapped", "node": "inc"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := app.Step(context.Background(), event); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataToolHostFunctionStep(b *testing.B) {
	app := newTestApp(b, ModeHostFunctions,
		"testdata/customers/customers.scene.json",
		"testdata/customers/host-functions.risor")
	stepTB(b, app, map[string]any{"type": "init"})
	event := map[string]any{"type": "tapped", "node": "refresh"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := app.Step(context.Background(), event); err != nil {
			b.Fatal(err)
		}
	}
}

func stepTB(tb testing.TB, app *App, event map[string]any) {
	tb.Helper()
	if err := app.Step(context.Background(), event); err != nil {
		tb.Fatal(err)
	}
}

func BenchmarkCompileCounter(b *testing.B) {
	doc, err := LoadDocument("testdata/counter.scene.json")
	if err != nil {
		b.Fatal(err)
	}
	src, err := os.ReadFile("testdata/counter/update.risor")
	if err != nil {
		b.Fatal(err)
	}
	cfg := Config{Document: doc, Script: string(src), Mode: ModeActions}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Check(cfg); err != nil {
			b.Fatal(err)
		}
	}
}
