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
	"os"
)

func readExcelFile(filePath string) ([]map[string]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	rows, err := f.GetRows("Comp490 Jobs")
	if err != nil {
		log.Fatal(err)
	}

	headers := rows[0]
	var data []map[string]string

	for _, row := range rows[1:] {
		record := make(map[string]string)
		for j, cell := range row {
			if j < len(headers) {
				record[headers[j]] = cell
			}
		}
		data = append(data, record)
	}

	return data, nil
}

func writeExcelFile(filePath string, data []map[string]string, headers []string) error {
	f := excelize.NewFile()
	sheetName := "Comp490 Jobs"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal(err)
	}
	f.SetActiveSheet(index)

	for colIndex, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data rows
	for i, record := range data {
		for colIndex, header := range headers {
			value := record[header]
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, i+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	if err := f.SaveAs(filePath); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current working directory:", wd)

	myApp := app.New()
	myWindow := myApp.NewWindow("Excel Data Manager")
	myWindow.Resize(fyne.NewSize(1200, 800))

	selectedIndex := -1

	headers := []string{"Company Name", "Posting Age", "Job Id", "Country", "Location", "Publication Date", "Salary Max", "Salary Min", "Salary Type", "Job Title"}

	list := widget.NewList(
		func() int {
			data, _ := readExcelFile("Project2Data.xlsx")

			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			data, _ := readExcelFile("Project2Data.xlsx")

			if id < len(data) {
				co.(*widget.Label).SetText(data[id]["Company Name"])
			}
		},
	)

	selectedLabel := widget.NewLabel("Select a record")
	selectedLabel.Wrapping = fyne.TextWrapWord

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
		data, _ := readExcelFile("Project2Data.xlsx")

		if id < len(data) {

			selectedRecordText := "Selected Record:\n"

			fieldOrder := []string{
				"Company Name",
				"Posting Age",
				"Job Title",
				"Country",
				"Location",
				"Publication Date",
				"Min Salary",
				"Max Salary",
				"Salary Type",
				"Job ID",
			}

			for _, field := range fieldOrder {
				value := data[id][field]
				selectedRecordText += fmt.Sprintf("%s: %s\n", field, value)
			}

			selectedLabel.SetText(selectedRecordText)
		}
	}

	companyNameEntry := widget.NewEntry()
	postingAgeEntry := widget.NewEntry()
	jobIDEntry := widget.NewEntry()
	countryEntry := widget.NewEntry()
	locationEntry := widget.NewEntry()
	publicationDateEntry := widget.NewEntry()
	salaryMaxEntry := widget.NewEntry()
	salaryMinEntry := widget.NewEntry()
	salaryTypeEntry := widget.NewEntry()
	jobTitleEntry := widget.NewEntry()

	addForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Company Name", Widget: companyNameEntry},
			{Text: "Posting Age", Widget: postingAgeEntry},
			{Text: "Job ID", Widget: jobIDEntry},
			{Text: "Country", Widget: countryEntry},
			{Text: "Location", Widget: locationEntry},
			{Text: "Publication Date", Widget: publicationDateEntry},
			{Text: "Min Salary", Widget: salaryMinEntry},
			{Text: "Max Salary", Widget: salaryMaxEntry},
			{Text: "Salary Type", Widget: salaryTypeEntry},
			{Text: "Job Title", Widget: jobTitleEntry},
		},
		OnSubmit: func() {

			newRecord := map[string]string{
				"Company Name":     companyNameEntry.Text,
				"Posting Age":      postingAgeEntry.Text,
				"Job ID":           jobIDEntry.Text,
				"Country":          countryEntry.Text,
				"Location":         locationEntry.Text,
				"Publication Date": publicationDateEntry.Text,
				"Min Salary":       salaryMinEntry.Text,
				"Max Salary":       salaryMaxEntry.Text,
				"Salary Type":      salaryTypeEntry.Text,
				"Job Title":        jobTitleEntry.Text,
			}
			data, _ := readExcelFile("Project2Data.xlsx")

			data = append(data, newRecord)
			err := writeExcelFile("Project2Data.xlsx", data, headers)
			if err != nil {
				log.Fatal(err) // Properly handle the error
			}
			companyNameEntry.SetText("")
			postingAgeEntry.SetText("")
			jobIDEntry.SetText("")
			countryEntry.SetText("")
			locationEntry.SetText("")
			publicationDateEntry.SetText("")
			salaryMinEntry.SetText("")
			salaryMaxEntry.SetText("")
			salaryTypeEntry.SetText("")
			jobTitleEntry.SetText("")
			list.Refresh()
		},
	}

	// Delete record button
	deleteButton := widget.NewButton("Delete Selected Record", func() {
		data, _ := readExcelFile("Project2Data.xlsx")

		selectedIndex := selectedIndex
		if selectedIndex >= 0 && selectedIndex < len(data) {
			// Remove the selected index
			data = append(data[:selectedIndex], data[selectedIndex+1:]...)
			err := writeExcelFile("Project2Data.xlsx", data, headers)
			if err != nil {
				log.Fatal(err)
			}
			list.UnselectAll()
			selectedLabel.SetText("Select a record")
			list.Refresh()
		}
	})

	updateForm := widget.NewForm(
		widget.NewFormItem("Company Name", companyNameEntry),
		widget.NewFormItem("Posting Age", postingAgeEntry),
		widget.NewFormItem("Job ID", jobIDEntry),
		widget.NewFormItem("Country", countryEntry),
		widget.NewFormItem("Location", locationEntry),
		widget.NewFormItem("Publication Date", publicationDateEntry),
		widget.NewFormItem("Min Salary", salaryMinEntry),
		widget.NewFormItem("Max Salary", salaryMaxEntry),
		widget.NewFormItem("Salary Type", salaryTypeEntry),
		widget.NewFormItem("Job Title", jobTitleEntry),
	)
	updateForm.SubmitText = "Update Record"
	updateForm.OnSubmit = func() {
		selectedIndex := selectedIndex
		if selectedIndex >= 0 {
			data, _ := readExcelFile("Project2Data.xlsx")

			if selectedIndex < len(data) {
				data[selectedIndex]["Company Name"] = companyNameEntry.Text
				data[selectedIndex]["Posting Age"] = postingAgeEntry.Text
				data[selectedIndex]["Job ID"] = jobIDEntry.Text
				data[selectedIndex]["Country"] = countryEntry.Text
				data[selectedIndex]["Location"] = locationEntry.Text
				data[selectedIndex]["Publication Date"] = publicationDateEntry.Text
				data[selectedIndex]["Min Salary"] = salaryMinEntry.Text
				data[selectedIndex]["Max Salary"] = salaryMaxEntry.Text
				data[selectedIndex]["Salary Type"] = salaryTypeEntry.Text
				data[selectedIndex]["Job Title"] = jobTitleEntry.Text
				err := writeExcelFile("Project2Data.xlsx", data, headers)
				if err != nil {
					log.Fatal(err)
				}
				list.Refresh()
				companyNameEntry.SetText("")
				postingAgeEntry.SetText("")
				jobIDEntry.SetText("")
				countryEntry.SetText("")
				locationEntry.SetText("")
				publicationDateEntry.SetText("")
				salaryMinEntry.SetText("")
				salaryMaxEntry.SetText("")
				salaryTypeEntry.SetText("")
				jobTitleEntry.SetText("")
			}
		}
	}

	buttonLayout := container.NewVBox(
		widget.NewButton("Add New Record", func() {
			dialog.ShowCustomConfirm("Add New Record", "Add", "Cancel", addForm, func(b bool) {}, myWindow)
		}),
		deleteButton,
		widget.NewButton("Update Selected Record", func() {
			dialog.ShowCustomConfirm("Update Record", "Update", "Cancel", updateForm, func(b bool) {}, myWindow)
		}),
	)

	leftPane := container.NewBorder(nil, buttonLayout, nil, nil, selectedLabel)

	rightPane := list

	split := container.NewHSplit(leftPane, rightPane)
	split.Offset = 0.25

	myWindow.SetContent(split)
	myWindow.ShowAndRun()
}
