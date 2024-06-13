package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type gradeInfo struct {
	Subject string
	Grade   string
	Show    bool
	Percent float64
}

func findGrade(record []string, grade string) float64 {
	grades := [7]string{"A*", "A", "B", "C", "D", "E", "U"}

	for j := 0; j < len(grades); j++ {
		if grade == grades[j] {
			percentile, err := strconv.ParseFloat(record[j+3], 64)
			if err != nil {
				fmt.Println(err)
			}
			return percentile
		}
	}
	return 0.0
}

func searchRecords(word string, record []string) bool {
	for _, field := range record {
		if strings.Contains(field, word) {
			return true
		}
	}
	return false
}

func convertToCsv(fileName string, subject string, grade string) (float64, []string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error", err)
		return -1.0, nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reader", err)
		return -1.0, nil
	}

	for _, record := range records {
		if searchRecords(subject, record) {
			counter := findGrade(record, grade)
			return counter, record
		}
	}

	return -1.0, nil
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		details := gradeInfo{
			Show: false,
		}

		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				fmt.Println(err)
				http.Error(w, "Unable to parse form", http.StatusBadRequest)
				return
			}

			details.Subject = r.FormValue("subject")
			details.Grade = r.FormValue("grade")

			details.Percent, _ = convertToCsv("results.csv", details.Subject, details.Grade)
			if details.Percent > 0 {
				details.Show = true
			}
		}

		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Unable to load template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, details)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Unable to render template", http.StatusInternalServerError)
		}
	})

	fmt.Println("Server Running at LocalHost")
	http.ListenAndServe(":8080", nil)
}
