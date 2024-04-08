package main

import (
	"database/sql"
	"embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
	"io/ioutil"
	"log"
	_ "os"
)

//go:embed jobinfo.db
var dbAsset embed.FS

func initDB() *sql.DB {
	tempFile, err := ioutil.TempFile("", "jobinfo-*.db")
	if err != nil {
		log.Fatal(err)
	}
	defer tempFile.Close()

	dbBytes, err := dbAsset.ReadFile("jobinfo.db")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tempFile.Write(dbBytes); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", tempFile.Name())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS jobs (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        CompanyName TEXT,
        PostingAge TEXT,
        JobId TEXT UNIQUE,
        Country TEXT,
        Location TEXT,
        PublicationDate TEXT,
        SalaryMax TEXT,
        SalaryMin TEXT,
        SalaryType TEXT,
        JobTitle TEXT
    );`
	if _, err := db.Exec(sqlStmt); err != nil {
		log.Fatalf("Error creating jobs table: %v", err)
	}

	return db
}

func loadExcelDataOnce(db *sql.DB, filePath string) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Failed to open Excel file: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Comp490 Jobs")
	if err != nil {
		log.Fatalf("Failed to get rows from Excel file: %v", err)
	}

	for i, row := range rows {
		if i == 0 { // Skip header row
			continue
		}
		_, err = db.Exec(`INSERT OR IGNORE INTO jobs (CompanyName, PostingAge, JobId, Country, Location, PublicationDate, SalaryMax, SalaryMin, SalaryType, JobTitle) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[7], row[8], row[9])

		if err != nil {
			log.Printf("Failed to insert record into database: %v\n", err)
		}
	}
	log.Println("Excel data loaded into database successfully.")
}

func main() {
	db := initDB()
	defer db.Close()

	loadExcelDataOnce(db, "Project3Data.xlsx")

	myApp := app.NewWithID("fyne.demo")
	myWindow := myApp.NewWindow("Job Database Manager")
	myWindow.Resize(fyne.NewSize(800, 600))

	var selectedIndex int
	//var globalData []map[string]string

	list := widget.NewList(
		func() int {
			count, err := countRecords(db)
			if err != nil {
				log.Println("Failed to count records:", err)
				return 0
			}
			return count
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			record, err := getRecordByIndex(db, id)
			if err != nil {
				log.Println("Failed to get record by index:", err)
				co.(*widget.Label).SetText("Error loading record")
				return
			}
			co.(*widget.Label).SetText(fmt.Sprintf("%s - %s", record["CompanyName"], record["JobTitle"]))
		},
	)

	selectedRecord := widget.NewLabel("Select a record")
	selectedRecord.Wrapping = fyne.TextWrapWord

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		record, err := getRecordByIndex(db, id)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		selectedRecord.SetText(formatRecordDetails(record))
	}

	addButton := widget.NewButton("Add New Record", func() {
		addFormWindow := myApp.NewWindow("Add New Job")
		addForm := createAddForm(func(data map[string]string) {
			err := addRecord(db, data)
			if err != nil {
				dialog.ShowError(err, myWindow)
			} else {
				list.Refresh()
				addFormWindow.Close()
			}
		})
		addFormWindow.SetContent(addForm)
		addFormWindow.Show()
	})

	deleteButton := widget.NewButton("Delete Selected Record", func() {
		record, err := getRecordByIndex(db, selectedIndex)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		err = deleteRecord(db, record["JobId"])
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		list.Refresh()
		selectedRecord.SetText("Select a record")
	})

	updateButton := widget.NewButton("Update Selected Record", func() {
		record, err := getRecordByIndex(db, selectedIndex)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		updateFormWindow := myApp.NewWindow("Update Job")
		updateForm := createUpdateForm(record, func(updatedData map[string]string) {
			err := updateRecord(db, updatedData)
			if err != nil {
				dialog.ShowError(err, myWindow)
			} else {
				list.Refresh()
				updateFormWindow.Close()
				selectedRecord.SetText(formatRecordDetails(updatedData))
			}
		})
		updateFormWindow.SetContent(updateForm)
		updateFormWindow.Show()
	})

	buttons := container.NewVBox(addButton, deleteButton, updateButton)
	rightSide := container.NewBorder(nil, buttons, nil, nil, selectedRecord)
	content := container.NewHSplit(list, rightSide)
	content.Offset = 0.3

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func countRecords(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&count)
	return count, err
}

func getRecordByIndex(db *sql.DB, index int) (map[string]string, error) {
	query := `SELECT CompanyName, PostingAge, JobId, Country, Location, PublicationDate, SalaryMax, SalaryMin, SalaryType, JobTitle FROM jobs ORDER BY id LIMIT 1 OFFSET ?`
	row := db.QueryRow(query, index)

	// Temporary variables to scan the values into
	var companyName, postingAge, jobId, country, location, publicationDate, salaryMax, salaryMin, salaryType, jobTitle string

	err := row.Scan(&companyName, &postingAge, &jobId, &country, &location, &publicationDate, &salaryMax, &salaryMin, &salaryType, &jobTitle)
	if err != nil {
		return nil, err
	}

	// Assigning values to the map after successfully scanning
	record := map[string]string{
		"CompanyName":     companyName,
		"PostingAge":      postingAge,
		"JobId":           jobId,
		"Country":         country,
		"Location":        location,
		"PublicationDate": publicationDate,
		"SalaryMax":       salaryMax,
		"SalaryMin":       salaryMin,
		"SalaryType":      salaryType,
		"JobTitle":        jobTitle,
	}
	return record, nil
}

func formatRecordDetails(data map[string]string) string {
	details := "Selected Record:\n\n"
	for key, value := range data {
		details += fmt.Sprintf("%s: %s\n", key, value)
	}
	return details
}

func addRecord(db *sql.DB, data map[string]string) error {
	_, err := db.Exec(`INSERT INTO jobs (CompanyName, PostingAge, JobId, Country, Location, PublicationDate, SalaryMax, SalaryMin, SalaryType, JobTitle) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		data["CompanyName"], data["PostingAge"], data["JobId"], data["Country"], data["Location"], data["PublicationDate"], data["SalaryMax"], data["SalaryMin"], data["SalaryType"], data["JobTitle"])
	return err
}

func updateRecord(db *sql.DB, data map[string]string) error {
	_, err := db.Exec(`UPDATE jobs SET CompanyName = ?, PostingAge = ?, Country = ?, Location = ?, PublicationDate = ?, SalaryMax = ?, SalaryMin = ?, SalaryType = ?, JobTitle = ? WHERE JobId = ?`,
		data["CompanyName"], data["PostingAge"], data["Country"], data["Location"], data["PublicationDate"], data["SalaryMax"], data["SalaryMin"], data["SalaryType"], data["JobTitle"], data["JobId"])
	return err
}

func deleteRecord(db *sql.DB, jobId string) error {
	_, err := db.Exec(`DELETE FROM jobs WHERE JobId = ?`, jobId)
	return err
}

func createAddForm(onSubmit func(map[string]string)) *widget.Form {
	entries := make(map[string]*widget.Entry)
	items := []*widget.FormItem{}
	headers := []string{"CompanyName", "PostingAge", "JobId", "Country", "Location", "PublicationDate", "SalaryMax", "SalaryMin", "SalaryType", "JobTitle"}

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
				entry.SetText("") // Clear the form
			}
			onSubmit(data)
		},
	}
}

func createUpdateForm(currentData map[string]string, onSubmit func(map[string]string)) *widget.Form {
	entries := make(map[string]*widget.Entry)
	items := []*widget.FormItem{}
	for key, value := range currentData {
		entry := widget.NewEntry()
		entry.SetText(value)
		entries[key] = entry
		items = append(items, widget.NewFormItem(key, entry))
	}

	return &widget.Form{
		Items: items,
		OnSubmit: func() {
			data := map[string]string{}
			for key, entry := range entries {
				data[key] = entry.Text
			}
			onSubmit(data)
		},
	}
}
