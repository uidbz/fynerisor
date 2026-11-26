// Package handlers rewrites scene sidecar scripts for the editor workflow.
//
// The convention is deliberately small and visible:
//
//   - one standalone handler per node event, named on_<node>_<event>;
//   - handlers are ordinary let-bound functions, never inline in page.Bind;
//   - a page.Bind({...}) section lists the handler names.
//
// The rewrite edits the source in place: it lifts inline page.Bind functions
// to let declarations above their page.Bind call, adds stubs for missing
// handlers, and appends new names to the first page.Bind map. Everything else,
// including comments and formatting, is left as written.
//
// The package works on source text and the scene document only. It does not
// execute a script, touch Fyne, or depend on the designer.
package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
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

func (r *Report) warnf(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// Result is a rewritten script and document.
type Result struct {
	Source   string
	Document *scene.Document
	Report   Report
}

// Apply runs the whole editor pass: it gives every call action a handler
// name, moves inline page.Bind functions out, creates stubs for missing
// handlers and binds them. The document is updated in place; the caller
// decides whether to write it (compare its encoding before and after).
//
// A script that does not parse is returned unchanged with a warning; the
// document still gets its handler names, so a host can save layout work while
// the author is half-way through editing a handler.
func Apply(doc *scene.Document, source string) (Result, error) {
	if doc == nil {
		return Result{}, fmt.Errorf("handlers: nil document")
	}
	var rep Report
	refs := nameHandlers(doc, &rep)
	s, err := parse(source)
	if err != nil {
		rep.warnf("script not rewritten: %v", err)
		return Result{Source: source, Document: doc, Report: rep}, nil
	}
	out := s.rewrite(refs, true, &rep)
	return Result{Source: out, Document: doc, Report: rep}, nil
}

// EnsureHandlers gives every call action a handler name and creates a stub
// for a handler that the script neither declares nor binds inline. It
// returns the names that must be added to the bindings section; pass them to
// ConvertInlineBindings. Apply does both in one pass.
func EnsureHandlers(doc *scene.Document, source string) (string, []string, Report, error) {
	if doc == nil {
		return "", nil, Report{}, fmt.Errorf("handlers: nil document")
	}
	s, err := parse(source)
	if err != nil {
		return "", nil, Report{}, err
	}
	var rep Report
	refs := nameHandlers(doc, &rep)
	out := s.addStubs(refs, &rep)
	names := make([]string, 0, len(refs))
	for _, r := range refs {
		names = append(names, r.name)
	}
	return out, names, rep, nil
}

// ConvertInlineBindings moves function values out of page.Bind(...) into
// standalone let declarations above the call and binds the extra names that
// are not bound yet. Items are edited in place; a script that needs no change
// is returned as is.
func ConvertInlineBindings(source string, extra []string) (string, Report, error) {
	s, err := parse(source)
	if err != nil {
		return "", Report{}, err
	}
	refs := make([]handlerRef, 0, len(extra))
	for _, name := range extra {
		refs = append(refs, handlerRef{name: name})
	}
	var rep Report
	return s.rewrite(refs, false, &rep), rep, nil
}

// handlerRef is a handler name used by a call action, with the first node
// event that uses it (for the stub comment).
type handlerRef struct {
	name, node, event string
}

// HandlerName is the convention: on_<node>_<event>, with dashes in node IDs
// and camelCase event names turned into underscores. The result is always a
// Risor identifier. Different node/event pairs can map to one name (a-b and
// a--b); NameFor resolves that against a document.
func HandlerName(nodeID, event string) string {
	return "on_" + identPart(nodeID) + "_" + snakeCase(event)
}

// NameFor returns the conventional handler name for a node event in doc. If
// another node event already uses that name, a numeric suffix keeps the
// handlers apart (on_table_cell_tapped_2).
func NameFor(doc *scene.Document, nodeID, event string) string {
	taken := map[string]bool{}
	if doc != nil {
		doc.Walk(func(n *scene.Node) bool {
			for ev, acts := range n.On {
				if n.ID == nodeID && ev == event {
					continue
				}
				for _, a := range acts {
					if a.Do == "call" && a.Handler != "" {
						taken[a.Handler] = true
					}
				}
			}
			return true
		})
	}
	base := HandlerName(nodeID, event)
	name := base
	for i := 2; taken[name]; i++ {
		name = fmt.Sprintf("%s_%d", base, i)
	}
	return name
}

// nameHandlers fills every empty call handler and returns the handler names
// in document order, each once.
func nameHandlers(doc *scene.Document, rep *Report) []handlerRef {
	var refs []handlerRef
	seen := map[string]bool{}
	doc.Walk(func(n *scene.Node) bool {
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
					a.Handler = NameFor(doc, n.ID, event)
					if base := HandlerName(n.ID, event); a.Handler != base {
						rep.warnf("handler name %s for %q %s is used by another node event; using %s", base, n.ID, event, a.Handler)
					}
				}
				if !seen[a.Handler] {
					seen[a.Handler] = true
					refs = append(refs, handlerRef{name: a.Handler, node: n.ID, event: event})
				}
			}
		}
		return true
	})
	return refs
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

// snakeCase turns an event name into snake_case: cellTapped -> cell_tapped,
// HTTPDone -> http_done. Characters that cannot appear in an identifier
// become separators.
func snakeCase(s string) string {
	rs := []rune(s)
	var b strings.Builder
	sep := func() {
		if b.Len() > 0 && !strings.HasSuffix(b.String(), "_") {
			b.WriteByte('_')
		}
	}
	for i, r := range rs {
		switch {
		case unicode.IsUpper(r):
			prevLower := i > 0 && (unicode.IsLower(rs[i-1]) || unicode.IsDigit(rs[i-1]))
			acronymEnd := i > 0 && unicode.IsUpper(rs[i-1]) && i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if prevLower || acronymEnd {
				sep()
			}
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			sep()
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "event"
	}
	return out
}

// IsIdent reports whether s can be used as a Risor variable name, and so as
// a let-bound handler and a shorthand page.Bind key.
func IsIdent(s string) bool {
	for i, r := range s {
		if !(unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r))) {
			return false
		}
	}
	if s == "" {
		return false
	}
	// Keywords pass the character check; the parser rejects them.
	prog, err := parser.Parse(context.Background(), "let "+s+" = 0", nil)
	if err != nil || len(prog.Stmts) != 1 {
		return false
	}
	v, ok := prog.Stmts[0].(*ast.Var)
	return ok && v.Name != nil && v.Name.Name == s
}

// script is one parse of a sidecar with the facts the rewrite needs.
type script struct {
	src  string
	nl   string // line ending for generated text
	off  []int  // byte offset of each rune index (lexer positions count runes)
	prog *ast.Program
	// declared holds every top-level name: let, const, multi-name let,
	// destructuring, function declarations and plain assignments.
	declared map[string]bool
	// bound holds every key of every page.Bind map literal in the script.
	bound map[string]bool
	// binds are the page.Bind calls with a map literal that are not nested
	// in a function or block, in source order. Their inline functions can
	// move to top-level lets.
	binds []bindSite
	notes []string
}

type bindSite struct {
	stmt ast.Node
	m    *ast.Map
	line int
}

func parse(source string) (*script, error) {
	prog, err := parser.Parse(context.Background(), source, nil)
	if err != nil {
		return nil, err
	}
	s := &script{src: source, nl: "\n", prog: prog, declared: map[string]bool{}, bound: map[string]bool{}}
	if strings.Contains(source, "\r\n") {
		s.nl = "\r\n"
	}
	s.off = make([]int, 0, len(source)+1)
	for i := range source {
		s.off = append(s.off, i)
	}
	s.off = append(s.off, len(source))

	for _, stmt := range prog.Stmts {
		for _, name := range declaredNames(stmt) {
			s.declared[name] = true
		}
		var scopes []ast.Node // functions and blocks inside stmt
		var calls []*ast.ObjectCall
		for node := range ast.Preorder(stmt) {
			switch x := node.(type) {
			case *ast.Func, *ast.Block:
				scopes = append(scopes, x)
			case *ast.ObjectCall:
				if isBindCall(x) {
					calls = append(calls, x)
				}
			}
		}
		for _, call := range calls {
			line := call.Pos().LineNumber()
			m, ok := call.Call.Args[0].(*ast.Map)
			if !ok {
				s.notes = append(s.notes, fmt.Sprintf("page.Bind at line %d is not a map literal; left alone", line))
				continue
			}
			for _, item := range m.Items {
				if name := mapKey(item.Key); name != "" {
					s.bound[name] = true
				}
			}
			nested := slices.ContainsFunc(scopes, func(n ast.Node) bool {
				return n.Pos().Char <= call.Pos().Char && call.End().Char <= n.End().Char
			})
			if !nested {
				s.binds = append(s.binds, bindSite{stmt: stmt, m: m, line: line})
			}
		}
	}
	return s, nil
}

func declaredNames(stmt ast.Node) []string {
	var out []string
	add := func(id *ast.Ident) {
		if id != nil {
			out = append(out, id.Name)
		}
	}
	switch x := stmt.(type) {
	case *ast.Var:
		add(x.Name)
	case *ast.Const:
		add(x.Name)
	case *ast.MultiVar:
		for _, id := range x.Names {
			add(id)
		}
	case *ast.ObjectDestructure:
		for _, b := range x.Bindings {
			if b.Alias != "" {
				out = append(out, b.Alias)
			} else {
				out = append(out, b.Key)
			}
		}
	case *ast.ArrayDestructure:
		for _, e := range x.Elements {
			add(e.Name)
		}
	case *ast.Func:
		add(x.Name)
	case *ast.Assign:
		add(x.Name)
	}
	return out
}

func isBindCall(oc *ast.ObjectCall) bool {
	if oc.Call == nil || len(oc.Call.Args) == 0 {
		return false
	}
	ident, ok := oc.X.(*ast.Ident)
	if !ok || ident.Name != "page" {
		return false
	}
	fn, ok := oc.Call.Fun.(*ast.Ident)
	return ok && fn.Name == "Bind"
}

// at converts a lexer position (a rune index) to a byte offset.
func (s *script) at(char int) int {
	return s.off[max(0, min(char, len(s.off)-1))]
}

// end is the byte offset after e. ast.String.End is unreliable (it does not
// count the quotes), so strings anywhere in e are measured from the source.
func (s *script) end(e ast.Node) int {
	out := s.at(e.End().Char)
	for node := range ast.Preorder(e) {
		if str, ok := node.(*ast.String); ok {
			_, stop := s.stringSpan(str)
			out = max(out, stop)
		}
	}
	return out
}

// stringSpan returns the byte range of a string literal including quotes.
func (s *script) stringSpan(str *ast.String) (int, int) {
	start := s.at(str.Pos().Char)
	if start >= len(s.src) {
		return start, start
	}
	q := s.src[start]
	for i := start + 1; i < len(s.src); i++ {
		switch s.src[i] {
		case '\\':
			if q != '`' {
				i++
			}
		case q:
			return start, i + 1
		}
	}
	return start, len(s.src)
}

type edit struct {
	start, end int // byte offsets
	text       string
}

// rewrite lifts inline functions, optionally creates stubs, and binds the
// referenced names that are not bound yet.
func (s *script) rewrite(refs []handlerRef, stubs bool, rep *Report) string {
	rep.Warnings = append(rep.Warnings, s.notes...)
	declared := maps.Clone(s.declared)
	blocks := map[int][]string{} // statement start -> declarations to insert
	var edits []edit

	for _, b := range s.binds {
		start := s.at(b.stmt.Pos().Char)
		for _, item := range b.m.Items {
			name := mapKey(item.Key)
			fn, isFunc := item.Value.(*ast.Func)
			if !IsIdent(name) || isShorthand(item) || !isFunc || fn.Name != nil {
				continue
			}
			if declared[name] {
				rep.warnf("page.Bind at line %d: inline handler %s kept because %s is already declared", b.line, name, name)
				continue
			}
			declared[name] = true
			rep.Converted = append(rep.Converted, name)
			blocks[start] = append(blocks[start], "let "+name+" = "+s.dedent(fn))
			edits = append(edits, edit{s.at(item.Key.Pos().Char), s.end(fn), name})
		}
	}

	var created, bind []string
	for _, r := range refs {
		switch {
		case r.name == "" || s.bound[r.name] || slices.Contains(bind, r.name):
		case !IsIdent(r.name):
			rep.warnf("handler %q is not a Risor identifier; bind it with a quoted key in page.Bind", r.name)
		default:
			if stubs && !declared[r.name] {
				declared[r.name] = true
				created = append(created, s.stub(r))
				rep.Created = append(rep.Created, r.name)
			}
			bind = append(bind, r.name)
		}
	}
	rep.Bound = append(rep.Bound, bind...)

	tail := ""
	if len(s.binds) > 0 {
		target := s.binds[0]
		start := s.at(target.stmt.Pos().Char)
		blocks[start] = append(blocks[start], created...)
		if len(bind) > 0 {
			edits = append(edits, s.appendItems(target.m, bind)...)
		}
	} else if len(created) > 0 || len(bind) > 0 {
		parts := append(created, "page.Bind({"+s.nl+"    "+strings.Join(bind, ","+s.nl+"    ")+","+s.nl+"})")
		tail = strings.Join(parts, s.nl+s.nl) + s.nl
	}
	for start, decls := range blocks {
		if len(decls) > 0 {
			edits = append(edits, edit{start, start, strings.Join(decls, s.nl+s.nl) + s.nl + s.nl})
		}
	}

	out := applyEdits(s.src, edits)
	if tail != "" {
		if trimmed := strings.TrimRight(out, "\r\n"); trimmed != "" {
			out = trimmed + s.nl + s.nl
		} else {
			out = ""
		}
		out += tail
	}
	return out
}

// addStubs inserts stubs for refs that are neither declared nor bound, above
// the first page.Bind call (or at the end).
func (s *script) addStubs(refs []handlerRef, rep *Report) string {
	var created []string
	seen := map[string]bool{}
	for _, r := range refs {
		if seen[r.name] || s.declared[r.name] || s.bound[r.name] {
			continue
		}
		seen[r.name] = true
		if !IsIdent(r.name) {
			rep.warnf("handler %q is not a Risor identifier; bind it with a quoted key in page.Bind", r.name)
			continue
		}
		created = append(created, s.stub(r))
		rep.Created = append(rep.Created, r.name)
	}
	if len(created) == 0 {
		return s.src
	}
	block := strings.Join(created, s.nl+s.nl)
	if len(s.binds) > 0 {
		at := s.at(s.binds[0].stmt.Pos().Char)
		return s.src[:at] + block + s.nl + s.nl + s.src[at:]
	}
	if trimmed := strings.TrimRight(s.src, "\r\n"); trimmed != "" {
		return trimmed + s.nl + s.nl + block + s.nl
	}
	return block + s.nl
}

func (s *script) stub(r handlerRef) string {
	if r.node == "" {
		return fmt.Sprintf("let %s = (evt) => {%s    // TODO%s}", r.name, s.nl, s.nl)
	}
	return fmt.Sprintf("// %s handles %q on %q.%slet %s = (evt) => {%s    // TODO: handle %s on %s%s}",
		r.name, r.event, r.node, s.nl, r.name, s.nl, r.event, r.node, s.nl)
}

// appendItems adds shorthand names after the last item of m, keeping the
// map's layout: one name per line when the closing brace is on its own line,
// otherwise on the brace's line.
func (s *script) appendItems(m *ast.Map, names []string) []edit {
	rb := s.at(m.Rbrace.Char)
	lb := s.at(m.Lbrace.Char)
	ls := lineStart(s.src, rb)
	ownLine := ls > lb && strings.TrimSpace(s.src[ls:rb]) == ""
	braceIndent := indentAt(s.src, rb)
	if len(m.Items) == 0 {
		ind := braceIndent + "    "
		if ownLine {
			return []edit{{ls, ls, ind + strings.Join(names, ","+s.nl+ind) + "," + s.nl}}
		}
		return []edit{{rb, rb, s.nl + ind + strings.Join(names, ","+s.nl+ind) + "," + s.nl + braceIndent}}
	}
	last := m.Items[len(m.Items)-1]
	lastEnd := s.end(last.Value)
	var out []edit
	if !hasComma(s.src[lastEnd:rb]) {
		out = append(out, edit{lastEnd, lastEnd, ","})
	}
	if !ownLine {
		return append(out, edit{rb, rb, " " + strings.Join(names, ", ")})
	}
	first := last.Value
	if last.Key != nil {
		first = last.Key
	}
	ind := indentAt(s.src, s.at(first.Pos().Char))
	if lineStart(s.src, s.at(first.Pos().Char)) <= lb {
		ind = braceIndent + "    "
	}
	return append(out, edit{ls, ls, ind + strings.Join(names, ","+s.nl+ind) + "," + s.nl})
}

func lineStart(src string, at int) int {
	return strings.LastIndexByte(src[:at], '\n') + 1
}

// indentAt is the leading whitespace of the line containing byte offset at.
func indentAt(src string, at int) string {
	ls := lineStart(src, at)
	n := ls
	for n < len(src) && (src[n] == ' ' || src[n] == '\t') {
		n++
	}
	return src[ls:n]
}

// hasComma reports whether the text between two map items (whitespace,
// commas and comments only) contains a comma outside comments.
func hasComma(gap string) bool {
	for i := 0; i < len(gap); i++ {
		switch {
		case strings.HasPrefix(gap[i:], "//"):
			if j := strings.IndexByte(gap[i:], '\n'); j >= 0 {
				i += j
			} else {
				return false
			}
		case strings.HasPrefix(gap[i:], "/*"):
			if j := strings.Index(gap[i+2:], "*/"); j >= 0 {
				i += j + 3
			} else {
				return false
			}
		case gap[i] == ',':
			return true
		}
	}
	return false
}

// dedent returns the source of fn with the common indentation of the lines
// after the first removed, so a handler lifted out of a page.Bind map lines
// up as a top-level let. Lines that start inside a multi-line string are part
// of the string's value and are left alone.
func (s *script) dedent(fn *ast.Func) string {
	start, stop := s.at(fn.Pos().Char), s.end(fn)
	var strs [][2]int
	for node := range ast.Preorder(fn) {
		if str, ok := node.(*ast.String); ok {
			a, b := s.stringSpan(str)
			strs = append(strs, [2]int{a, b})
		}
	}
	lines := strings.SplitAfter(s.src[start:stop], "\n")
	inString := make([]bool, len(lines))
	pos := start
	for i, line := range lines {
		for _, r := range strs {
			if i > 0 && r[0] < pos && pos < r[1] {
				inString[i] = true
			}
		}
		pos += len(line)
	}
	common := -1
	for i, line := range lines[1:] {
		if inString[i+1] || strings.TrimSpace(line) == "" {
			continue
		}
		n := len(line) - len(strings.TrimLeft(line, " \t"))
		if common < 0 || n < common {
			common = n
		}
	}
	if common <= 0 {
		return s.src[start:stop]
	}
	for i := 1; i < len(lines); i++ {
		if !inString[i] && len(lines[i]) >= common && strings.TrimSpace(lines[i][:common]) == "" {
			lines[i] = lines[i][common:]
		}
	}
	return strings.Join(lines, "")
}

// applyEdits applies non-overlapping edits given in byte offsets. Insertions
// at the same offset come out in the order they were added.
func applyEdits(source string, edits []edit) string {
	order := make([]int, len(edits))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		ea, eb := edits[order[a]], edits[order[b]]
		if ea.start != eb.start {
			return ea.start > eb.start
		}
		return order[a] > order[b]
	})
	for _, i := range order {
		e := edits[i]
		if e.start < 0 || e.end > len(source) || e.start > e.end {
			continue
		}
		source = source[:e.start] + e.text + source[e.end:]
	}
	return source
}

// mapKey is the constant key of a map item: an identifier or a plain string.
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

func isShorthand(item ast.MapItem) bool {
	id, ok := item.Value.(*ast.Ident)
	if !ok || item.Key == nil || mapKey(item.Key) != id.Name {
		return false
	}
	return item.Key.Pos().Char == item.Value.Pos().Char
}

// ErrRemoteSidecar means the document's sidecar is not a local file, so this
// package cannot read or write it. Hosts with their own storage handle it.
var ErrRemoteSidecar = errors.New("handlers: sidecar is not a local file")

// ApplySidecar resolves doc.Script against docRef, rewrites the sidecar, and
// returns whether the document changed. The caller saves the document after
// it returns true (the designer's BeforeSave hook does exactly that).
//
// Handler names are filled in on every path, so changed is meaningful even
// with an error: without a sidecar (doc.Script empty), with a missing sidecar
// (an error matching os.ErrNotExist) and with ErrRemoteSidecar. A sidecar that
// does not parse is not rewritten; that is a Report warning, not an error.
// The sidecar is replaced atomically and keeps its file mode.
func ApplySidecar(doc *scene.Document, docRef string) (Report, bool, error) {
	if doc == nil {
		return Report{}, false, nil
	}
	before, _ := scene.Encode(doc)
	changed := func() bool {
		after, err := scene.Encode(doc)
		return err == nil && !bytes.Equal(before, after)
	}
	var rep Report
	nameHandlers(doc, &rep)
	if doc.Script == "" {
		return rep, changed(), nil
	}
	ref, err := scene.ResolveRef(docRef, doc.Script)
	if err != nil {
		return rep, changed(), err
	}
	path, err := scene.FilePath(ref)
	if err != nil {
		return rep, changed(), fmt.Errorf("%w: %s", ErrRemoteSidecar, ref)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return rep, changed(), err
	}
	res, err := Apply(doc, string(src))
	if err != nil {
		return rep, changed(), err
	}
	// Collision warnings were reported by the first naming pass.
	res.Report.Warnings = append(rep.Warnings, res.Report.Warnings...)
	if res.Source != string(src) {
		if err := WriteFile(path, []byte(res.Source)); err != nil {
			return res.Report, changed(), err
		}
	}
	return res.Report, changed(), nil
}

// WriteFile replaces path with data atomically: it writes a temporary file in
// the same directory and renames it over path, so a failed write never leaves
// a truncated script. An existing file keeps its permission bits; a new file
// gets 0644.
func WriteFile(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, err = tmp.Write(data)
	if err == nil {
		err = tmp.Chmod(mode)
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		os.Remove(name)
	}
	return err
}
