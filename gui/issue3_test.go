package gui

import (
	"context"
	"strings"
	"testing"

	fyneewidget "fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
	guiw "github.com/uidbz/fynerisor/gui/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Issue #3: Markdown gains SetText and a readable/writable .Text.
func TestMarkdownSetText(t *testing.T) {
	guithread.SetMain() // run bound GUI ops inline (no running event loop)
	_ = newTestApp(t)

	md := guiw.NewMarkdown("# hello")
	attr, ok := md.GetAttr("Text")
	if !ok {
		t.Fatal("Markdown.Text missing")
	}
	if attr.(*object.String).Value() != "# hello" {
		t.Fatalf("Markdown.Text initial: %q", attr.(*object.String).Value())
	}

	setText, _ := md.GetAttr("SetText")
	if _, err := setText.(*object.Builtin).Call(context.Background(), object.NewString("## updated")); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	attr, _ = md.GetAttr("Text")
	if attr.(*object.String).Value() != "## updated" {
		t.Fatalf("Markdown.Text after SetText: %q", attr.(*object.String).Value())
	}

	if err := md.SetAttr("Text", object.NewString("# via attr")); err != nil {
		t.Fatalf("Markdown.Text set: %v", err)
	}
	attr, _ = md.GetAttr("Text")
	if attr.(*object.String).Value() != "# via attr" {
		t.Fatalf("Markdown.Text after SetAttr: %q", attr.(*object.String).Value())
	}

	// The underlying RichText really re-parsed.
	rt, ok := md.Interface().(*fyneewidget.RichText)
	if !ok {
		t.Fatalf("Markdown instance type: %T", md.Interface())
	}
	joined := &strings.Builder{}
	for _, seg := range rt.Segments {
		if ts, ok := seg.(*fyneewidget.TextSegment); ok {
			joined.WriteString(ts.Text)
		}
	}
	if !strings.Contains(joined.String(), "via attr") {
		t.Fatalf("RichText segments not updated: %q", joined.String())
	}
}
