package main

import (
	"encoding/csv"
	"encoding/json"
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

type Response struct {
	Message string `json:"message"`
}

type pageData struct {
	Results []gradeInfo
}

func multipleForms (subject string, grade string) Response {

	details := gradeInfo{
		Subject: subject,
		Grade:   grade,
		Show:    false,
	}

	record := []string{}
	// Process the form data
	details.Percent, record, details.Subject = convertToCsv("results.csv", details.Subject, details.Grade)
	if details.Percent > 0 {
		details.Show = true
		fmt.Println(record)
		fmt.Println(details)
	}

	// Prepare the response message
	var response Response
	if details.Show {
		response.Message = fmt.Sprintf("Your Grade Puts you into the top: %.2f%% in %s", details.Percent, details.Subject)
	} else {
		response.Message = "Grade not found or invalid."
	}

	return response 

}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	subject := r.FormValue("subject")
	grade   := r.FormValue("grade")

	response := multipleForms(subject,grade)

	// Set Content-Type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode the response as JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}


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

func convertToCsv(fileName string, subject string, grade string) (float64, []string, string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error", err)
		return -1.0, nil, "nil"
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reader", err)
		return -1.0, nil, subject
	}

	for _, record := range records {
		if searchRecords(subject, record) {
			counter := findGrade(record, grade)
			subject = record[1]
			return counter, record, subject
		}
	}

	return -1.0, nil, subject
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/submit", formHandler)
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

			record := []string{}
			details.Percent, record, details.Subject = convertToCsv("results.csv", details.Subject, details.Grade)
			if details.Percent > 0 {
				details.Show = true
				fmt.Println(record)
				fmt.Println(details)
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
