package tablewidget

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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

func (ctx *TableWidget) GetFlexTable() *FlexTable {
	return ctx.table
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
	footer := container.NewBorder(ctx.newfilterWidget(), nil, leftFooter, widget.NewButton("Export", ctx.showExportWindow))

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
			writeClipboard(c.Text())
		}),
		fyne.NewMenuItem("Copy row to clipboard", func() {
			row := make([]string, ctx.currentData.ColumnCount())
			for j := 0; j < ctx.currentData.ColumnCount(); j++ {
				val := ctx.currentData.Get(j, c.Id.Row)
				row[j] = val
			}
			writeClipboard(strings.Join(row, "\t"))
		}),
	}
	canvas := fyne.CurrentApp().Driver().CanvasForObject(c)
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu(c.Text(), items...), canvas, pos.AddXY(c.Size().Width/2, 0))
}

func (ctx *TableWidget) filter(text string) {
	ctx.filteredData = NewTableData("filtered data")
	ctx.filteredData.RowMapping = []int{} // Initialize row mapping

	for i := 0; i < ctx.currentData.RowCount(); i++ {
		stringRow := make([]string, ctx.currentData.ColumnCount())
		found := false
		for j := 0; j < ctx.currentData.ColumnCount(); j++ {
			val := ctx.currentData.Get(j, i)
			// Skip filtering on widget placeholder columns (marked with [Widget Type])
			isWidgetColumn := strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]")
			if !found && !isWidgetColumn && strings.Contains(strings.ToLower(val), strings.ToLower(text)) {
				found = true
			}
			stringRow[j] = val
		}
		if found {
			ctx.filteredData.AddStringRow(ctx.currentData.Columns, stringRow)
			// Store mapping: filtered row index -> original row index
			ctx.filteredData.RowMapping = append(ctx.filteredData.RowMapping, i)
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

	// Get default export directory (current working directory + /exports)
	defaultExportDir := "./exports"
	if cwd, err := os.Getwd(); err == nil {
		defaultExportDir = filepath.Join(cwd, "exports")
	}

	exportPath := widget.NewEntry()
	exportPath.Text = defaultExportDir
	filename := widget.NewEntry()
	filename.Text = ctx.Title + "-" + datetime + ".csv"

	format := widget.NewRadioGroup([]string{"CSV", "XLSX", "JSON"}, func(selection string) {
		// Update filename extension when format changes
		name := filename.Text
		ext := filepath.Ext(name)
		if ext != "" {
			name = name[:len(name)-len(ext)]
		}
		switch selection {
		case "CSV":
			filename.SetText(name + ".csv")
		case "XLSX":
			filename.SetText(name + ".xlsx")
		case "JSON":
			filename.SetText(name + ".json")
		}
	})
	format.SetSelected("CSV")

	h := widget.NewForm(
		widget.NewFormItem("Export", dataset),
		widget.NewFormItem("Format", format),
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

		var err error
		fullPath := filepath.Join(exportPath.Text, filename.Text)

		switch format.Selected {
		case "CSV":
			err = ctx.ExportToCSV(data, selected, fullPath)
		case "XLSX":
			err = ctx.ExportToExcel(data, selected, fullPath)
		case "JSON":
			err = ctx.ExportToJSON(data, selected, fullPath)
		default:
			err = fmt.Errorf("unsupported format: %s", format.Selected)
		}

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

func (ctx *TableWidget) getCellValue(data *TableData, columnName string, row int) string {
	// For both string and widget modes, use the Data callback
	// In widget mode, users should populate the Data callback with actual string values
	// for export functionality to work properly
	return data.GetFromColumn(columnName, row)
}

func (ctx *TableWidget) extractWidgetText(obj fyne.CanvasObject) string {
	// Unwrap WidgetCell to get actual content
	if widgetCell, ok := obj.(*WidgetCell); ok {
		if widgetCell.content != nil {
			return ctx.extractWidgetText(widgetCell.content)
		}
		return ""
	}

	switch v := obj.(type) {
	case *widget.Label:
		return v.Text
	case *widget.Button:
		return "[Button: " + v.Text + "]"
	case *widget.Entry:
		return v.Text
	case *widget.Check:
		if v.Checked {
			return "☑"
		}
		return "☐"
	case *widget.Select:
		return v.Selected
	case *widget.Icon:
		if v.Resource != nil {
			return "[Icon: " + v.Resource.Name() + "]"
		}
		return "[Icon]"
	case *fyne.Container:
		// For images in containers
		if len(v.Objects) > 0 {
			return ctx.extractWidgetText(v.Objects[0])
		}
		return "[Container]"
	case *canvas.Image:
		if v.File != "" {
			return "[Image: " + v.File + "]"
		}
		return "[Image]"
	default:
		return fmt.Sprintf("[%T]", obj)
	}
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

	// Write data rows
	for row := 0; row < data.RowCount(); row++ {
		for col, x := range columns {
			cellName, err := excelize.CoordinatesToCellName(col+1, row+2)
			if err != nil {
				fmt.Println("Error getting cellname:", err)
			} else {
				excel.SetCellValue(sheet, cellName, ctx.getCellValue(data, x, row))
			}
		}
	}
	excel.DeleteSheet("Sheet1")
	excel.SetActiveSheet(0)
	return excel.SaveAs(path)
}

func (ctx *TableWidget) ExportToCSV(data *TableData, columns []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(columns); err != nil {
		return err
	}

	// Write data rows
	for row := 0; row < data.RowCount(); row++ {
		record := make([]string, len(columns))
		for col, columnName := range columns {
			record[col] = ctx.getCellValue(data, columnName, row)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *TableWidget) ExportToJSON(data *TableData, columns []string, path string) error {
	// Build array of objects
	var records []map[string]string
	for row := 0; row < data.RowCount(); row++ {
		record := make(map[string]string)
		for _, columnName := range columns {
			record[columnName] = ctx.getCellValue(data, columnName, row)
		}
		records = append(records, record)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, jsonData, 0644)
}
