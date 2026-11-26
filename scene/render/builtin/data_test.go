package builtin

import (
	"slices"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

func dataDoc() *scene.Document {
	return &scene.Document{
		Title: "data",
		Root: &scene.Node{
			ID: "root", Type: "free",
			Props: map[string]any{"designW": 1200.0, "designH": 800.0, "scale": "none"},
			Children: []*scene.Node{
				{ID: "tbl", Type: "table", Place: scene.Place{X: 10, Y: 10, W: 560, H: 320},
					Props: map[string]any{"columns": []any{"A", "B"}}},
				{ID: "lg", Type: "log", Place: scene.Place{X: 10, Y: 340, W: 560, H: 240},
					Props: map[string]any{"text": "line1\nline2", "maxLines": 10.0}},
				{ID: "fm", Type: "form", Place: scene.Place{X: 600, Y: 10, W: 340, H: 240},
					Children: []*scene.Node{
						{ID: "f1", Type: "field", Props: map[string]any{"label": "Name", "help": "your name"},
							Children: []*scene.Node{{ID: "e1", Type: "entry"}}},
						{ID: "bare", Type: "button", Props: map[string]any{"text": "Bare"}},
					}},
			},
		},
	}
}

func TestTableRowModel(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	v := render.NewView(dataDoc(), render.Options{Registry: NewRegistry()})
	w := test.NewWindow(v)
	defer w.Close()
	w.Resize(fyne.NewSize(1000, 700))

	l, ok := v.Live("tbl")
	if !ok {
		t.Fatal("no live for tbl")
	}
	rm, ok := render.RowModelOf(l)
	if !ok {
		t.Fatal("table live does not expose RowModel through the hint wrapper")
	}
	rm.SetRows([][]string{{"1", "one"}, {"2", "two"}, {"3", "three"}})
	if !rm.SetCell(0, 1, "uno") {
		t.Fatal("SetCell(0,1) out of range")
	}
	if rm.SetCell(9, 0, "x") {
		t.Fatal("SetCell(9,0) should be out of range")
	}
	tl := render.UnwrapLive(l).(*tableLive)
	if len(tl.rows) != 3 || tl.rows[0][1] != "uno" {
		t.Fatalf("rows = %v", tl.rows)
	}
	if td := tl.page(0, 2); td.RowCount() != 2 || td.Get(0, 0) != "1" {
		t.Fatalf("page(0,2) = %d rows, first cell %q", td.RowCount(), td.Get(0, 0))
	}
	if td := tl.page(2, 2); td.RowCount() != 1 || td.Get(1, 0) != "three" {
		t.Fatalf("page(2,2) = %d rows", td.RowCount())
	}
}

func TestLogAppender(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	v := render.NewView(dataDoc(), render.Options{Registry: NewRegistry()})
	w := test.NewWindow(v)
	defer w.Close()

	l, ok := v.Live("lg")
	if !ok {
		t.Fatal("no live for lg")
	}
	al, ok := render.AppenderOf(l)
	if !ok {
		t.Fatal("log live does not expose Appender")
	}
	if ll := render.UnwrapLive(l).(*logLive); len(ll.lines) != 2 {
		t.Fatalf("initial lines = %v", ll.lines)
	}
	al.AppendText("line3")
	al.AppendText("line4\nline5")
	ll := render.UnwrapLive(l).(*logLive)
	if len(ll.lines) != 5 || ll.lines[4] != "line5" {
		t.Fatalf("lines = %v", ll.lines)
	}
	// maxLines cap from props (10): append 10 more, keep the newest 10.
	for i := 0; i < 10; i++ {
		al.AppendText("x")
	}
	if len(ll.lines) != 10 || ll.lines[0] != "x" || ll.lines[9] != "x" {
		t.Fatalf("cap: lines = %v", ll.lines)
	}
}

func TestFormFieldItems(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	v := render.NewView(dataDoc(), render.Options{Registry: NewRegistry()})
	w := test.NewWindow(v)
	defer w.Close()

	l, ok := v.Live("fm")
	if !ok {
		t.Fatal("no live for fm")
	}
	fl := render.UnwrapLive(l).(*formLive)
	if len(fl.form.Items) != 2 {
		t.Fatalf("form items = %d", len(fl.form.Items))
	}
	it := fl.form.Items[0]
	if it.Text != "Name" || it.HintText != "your name" || it.Widget == nil {
		t.Fatalf("field item = %+v", it)
	}
	fw, ok := it.Widget.(fyne.Widget)
	if !ok || fw == nil {
		t.Fatalf("field widget is %T, want a wrapped entry", it.Widget)
	}
	// The bare button becomes an unlabeled row, not a broken page.
	if fl.form.Items[1].Text != "" {
		t.Fatalf("bare child item = %+v", fl.form.Items[1])
	}

	// The field's input sits in the enclosing form only.
	fld := render.UnwrapLive(mustLive(t, v, "f1")).(*fieldLive)
	if len(fld.form.Items) != 0 {
		t.Fatalf("attached field's own form has %d items", len(fld.form.Items))
	}
}

func mustLive(t *testing.T, v *render.View, id string) render.Live {
	t.Helper()
	l, ok := v.Live(id)
	if !ok {
		t.Fatalf("no live %s", id)
	}
	return l
}

// formTexts returns every text shown in a form: labels (rich text) and
// hint lines (canvas text).
func formTexts(f *widget.Form) []string {
	var out []string
	for _, o := range treeObjects(f) {
		switch x := o.(type) {
		case *widget.RichText:
			out = append(out, x.String())
		case *canvas.Text:
			if x.Visible() {
				out = append(out, x.Text)
			}
		}
	}
	return out
}

// Runtime label/help changes of a field reach the enclosing form's rendered
// rows (#64.3: fieldLive.Update only re-synced its own, undisplayed form).
func TestFormFieldRuntimeLabelAndHelp(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	v := render.NewView(dataDoc(), render.Options{Registry: NewRegistry()})
	w := test.NewWindow(v)
	defer w.Close()
	w.Resize(fyne.NewSize(1000, 700))
	fl := render.UnwrapLive(mustLive(t, v, "fm")).(*formLive)

	if !v.RuntimeSet("f1", "label", "Full name") {
		t.Fatal("RuntimeSet(label) failed")
	}
	if fl.form.Items[0].Text != "Full name" {
		t.Fatalf("form item label = %q", fl.form.Items[0].Text)
	}
	if !slices.Contains(formTexts(fl.form), "Full name") {
		t.Fatalf("rendered form texts %q lack the new label", formTexts(fl.form))
	}

	v.RuntimeSet("f1", "help", "first and last")
	if got := formTexts(fl.form); !slices.Contains(got, "first and last") || slices.Contains(got, "your name") {
		t.Fatalf("rendered form texts %q, want the new help only", got)
	}

	// A hint added to a row that had none shows too.
	doc := dataDoc()
	doc.Root.Children[2].Children[0].Props["help"] = ""
	v2 := render.NewView(doc, render.Options{Registry: NewRegistry()})
	w2 := test.NewWindow(v2)
	defer w2.Close()
	v2.RuntimeSet("f1", "help", "now with help")
	fl2 := render.UnwrapLive(mustLive(t, v2, "fm")).(*formLive)
	if got := formTexts(fl2.form); !slices.Contains(got, "now with help") {
		t.Fatalf("rendered form texts %q lack the added help", got)
	}
}
