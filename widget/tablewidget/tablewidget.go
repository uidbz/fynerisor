package tablewidget

import (
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"golang.design/x/clipboard"

	"github.com/xuri/excelize/v2"
)

//go:embed img/navigate-first.svg
var navigateFirst []byte

//go:embed img/navigate-last.svg
var navigateLast []byte

type TableWidget struct {
	table                *FlexTable
	currentPage          binding.Int
	currentFilter        string
	currentObjectUid     binding.String
	currentData          *TableData
	filteredData         *TableData
	totalPages           binding.Int
	totalResults         binding.Int
	fieldSelectionWidget *widget.Select // for attachments
	Instance             *fyne.Container
	Title                string
	pageSize             int
	RowCount             func() int
	Data                 func(offset, limit int) *TableData
	Offset               int
	Limit                int // this is currently not used. Offset for fetching data is pageSize
}

func (ctx *TableWidget) prevPage() {
	page, err := ctx.currentPage.Get()
	if err != nil {
		fmt.Println(err)
		return
	}
	page -= 1
	if page < 1 {
		return
	}
	ctx.currentPage.Set(page)
	ctx.Refresh()
}

func (ctx *TableWidget) nextPage() {
	page, err := ctx.currentPage.Get()
	if err != nil {
		fmt.Println(err)
		return
	}
	page += 1
	max, err := ctx.totalPages.Get()
	if err != nil || page > max {
		return
	}
	ctx.currentPage.Set(page)
	ctx.Refresh()
}

func (ctx *TableWidget) firstPage() {
	ctx.currentPage.Set(1)
	ctx.Refresh()
}

func (ctx *TableWidget) lastPage() {
	max, err := ctx.totalPages.Get()
	if err != nil {
		return
	}
	ctx.currentPage.Set(max)
	ctx.Refresh()
}

func NewTableWidget(title string, pageSize int) *TableWidget {
	ctx := &TableWidget{
		currentPage:      binding.NewInt(),
		totalPages:       binding.NewInt(),
		totalResults:     binding.NewInt(),
		currentObjectUid: binding.NewString(),
		Title:            title,
		pageSize:         pageSize,
	}

	ctx.currentPage.Set(1)
	ctx.table = NewFlexTable(NewTableData("empty"), ctx.cellOnClick)
	ctx.table.OnClickSecondary = ctx.cellOnSecondaryClick

	currentPage := widget.NewEntryWithData(binding.IntToString(ctx.currentPage))
	currentPage.Validator = nil
	currentPage.OnSubmitted = func(s string) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		ctx.currentPage.Set(i)
		ctx.Refresh()
	}
	totalPages := widget.NewLabelWithData(binding.IntToString(ctx.totalPages))
	first := theme.NewThemedResource(fyne.NewStaticResource("navigateFirst", navigateFirst))
	last := theme.NewThemedResource(fyne.NewStaticResource("navigateLast", navigateLast))

	leftFooter := container.NewHBox(
		widget.NewButtonWithIcon("", first, ctx.firstPage),
		widget.NewButtonWithIcon("", theme.NavigateBackIcon(), ctx.prevPage),
		currentPage,
		widget.NewLabel("/"),
		totalPages,
		widget.NewButtonWithIcon("", theme.NavigateNextIcon(), ctx.nextPage),
		widget.NewButtonWithIcon("", last, ctx.lastPage),
		widget.NewLabel("Count:"),
		widget.NewLabelWithData(binding.IntToString(ctx.totalResults)),
	)
	footer := container.NewBorder(ctx.newfilterWidget(), nil, leftFooter, widget.NewButton("Export to Excel", ctx.showExportWindow))

	ctx.table.SelectionColor = color.RGBA{85, 85, 85, 128}
	ctx.Instance = container.NewBorder(nil, footer, nil, nil, container.NewPadded(ctx.table))
	ctx.Data = func(offset, limit int) *TableData { return NewTableData("empty") }
	ctx.RowCount = func() int { return 0 }
	ctx.Refresh()

	return ctx
}

func moveItemToFront(slice []string, target string) []string {
	for i, item := range slice {
		if item == target {
			// Swap the found item with the first element
			slice[0], slice[i] = slice[i], slice[0]
			break
		}
	}
	return slice
}

func (ctx *TableWidget) Refresh() {
	page, err := ctx.currentPage.Get()
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx.Offset = (page - 1) * ctx.pageSize
	ctx.Limit = ctx.pageSize // ctx.Offset + ctx.pageSize
	data := ctx.Data(ctx.Offset, ctx.Limit)
	totalCount := ctx.RowCount()

	ctx.totalResults.Set(totalCount)
	ctx.totalPages.Set((totalCount + ctx.pageSize - 1) / ctx.pageSize)

	ctx.currentData = data
	if ctx.currentFilter != "" {
		ctx.filter(ctx.currentFilter)
		ctx.table.SetData(ctx.filteredData)
	} else {
		ctx.table.SetData(data)
	}
	//ctx.table.SetColumnWidth(0, 150) // Why was this here?
	ctx.table.Refresh()
}

func (ctx *TableWidget) SetColumnWidth(id int, width float32) {
	ctx.table.SetColumnWidth(id, width)
}

func (ctx *TableWidget) SetOnClick(onClick func(row, col int)) {
	ctx.table.OnClick = func(c *TableCell) {
		onClick(c.Id.Row, c.Id.Col)
	}
}

func (ctx *TableWidget) cellOnClick(c *TableCell) {
	ctx.currentObjectUid.Set(ctx.currentData.Get(0, c.Id.Row))
}

func (ctx *TableWidget) cellOnSecondaryClick(c *TableCell) {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Copy value to clipboard", func() {
			clipboard.Write(clipboard.FmtText, []byte(c.Text()))
		}),
		fyne.NewMenuItem("Copy row to clipboard", func() {
			row := make([]string, ctx.currentData.ColumnCount())
			for j := 0; j < ctx.currentData.ColumnCount(); j++ {
				val := ctx.currentData.Get(j, c.Id.Row)
				row[j] = val
			}
			clipboard.Write(clipboard.FmtText, []byte(strings.Join(row, "\t")))
		}),
	}
	canvas := fyne.CurrentApp().Driver().CanvasForObject(c)
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu(c.Text(), items...), canvas, pos.AddXY(c.Size().Width/2, 0))
}

func (ctx *TableWidget) filter(text string) {
	ctx.filteredData = NewTableData("filtered data")
	// for key, val := range currentData.ColumnIdToField {
	// 	filteredData.ColumnIdToField[key] = val
	// }
	for i := 0; i < ctx.currentData.RowCount(); i++ {
		stringRow := make([]string, ctx.currentData.ColumnCount())
		found := false
		for j := 0; j < ctx.currentData.ColumnCount(); j++ {
			val := ctx.currentData.Get(j, i)
			if !found && strings.Contains(strings.ToLower(val), strings.ToLower(text)) {
				found = true
			}
			stringRow[j] = val
		}
		if found {
			ctx.filteredData.AddStringRow(ctx.currentData.Columns, stringRow)
		}
	}
	if len(ctx.filteredData.Columns) > 1 {
		cols := ctx.filteredData.Columns[1:]
		// sort.Slice(cols, func(i, j int) bool {
		// 	a := filteredData.ColumnIdToField[cols[i]]
		// 	b := filteredData.ColumnIdToField[cols[j]]
		// 	return a.Order < b.Order
		// })
		ctx.filteredData.Columns = append(ctx.filteredData.Columns[0:1], cols...)
		ctx.filteredData.Sort("field-uid", true)
	}
}

func (ctx *TableWidget) newfilterWidget() *widget.Entry {
	filter := widget.NewEntry()
	filter.PlaceHolder = "Filter in page"
	filter.OnChanged = func(text string) {
		if text == "" {
			ctx.filteredData = ctx.currentData
			ctx.currentFilter = ""
		} else {
			ctx.currentFilter = text
			ctx.filter(text)
		}
		ctx.table.SetData(ctx.filteredData)
		ctx.table.SetColumnWidth(0, 150)
		ctx.table.Refresh()
	}

	return filter
}

func (ctx *TableWidget) showExportWindow() {
	w := fyne.CurrentApp().NewWindow("Export")

	selected := make([]string, ctx.currentData.ColumnCount())
	copy(selected, ctx.currentData.Columns)

	columnsSelection := widget.NewCheckGroup(selected, func(sel []string) {
		selected = sel
	})
	columnsSelection.SetSelected(selected)

	dataset := widget.NewRadioGroup([]string{"Current page", "All data"}, func(selection string) {

	})
	dataset.SetSelected("Current page")

	datetime := strconv.Itoa(time.Now().Year()) +
		"-" + fmt.Sprintf("%02d", time.Now().Month()) +
		"-" + fmt.Sprintf("%02d", time.Now().Day()) +
		"-" + fmt.Sprintf("%02d", time.Now().Hour()) +
		"-" + fmt.Sprintf("%02d", time.Now().Minute()) +
		"-" + fmt.Sprintf("%02d", time.Now().Second())

	exportPath := widget.NewEntry()
	exportPath.Text = "C:/goto/export"
	filename := widget.NewEntry()
	filename.Text = ctx.Title + "-" + datetime + ".xlsx"

	h := widget.NewForm(
		widget.NewFormItem("Export", dataset),
		widget.NewFormItem("Selected columns", columnsSelection),
		widget.NewFormItem("Export path", exportPath),
		widget.NewFormItem("Filename", filename),
	)

	h.CancelText = "Close"
	h.SubmitText = "Export"
	h.OnCancel = func() {
		w.Close()
	}

	h.OnSubmit = func() {
		var data *TableData
		if dataset.Selected == "Current page" {
			data = ctx.Data(ctx.Offset, ctx.pageSize)
		} else {
			data = ctx.Data(0, ctx.pageSize)
		}

		os.MkdirAll(exportPath.Text, 0755)

		err := ctx.ExportToExcel(data, selected, filepath.Join(exportPath.Text, filename.Text))

		if err != nil {
			dialog.ShowError(err, w)
		} else {
			w.Close()
		}
	}
	w.SetContent(h)
	w.Resize(fyne.NewSize(800, 600))
	w.Show()
}

func (ctx *TableWidget) ExportToExcel(data *TableData, columns []string, path string) error {
	excel := excelize.NewFile()
	sheet := data.TableName
	excel.NewSheet(sheet)

	// Write headers
	for col, x := range columns {
		cellName, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			fmt.Println("Error getting cellname:", err)
		} else {

		}
		excel.SetCellValue(sheet, cellName, x)
	}

	// Write headers
	for row := 0; row < data.RowCount(); row++ {
		for col, x := range columns {
			cellName, err := excelize.CoordinatesToCellName(col+1, row+2)
			if err != nil {
				fmt.Println("Error getting cellname:", err)
			} else {
				excel.SetCellValue(sheet, cellName, data.GetFromColumn(x, row))
			}
		}
	}
	excel.DeleteSheet("Sheet1")
	excel.SetActiveSheet(0)
	return excel.SaveAs(path)
}
