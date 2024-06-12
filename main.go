package main

import (
	"encoding/csv"
	"fmt"

	//"html/template"
	//"net/http"
	"os"
	"strconv"
	"strings"
)

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

func convertToCsv(fileName string) [][]string {
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
				contains := searchRecords("CHEMISTRY", record)
				fmt.Println(contains)
				if contains {
					counter := findGrade(record, "A*")
					fmt.Println(counter)
				}
			}
		}
		return nil
	}
	return nil
}

func main() {

	convertToCsv("results.csv")
	fmt.Println("===============================")

	/*
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.ParseFiles("templates/index.html")
			tmpl.Execute(w, nil)
			if err != nil {
				fmt.Println(err)
			}
		})
		fmt.Println("Server Running at LocalHost")
		http.ListenAndServe(":8080", nil)
	*/
}
