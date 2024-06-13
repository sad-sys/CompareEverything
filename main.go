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
	subject string
	grade   string
}

func findGrade(record []string, grade string) float64 {
	grades := [7]string{"A*", "A", "B", "C", "D", "E", "U"}

	counter := 0.0

	for j := 0; j < len(grades); j++ {
		if grade == grades[j] {
			percentile, err := strconv.ParseFloat(record[j+3], 64)
			if err != nil {
				fmt.Println(err)
			}
			counter = counter + percentile
		}
	}

	return counter
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
	} else {
		reader := csv.NewReader(file)
		records, errReader := reader.ReadAll()
		if errReader != nil {
			fmt.Println("Error reader", errReader)
		} else {
			for _, record := range records {
				fmt.Println(record)
				contains := searchRecords(subject, record)
				if contains {
					counter := findGrade(record, grade)
					fmt.Println(counter)
					return counter, record
				}
			}
		}
	}
	return -1.0, []string{"7127", "ACCOUNTING ADV", "2314", "1.8", "13.8", "34.8", "58.9", "79.2", "93.7", "100.0"}
}

func main() {

	counter := 0.0
	record := []string{"7127", "ACCOUNTING ADV", "2314", "1.8", "13.8", "34.8", "58.9", "79.2", "93.7", "100.0"}
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if err := r.ParseForm(); err != nil {
			fmt.Println(err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		details := gradeInfo{
			subject: r.FormValue("subject"),
			grade:   r.FormValue("grade"),
		}
		counter, record = convertToCsv("results.csv", details.subject, details.grade)
		fmt.Println("===============================")
		fmt.Println(details)

		tmpl, err := template.ParseFiles("templates/index.html")
		tmpl.Execute(w, nil)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(counter, record)

	})
	fmt.Println("Server Running at LocalHost")
	http.ListenAndServe(":8080", nil)

}
