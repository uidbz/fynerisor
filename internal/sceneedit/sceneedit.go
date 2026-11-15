// Package sceneedit rewrites scene sidecar scripts for the editor workflow.
//
// The convention is deliberately small and visible:
//
//   - one standalone handler per node event, named on_<node>_<event>;
//   - handlers are ordinary let-bound functions, never inline in page.Bind;
//   - one page.Bind({...}) section at the bottom of the file.
//
// The package works on source text and the scene document only. It does not
// execute a script, touch Fyne, or depend on the designer.
package sceneedit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/deepnoodle-ai/risor/v2/pkg/ast"
	"github.com/deepnoodle-ai/risor/v2/pkg/parser"

	"github.com/uidbz/fynerisor/scene"
)

// Report says what a rewrite changed.
type Report struct {
	// Created lists handlers added as stubs.
	Created []string
	// Converted lists inline page.Bind handlers moved to standalone lets.
	Converted []string
	// Bound lists names added to the page.Bind section.
	Bound []string
	// Warnings lists constructs the rewrite left alone.
	Warnings []string
}

// Result is a rewritten script and document.
type Result struct {
	Source   string
	Document *scene.Document
	Report   Report
}

// Apply runs the whole editor pass: it gives every call action a handler,
// creates missing stubs, and moves inline page.Bind functions out. The
// document is updated in place; the caller writes it only if WroteDocument
// is true.
func Apply(doc *scene.Document, source string) (Result, error) {
	if doc == nil {
		return Result{}, fmt.Errorf("sceneedit: nil document")
	}
	src, names, rep, err := EnsureHandlers(doc, source)
	if err != nil {
		return Result{}, err
	}
	src, rep2, err := ConvertInlineBindings(src, names)
	if err != nil {
		return Result{}, err
	}
	rep.Converted = append(rep.Converted, rep2.Converted...)
	rep.Bound = append(rep.Bound, rep2.Bound...)
	rep.Warnings = append(rep.Warnings, rep2.Warnings...)
	return Result{Source: src, Document: doc, Report: rep}, nil
}

// HandlerName is the convention: on_<node>_<event>, with dashes in node IDs
// and camelCase event names turned into underscores.
func HandlerName(nodeID, event string) string {
	return "on_" + identPart(nodeID) + "_" + snapshotCase(event)
}

func identPart(s string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		default:
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "node"
	}
	if unicode.IsDigit(rune(out[0])) {
		return "n" + out
	}
	return out
}

func snapshotCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && b.Len() > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if r == '-' || r == ' ' || r == '.' {
			if b.Len() > 0 {
				b.WriteByte('_')
			}
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// EnsureHandlers gives every call action a handler name and creates a stub
// for a handler that the script does not declare. It returns the names that
// must be added to the bindings section.
func EnsureHandlers(doc *scene.Document, source string) (string, []string, Report, error) {
	declared, err := declaredFunctions(source)
	if err != nil {
		return "", nil, Report{}, err
	}
	var rep Report
	var stubs []string
	var names []string
	seen := map[string]bool{}

	doc.Walk(func(n *scene.Node) bool {
		if len(n.On) == 0 {
			return true
		}
		events := make([]string, 0, len(n.On))
		for event := range n.On {
			events = append(events, event)
		}
		sort.Strings(events)
		for _, event := range events {
			actions := n.On[event]
			for i := range actions {
				a := &actions[i]
				if a.Do != "call" {
					continue
				}
				if a.Handler == "" {
					a.Handler = HandlerName(n.ID, event)
				}
				name := a.Handler
				if seen[name] {
					continue
				}
				seen[name] = true
				names = append(names, name)
				if declared[name] {
					continue
				}
				declared[name] = true
				stub := stubFor(name, n.ID, event)
				stubs = append(stubs, stub)
				rep.Created = append(rep.Created, name)
			}
			n.On[event] = actions
		}
		return true
	})

	src := source
	if len(stubs) > 0 {
		src, err = insertBeforeBind(src, strings.Join(stubs, "\n\n"))
		if err != nil {
			return "", nil, Report{}, err
		}
	}
	return src, names, rep, nil
}

func stubFor(name, nodeID, event string) string {
	return fmt.Sprintf("// %s handles %q on %q.\nlet %s = (evt) => {\n    // TODO: handle %s on %s\n}",
		name, event, nodeID, name, event, nodeID)
}

// declaredFunctions returns the top-level let names whose value is a
// function. It is intentionally liberal: a let of any value also counts as
// declared, so a rewrite never duplicates a name.
func declaredFunctions(source string) (map[string]bool, error) {
	prog, err := parser.Parse(context.Background(), source, nil)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, stmt := range prog.Stmts {
		switch x := stmt.(type) {
		case *ast.Var:
			if x.Name != nil {
				out[x.Name.Name] = true
			}
		case *ast.Assign:
			if x.Name != nil {
				out[x.Name.Name] = true
			}
		}
	}
	return out, nil
}

// ConvertInlineBindings moves function values out of page.Bind(...) into
// standalone let declarations and rewrites the bindings section at the same
// place. extra names are appended to the section.
func ConvertInlineBindings(source string, extra []string) (string, Report, error) {
	prog, err := parser.Parse(context.Background(), source, nil)
	if err != nil {
		return "", Report{}, err
	}
	var rep Report
	calls := bindCalls(prog)
	if len(calls) == 0 {
		if len(extra) == 0 {
			return source, rep, nil
		}
		names := dedupe(extra)
		rep.Bound = append(rep.Bound, names...)
		return source + "\n\n" + renderBind(nil, names) + "\n", rep, nil
	}

	declared, err := declaredFunctions(source)
	if err != nil {
		return "", Report{}, err
	}
	created := map[string]bool{}
	added := map[string]bool{}
	var edits []edit

	for callIndex, oc := range calls {
		call := oc.Call
		m, ok := call.Args[0].(*ast.Map)
		if !ok {
			rep.Warnings = append(rep.Warnings, fmt.Sprintf("page.Bind at line %d is not a map literal; left alone", call.Pos().LineNumber()))
			continue
		}
		var kept []string
		var names []string
		for _, item := range m.Items {
			name := mapKey(item.Key)
			if name == "" {
				kept = append(kept, exprText(source, item.Value))
				continue
			}
			if isShorthand(item) {
				names = append(names, name)
				continue
			}
			fn, isFunc := item.Value.(*ast.Func)
			if isFunc && fn.Name == nil {
				if !declared[name] {
					declared[name] = true
					created[name] = true
					rep.Converted = append(rep.Converted, name)
					// The standalone declaration goes in the same edit as the
					// rewritten binding, above it.
					names = append(names, name)
					continue
				}
				// Already a standalone let: only keep the shorthand binding.
				names = append(names, name)
				continue
			}
			kept = append(kept, exprText(source, item.Value))
			if item.Key != nil {
				kept[len(kept)-1] = source[runeToByte(source, item.Key.Pos().Char):runeToByte(source, item.Value.End().Char)]
			}
		}
		if callIndex == 0 {
			for _, name := range dedupe(extra) {
				if !contains(names, name) {
					names = append(names, name)
					rep.Bound = append(rep.Bound, name)
				}
			}
		}
		names = dedupe(names)
		var handlerTexts []string
		for _, name := range names {
			if created[name] && !added[name] {
				added[name] = true
				handlerTexts = append(handlerTexts, "let "+name+" = "+inlineFuncText(source, name, m))
			}
		}
		replacement := renderBind(kept, names)
		if len(handlerTexts) > 0 {
			replacement = strings.Join(handlerTexts, "\n\n") + "\n\n" + replacement
		}
		edits = append(edits, edit{start: oc.Pos().Char, end: oc.End().Char, text: replacement})
	}

	out := applyEdits(source, edits)
	if !strings.Contains(out, "page.Bind(") && len(extra) > 0 {
		out += "\n\n" + renderBind(nil, dedupe(extra)) + "\n"
	}
	return out, rep, nil
}

// inlineFuncText recovers the source text of the inline function for name.
func inlineFuncText(source, name string, m *ast.Map) string {
	for _, item := range m.Items {
		if mapKey(item.Key) != name {
			continue
		}
		if fn, ok := item.Value.(*ast.Func); ok {
			return dedentFunc(strings.TrimSpace(source[runeToByte(source, fn.Pos().Char):runeToByte(source, fn.End().Char)]))
		}
	}
	return "(evt) => {\n    // TODO\n}"
}

func bindCalls(prog *ast.Program) []*ast.ObjectCall {
	var out []*ast.ObjectCall
	for node := range ast.Preorder(prog) {
		oc, ok := node.(*ast.ObjectCall)
		if !ok || oc.Call == nil || len(oc.Call.Args) == 0 {
			continue
		}
		ident, ok := oc.X.(*ast.Ident)
		if !ok || ident.Name != "page" {
			continue
		}
		fn, ok := oc.Call.Fun.(*ast.Ident)
		if !ok || fn.Name != "Bind" {
			continue
		}
		out = append(out, oc)
	}
	return out
}

func mapKey(e ast.Expr) string {
	switch k := e.(type) {
	case *ast.Ident:
		return k.Name
	case *ast.String:
		if len(k.Exprs) == 0 {
			return k.Value
		}
	}
	return ""
}

func exprText(source string, e ast.Expr) string {
	if e == nil {
		return ""
	}
	return source[runeToByte(source, e.Pos().Char):runeToByte(source, e.End().Char)]
}

// runeToByte converts a lexer position (a rune index, not a byte offset) to
// a byte offset in source. The v2 lexer counts runes; slicing bytes directly
// breaks on any non-ASCII text, such as an em dash in a sentence.
func runeToByte(source string, index int) int {
	if index <= 0 {
		return 0
	}
	n := 0
	for i := range source {
		if n == index {
			return i
		}
		n++
	}
	return len(source)
}

func renderBind(kept, names []string) string {
	var lines []string
	lines = append(lines, kept...)
	lines = append(lines, names...)
	if len(lines) == 0 {
		return "page.Bind({})"
	}
	return "page.Bind({\n    " + strings.Join(lines, ",\n    ") + ",\n})"
}

func insertBeforeBind(source, block string) (string, error) {
	prog, err := parser.Parse(context.Background(), source, nil)
	if err != nil {
		return "", err
	}
	calls := bindCalls(prog)
	if len(calls) == 0 {
		if strings.TrimSpace(block) == "" {
			return source, nil
		}
		return strings.TrimRight(source, "\n") + "\n\n" + block + "\n", nil
	}
	pos := runeToByte(source, calls[0].Pos().Char)
	return source[:pos] + block + "\n\n" + source[pos:], nil
}

type edit struct {
	start, end int
	text       string
}

func applyEdits(source string, edits []edit) string {
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	for _, e := range edits {
		start, end := runeToByte(source, e.start), runeToByte(source, e.end)
		if start < 0 || end > len(source) || start > end {
			continue
		}
		source = source[:start] + e.text + source[end:]
	}
	return source
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// dedentFunc removes the common indentation of the lines after the first,
// so a handler lifted out of a page.Bind map lines up as a top-level let.
func dedentFunc(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return text
	}
	min := -1
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		n := 0
		for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
			n++
		}
		if min < 0 || n < min {
			min = n
		}
	}
	if min <= 0 {
		return text
	}
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) >= min {
			lines[i] = lines[i][min:]
		}
	}
	return strings.Join(lines, "\n")
}

func isShorthand(item ast.MapItem) bool {
	id, ok := item.Value.(*ast.Ident)
	if !ok || item.Key == nil || mapKey(item.Key) != id.Name {
		return false
	}
	return item.Key.Pos().Char == item.Value.Pos().Char
}

// ErrRemoteSidecar means the document's sidecar is not a local file, so this
// package cannot read or write it. Hosts with their own storage handle it.
var ErrRemoteSidecar = errors.New("sceneedit: sidecar is not a local file")

// ApplySidecar resolves doc.Script against docRef, rewrites the sidecar, and
// returns whether the document changed. The caller saves the document after
// it returns true (the designer's BeforeSave hook does exactly that).
func ApplySidecar(doc *scene.Document, docRef string) (Report, bool, error) {
	if doc == nil || doc.Script == "" {
		return Report{}, false, nil
	}
	ref, err := scene.ResolveRef(docRef, doc.Script)
	if err != nil {
		return Report{}, false, err
	}
	path, err := scene.FilePath(ref)
	if err != nil {
		return Report{}, false, fmt.Errorf("%w: %s", ErrRemoteSidecar, ref)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return Report{}, false, err
	}
	before, err := scene.Encode(doc)
	if err != nil {
		return Report{}, false, err
	}
	res, err := Apply(doc, string(src))
	if err != nil {
		return Report{}, false, err
	}
	if res.Source != string(src) {
		if err := os.WriteFile(path, []byte(res.Source), 0o644); err != nil {
			return res.Report, false, err
		}
	}
	after, err := scene.Encode(doc)
	if err != nil {
		return res.Report, false, err
	}
	return res.Report, !bytes.Equal(before, after), nil
}
