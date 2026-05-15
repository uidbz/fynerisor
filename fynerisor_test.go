package fynerisor

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestWidgetCreation tests that all widget types can be created
func TestWidgetCreation(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"Button", `widget.NewButton("Test", () => {})`},
		{"Label", `widget.NewLabel("Hello")`},
		{"Entry", `widget.NewEntry()`},
		{"Check", `widget.NewCheck("Test", (checked) => {})`},
		{"RadioGroup", `widget.NewRadioGroup(["A", "B"], (s) => {})`},
		{"Slider", `widget.NewSlider(0.0, 100.0)`},
		{"ProgressBar", `widget.NewProgressBar()`},
		{"Icon", `widget.NewIcon("search")`},
		{"Card", `widget.NewCard("T", "S", widget.NewLabel("C"))`},
		{"Accordion", `widget.NewAccordion(widget.NewAccordionItem("T", "", nil))`},
		{"Toolbar", `widget.NewToolbar(widget.NewToolbarAction("save", () => {}))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := test.NewApp()
			defer a.Quit()
			w := a.NewWindow("Test")
			fw := NewWindow(w)
			fw.LoadScript(tt.script)
			fw.Execute()
			time.Sleep(10 * time.Millisecond)
		})
	}
}

// TestContainerCreation tests container creation
func TestContainerCreation(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"VBox", `container.NewVBox(widget.NewLabel("1"), widget.NewLabel("2"))`},
		{"HBox", `container.NewHBox(widget.NewLabel("1"), widget.NewLabel("2"))`},
		{"Border", `container.NewBorder(widget.NewLabel("T"), nil, nil, nil, widget.NewLabel("C"))`},
		{"Scroll", `container.NewScroll(widget.NewLabel("Content"))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := test.NewApp()
			defer a.Quit()
			w := a.NewWindow("Test")
			fw := NewWindow(w)
			fw.LoadScript(tt.script)
			fw.Execute()
			time.Sleep(10 * time.Millisecond)
		})
	}
}

// TestWidgetProperties tests widget property access
func TestWidgetProperties(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let lbl = widget.NewLabel("Initial")
lbl.Text = "Updated"
if (lbl.Text != "Updated") {
	throw("Label text not updated")
}

let slider = widget.NewSlider(0.0, 100.0)
slider.Value = 50.0
if (slider.Value != 50.0) {
	throw("Slider value not set")
}

let pb = widget.NewProgressBar()
pb.Value = 0.75
if (pb.Value != 0.75) {
	throw("ProgressBar value not set")
}
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if strings.Contains(fw.Status, "ERROR") {
		t.Errorf("Script failed: %s", fw.Status)
	}
}

// TestComplexLayout tests nested containers and widgets
func TestComplexLayout(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let title = widget.NewLabel("Title")
let btn = widget.NewButton("Save", () => {})
let header = container.NewHBox(title, btn)

let content = widget.NewLabel("Content")
let footer = widget.NewLabel("Footer")

let layout = container.NewBorder(header, footer, nil, nil, content)
window.SetContent(layout)
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got %s", fw.Status)
	}
}

// TestCallbacks tests widget callbacks
func TestCallbacks(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let clicked = false
let btn = widget.NewButton("Click", () => {
	clicked = true
})

let checked = false
let chk = widget.NewCheck("Check", (c) => {
	checked = c
})
chk.SetChecked(true)

let selected = ""
let rg = widget.NewRadioGroup(["A", "B"], (s) => {
	selected = s
})
rg.SetSelected("B")

window.SetContent(container.NewVBox(btn, chk, rg))
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got %s", fw.Status)
	}
}

// TestTableWidget tests table functionality
func TestTableWidget(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let table = widget.NewTable("Test", 10)

table.Columns(() => {
	return ["ID", "Name", "Status"]
})

table.RowCount(() => {
	return 25
})

table.Data((offset, limit) => {
	return [
		[string(offset), "Item 1", "Active"],
		[string(offset + 1), "Item 2", "Active"]
	]
})

window.SetContent(table)
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got %s", fw.Status)
	}
}

// TestAccordionWidget tests accordion functionality
func TestAccordionWidget(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let item1 = widget.NewAccordionItem("Section 1", "", widget.NewLabel("Content 1"))
let item2 = widget.NewAccordionItem("Section 2", "", widget.NewLabel("Content 2"))

let acc = widget.NewAccordion(item1, item2)
acc.MultiOpen = true
acc.Open(0)

window.SetContent(acc)
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got %s", fw.Status)
	}
}

// TestToolbarWidget tests toolbar functionality
func TestToolbarWidget(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let save = widget.NewToolbarAction("documentSave", () => {})
let open = widget.NewToolbarAction("folderOpen", () => {})
let sep = widget.NewToolbarSeparator()
let spacer = widget.NewToolbarSpacer()
let settings = widget.NewToolbarAction("settings", () => {})

let toolbar = widget.NewToolbar(save, open, sep, spacer, settings)

window.SetContent(toolbar)
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got %s", fw.Status)
	}
}

// TestFormWidget tests form functionality
func TestFormWidget(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let nameEntry = widget.NewEntry()
let emailEntry = widget.NewEntry()

let item1 = widget.NewFormItem("Name:", nameEntry)
let item2 = widget.NewFormItem("Email:", emailEntry)

let form = widget.NewForm(item1, item2)

window.SetContent(form)
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond) // Increased timeout for async execution

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got %s", fw.Status)
	}
}
