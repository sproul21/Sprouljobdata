package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/xuri/excelize/v2"
	"log"
)

var (
	globalData    []map[string]string
	selectedIndex int
)

func main() {
	loadExcelDataOnce("Project2Data.xlsx")

	myApp := app.NewWithID("fyne.demo")
	myWindow := myApp.NewWindow("Excel Data Manager")
	myWindow.Resize(fyne.NewSize(800, 600))

	headers := []string{"Company Name", "Posting Age", "Job Id", "Country", "Location", "Publication Date", "Salary Max", "Salary Min", "Salary Type", "Job Title"}

	selectedRecord := widget.NewLabel("")
	selectedRecord.Wrapping = fyne.TextWrapWord

	list := widget.NewList(
		func() int {
			return len(globalData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(globalData[id]["Company Name"])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		selectedRecord.SetText(formatRecordDetails(globalData[id], headers))
	}

	addForm := createAddForm(headers, func(data map[string]string) {
		globalData = append(globalData, data)
		err := writeExcelFile("Project2Data.xlsx", globalData, headers)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		list.Refresh()
	})

	deleteButton := widget.NewButton("Delete Selected Record", func() {
		if selectedIndex >= 0 && selectedIndex < len(globalData) {
			globalData = append(globalData[:selectedIndex], globalData[selectedIndex+1:]...)
			err := writeExcelFile("Project2Data.xlsx", globalData, headers)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			selectedIndex = -1
			selectedRecord.SetText("Select a record")
			list.Refresh()
		}
	})

	updateButton := widget.NewButton("Update Selected Record", func() {
		if selectedIndex < 0 || selectedIndex >= len(globalData) {
			dialog.ShowInformation("No Selection", "Please select a record to update.", myWindow)
			return
		}

		selectedData := globalData[selectedIndex]

		updateForm := createUpdateForm(headers, selectedData, func(updatedData map[string]string) {

			globalData[selectedIndex] = updatedData

			err := writeExcelFile("Project2Data.xlsx", globalData, headers)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			list.Refresh()

			selectedRecord.SetText(formatRecordDetails(updatedData, headers))
		})

		dialog.ShowForm("Update Record", "Update", "Cancel", updateForm.Items, func(b bool) {}, myWindow)
	})

	buttons := container.NewVBox(
		widget.NewButton("Add New Record", func() {
			dialog.ShowCustomConfirm("Add New Record", "Add", "Cancel", addForm, func(b bool) {}, myWindow)
		}),
		deleteButton,
		updateButton,
	)

	content := container.NewHSplit(
		container.NewBorder(selectedRecord, buttons, nil, nil),
		list,
	)
	content.Offset = 0.3

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func loadExcelDataOnce(filePath string) {
	var err error
	globalData, err = readExcelFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load Excel data: %v", err)
	}
}

func readExcelFile(filePath string) ([]map[string]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Comp490 Jobs")
	if err != nil {
		return nil, err
	}

	var data []map[string]string
	for _, row := range rows[1:] {
		record := make(map[string]string)
		for j, cell := range row {
			record[rows[0][j]] = cell
		}
		data = append(data, record)
	}
	return data, nil
}

func writeExcelFile(filePath string, data []map[string]string, headers []string) error {
	f := excelize.NewFile()
	sheetName := "Comp490 Jobs"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	for i, record := range data {
		for j, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, record[header])
		}
	}

	return f.SaveAs(filePath)
}

func createAddForm(headers []string, onSubmit func(map[string]string)) *widget.Form {
	entries := make(map[string]*widget.Entry)
	items := []*widget.FormItem{}
	for _, header := range headers {
		entry := widget.NewEntry()
		entries[header] = entry
		items = append(items, widget.NewFormItem(header, entry))
	}
	return &widget.Form{
		Items: items,
		OnSubmit: func() {
			data := map[string]string{}
			for header, entry := range entries {
				data[header] = entry.Text
				entry.SetText("")
			}
			onSubmit(data)
		},
	}
}

func createUpdateForm(headers []string, currentData map[string]string, onUpdate func(map[string]string)) *widget.Form {
	entries := make(map[string]*widget.Entry)
	var items []*widget.FormItem

	for _, header := range headers {
		entry := widget.NewEntry()
		entry.SetText(currentData[header])
		entries[header] = entry
		items = append(items, widget.NewFormItem(header, entry))
	}

	return &widget.Form{
		Items: items,
		OnSubmit: func() {
			updatedData := make(map[string]string)
			for header, entry := range entries {
				updatedData[header] = entry.Text
			}
			onUpdate(updatedData)
		},
	}
}

func formatRecordDetails(data map[string]string, headers []string) string {
	details := "Selected Record:\n\n"
	for _, header := range headers {
		details += fmt.Sprintf("%s: %s\n", header, data[header])
	}
	return details
}
