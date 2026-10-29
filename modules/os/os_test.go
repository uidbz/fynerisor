package os

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")

	result, err := create(ctx, object.NewString(path))
	if err != nil {
		t.Fatalf("create returned Go error: %v", err)
	}
	if result != object.Nil {
		t.Errorf("expected Nil, got %s", result.Type())
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file was not created: %v", err)
	}

	// Fails on an invalid path: the error is *raised* (issue #30), so
	// scripts can catch it with try/catch instead of inspecting values.
	result, err = create(ctx, object.NewString(filepath.Join(dir, "missing", "x.txt")))
	if err == nil {
		t.Errorf("create on a missing dir should raise, got result %s", result.Type())
	}
}

func TestCopyFile(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	if err := os.WriteFile(src, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := copyFile(ctx, object.NewString(src), object.NewString(dst))
	if err != nil {
		t.Fatalf("copy_file returned Go error: %v", err)
	}
	if result != object.Nil {
		t.Errorf("expected Nil, got %s", result.Type())
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("dst was not created: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}

	// Mode must be preserved from the source file
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %v", info.Mode().Perm())
	}

	// Missing source raises (issue #30).
	result, err = copyFile(ctx, object.NewString(filepath.Join(dir, "nope")), object.NewString(dst))
	if err == nil {
		t.Errorf("copy_file of a missing source should raise, got result %s", result.Type())
	}
}

func TestMkdir(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "sub")

	result, err := mkdir(ctx, object.NewString(path), object.NewInt(0700))
	if err != nil {
		t.Fatalf("mkdir returned Go error: %v", err)
	}
	if result != object.Nil {
		t.Errorf("expected Nil, got %s", result.Type())
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("expected mode 0700, got %v", info.Mode().Perm())
	}

	// Nested path without -all raises (issue #30).
	result, err = mkdir(ctx, object.NewString(filepath.Join(dir, "missing", "deep")))
	if err == nil {
		t.Errorf("nested mkdir without -all should raise, got result %s", result.Type())
	}
}

func TestRemoveAll(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	sub := filepath.Join(dir, "tree")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := removeAll(ctx, object.NewString(sub))
	if err != nil {
		t.Fatalf("remove_all returned Go error: %v", err)
	}
	if result != object.Nil {
		t.Errorf("expected Nil, got %s", result.Type())
	}

	if _, err := os.Stat(sub); !os.IsNotExist(err) {
		t.Errorf("directory still exists: %v", err)
	}

	// Removing a non-existent path succeeds like os.RemoveAll
	result, err = removeAll(ctx, object.NewString(filepath.Join(dir, "gone")))
	if err != nil {
		t.Fatalf("remove_all returned Go error: %v", err)
	}
	if result != object.Nil {
		t.Errorf("expected Nil, got %s", result.Type())
	}
}

func TestArgs(t *testing.T) {
	ctx := context.Background()

	SetScriptArgs([]string{"input.csv", "--force"})
	defer SetScriptArgs(nil)

	result, err := scriptArgsFunc(ctx)
	if err != nil {
		t.Fatalf("args returned Go error: %v", err)
	}

	list, ok := result.(*object.List)
	if !ok {
		t.Fatalf("expected List, got %s", result.Type())
	}
	if list.Size() != 2 {
		t.Fatalf("expected 2 args, got %d", list.Size())
	}
	if list.Value()[0].(*object.String).Value() != "input.csv" {
		t.Errorf("unexpected first arg: %v", list.Value()[0])
	}
}

func TestHomeDir(t *testing.T) {
	ctx := context.Background()

	result, err := homeDir(ctx)
	if err != nil {
		t.Fatalf("home_dir returned Go error: %v", err)
	}

	s, ok := result.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %s", result.Type())
	}

	u, err := user.Current()
	if err != nil {
		t.Skipf("cannot resolve current user: %v", err)
	}
	if s.Value() != u.HomeDir {
		t.Errorf("expected %q, got %q", u.HomeDir, s.Value())
	}
}

func TestReadDirPath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := readDir(ctx, object.NewString(dir))
	if err != nil {
		t.Fatalf("read_dir returned Go error: %v", err)
	}

	list, ok := result.(*object.List)
	if !ok {
		t.Fatalf("expected List, got %s", result.Type())
	}
	if list.Size() != 1 {
		t.Fatalf("expected 1 entry, got %d", list.Size())
	}

	entry, ok := list.Value()[0].(*object.Map)
	if !ok {
		t.Fatalf("expected Map entry, got %s", list.Value()[0].Type())
	}
	pathObj := entry.Get("path")
	if pathObj == nil {
		t.Fatal("entry missing 'path' field")
	}
	want := filepath.Join(dir, "a.txt")
	if got, _ := object.AsString(pathObj); got != want {
		t.Errorf("expected path %q, got %q", want, got)
	}
}

// TestScriptArgsContextOverridesGlobal covers issue #44: args carried
// by the eval context (per-runtime configuration) win over the
// deprecated process global, and an explicitly empty list is distinct
// from "not configured".
func TestScriptArgsContextOverridesGlobal(t *testing.T) {
	SetScriptArgs([]string{"global"})
	defer SetScriptArgs(nil)

	result, err := scriptArgsFunc(ContextWithArgs(context.Background(), []string{"mine"}))
	if err != nil {
		t.Fatalf("args returned Go error: %v", err)
	}
	if got := fmt.Sprintf("%v", result.Inspect()); got != `["mine"]` {
		t.Errorf("context args = %s, want [\"mine\"] (global leaked)", got)
	}

	result, err = scriptArgsFunc(ContextWithArgs(context.Background(), []string{}))
	if err != nil {
		t.Fatalf("args returned Go error: %v", err)
	}
	if got := fmt.Sprintf("%v", result.Inspect()); got != `[]` {
		t.Errorf("explicit empty args = %s, want [] (global leaked)", got)
	}
}
