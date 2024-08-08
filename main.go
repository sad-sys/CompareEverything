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

// gradeInfo struct holds the subject, grade, whether to show the result, and the calculated percentile
type gradeInfo struct {
	Subject string
	Grade   string
	Show    bool
	Percent float64
}

// Response struct holds the message to be returned as JSON response
type Response struct {
	Message string `json:"message"`
}

// allDetails holds all the grade information for multiple submissions
var allDetails = []gradeInfo{}

// multipleForms processes the form input and determines the percentile
func multipleForms(subject string, grade string, formType string) (Response, gradeInfo) {
	// Initialize a new gradeInfo instance
	details := gradeInfo{
		Subject: subject,
		Grade:   grade,
		Show:    false,
	}

	var record []string
	// Determine which CSV file to use based on form type
	if formType == "IQ" {
		details.Percent, record, details.Subject = convertToCsv("iq.csv", details.Subject, details.Grade, formType)
	} else if formType == "A-Level" {
		details.Percent, record, details.Subject = convertToCsv("results.csv", details.Subject, details.Grade, formType)
	}
	if formType == "Height" {
		fmt.Println("FORMTYPE HEIGHT", details.Subject, details.Grade, formType)
		details.Percent, record, details.Subject = convertToCsvHeight("height.csv", details.Subject, details.Grade, formType)
	}

	// Print the record and details for debugging
	fmt.Printf("Record: %v\n", record)
	fmt.Printf("Details: %+v\n", details)

	// Show the result if the percentile is valid
	if details.Percent >= 0 {
		details.Show = true
		fmt.Println(record)
		fmt.Println(details)
	}

	// Prepare the response message
	var response Response
	if details.Show {
		response.Message = fmt.Sprintf("Your %s puts you into the top: %.2f%% in %s", formType, details.Percent, details.Subject)
	} else {
		response.Message = fmt.Sprintf("%s not found or invalid.", formType)
	}

	return response, details
}

// formHandler handles the form submission and processes the input
func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read form values
	formType := r.FormValue("formType")
	subject := strings.ToUpper(r.FormValue("subject"))
	grade := strings.ToUpper(r.FormValue("grade"))

	// Print form values for debugging
	fmt.Printf("Form Type: %s, Subject: %s, Grade: %s\n", formType, subject, grade)

	// Process the form and get the response and details
	response, details := multipleForms(subject, grade, formType)
	allDetails = append(allDetails, details)
	fmt.Printf("Response being sent: %+v\n", response)
	// Set response header and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}

// findIQ finds the percentile for the given IQ value from the CSV record
func findIQ(record []string, iq string) float64 {
	fmt.Printf("Received IQ value: %s\n", iq)
	iqValue, err := strconv.Atoi(iq)
	if err != nil {
		fmt.Println("Error converting IQ value to integer:", err)
		return -1.0
	}

	// IQ thresholds and corresponding CSV columns
	iqThresholds := []int{145, 130, 115, 100, 85, 70, 55}
	percentiles := make([]float64, len(iqThresholds)+1)
	for i := range iqThresholds {
		percentiles[i], err = strconv.ParseFloat(record[i+3], 64)
		if err != nil {
			fmt.Println("Error parsing percentile from record:", err)
			return -1.0
		}
		fmt.Printf("Threshold: %d, Percentile: %f\n", iqThresholds[i], percentiles[i])
	}
	// Handling the lowest threshold case separately
	percentiles[len(iqThresholds)], err = strconv.ParseFloat(record[len(record)-1], 64)
	if err != nil {
		fmt.Println("Error parsing the lowest threshold percentile from record:", err)
		return -1.0
	}

	// Iterate through thresholds to find the correct percentile
	for i, threshold := range iqThresholds {
		if iqValue >= threshold {
			fmt.Printf("IQ value %d is greater than or equal to threshold %d, returning percentile %f\n", iqValue, threshold, percentiles[i])
			return percentiles[i]
		}
	}

	// If the IQ value is less than the lowest threshold, return the highest percentile (closest to 100)
	fmt.Println("IQ value is less than the lowest threshold, returning highest percentile")
	return percentiles[len(iqThresholds)]
}

func findHeight(record []string, gender string, height string) float64 {
	percentiles := []int{95, 90, 85, 75, 50, 25, 15, 10, 5}
	maleHeightThresholds := []float64{64.8, 66.3, 67.0, 68.2, 70.2, 71.6, 72.8, 73.7, 74}
	femaleHeightThresholds := []float64{60.6, 61.5, 62.2, 62.7, 64.2, 65.3, 66.0, 66.5, 68.1}

	heightValue, err := strconv.ParseFloat(height, 64)
	if err != nil {
		fmt.Println(err)
		return 0.0 // Return if there's an error parsing the height
	}

	if gender == "MALE" {
		fmt.Println("Processing Male Height")
		// Find the percentile
		for i := 0; i < len(maleHeightThresholds); i++ {
			if heightValue < maleHeightThresholds[i] {
				if i == 0 {
					fmt.Println("THRESHOLD, PERCENTILE", maleHeightThresholds[i], percentiles[i])
					return float64(percentiles[i])
				}
				fmt.Println("THRESHOLD, PERCENTILE", maleHeightThresholds[i-1], percentiles[i-1])
				return float64(percentiles[i-1])
			}
		}
		// If the height is equal to or greater than the last threshold
		fmt.Println("THRESHOLD, PERCENTILE", maleHeightThresholds[len(maleHeightThresholds)-1], percentiles[len(percentiles)-1])
		return float64(percentiles[len(percentiles)-1])
	} else if gender == "FEMALE" {
		fmt.Println("Processing female Height")
		// Find the percentile
		for i := 0; i < len(femaleHeightThresholds); i++ {
			if heightValue < femaleHeightThresholds[i] {
				if i == 0 {
					fmt.Println("THRESHOLD, PERCENTILE", femaleHeightThresholds[i], percentiles[i])
					return float64(percentiles[i])
				}
				fmt.Println("THRESHOLD, PERCENTILE", femaleHeightThresholds[i-1], percentiles[i-1])
				return float64(percentiles[i-1])
			}
		}
	} else if gender == "OTHER" {
		fmt.Println("Processing other height")
		for i := 0; i < len(femaleHeightThresholds); i++ {
			if heightValue < femaleHeightThresholds[i] {
				if i == 0 {
					fmt.Println("THRESHOLD, PERCENTILE", femaleHeightThresholds[i], percentiles[i])
					return float64(percentiles[i])
				}
				fmt.Println("THRESHOLD, PERCENTILE", femaleHeightThresholds[i-1], percentiles[i-1])
				return float64(percentiles[i-1])
			}
		}
		// If the height is equal to or greater than the last threshold
		fmt.Println("THRESHOLD, PERCENTILE", femaleHeightThresholds[len(femaleHeightThresholds)-1], percentiles[len(percentiles)-1])
		return float64(percentiles[len(percentiles)-1])
	}
	return 0.0
}

// findGrade finds the percentile for the given grade from the CSV record
func findGrade(record []string, grade string) float64 {
	grades := map[string]int{
		"A*": 3,
		"A":  4,
		"B":  5,
		"C":  6,
		"D":  7,
		"E":  8,
		"U":  9,
	}

	// Get the index of the grade column and retrieve the percentile
	if index, ok := grades[grade]; ok {
		percentile, err := strconv.ParseFloat(record[index], 64)
		if err != nil {
			fmt.Println(err)
			return -1.0
		}
		return percentile
	}
	return -1.0
}

// searchRecords checks if the word exists in the CSV record
func searchRecords(word string, record []string) bool {
	for _, field := range record {
		if strings.Contains(field, word) {
			return true
		}
	}
	return false
}

// convertToCsv reads the CSV file and finds the corresponding percentile
func convertToCsv(fileName string, subject string, grade string, formType string) (float64, []string, string) {
	// Open the CSV file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", fileName, err)
		return -1.0, nil, "nil"
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV file %s: %v\n", fileName, err)
		return -1.0, nil, subject
	}

	// Print the records for debugging
	fmt.Printf("Records from %s: %v\n", fileName, records)

	// Process each record
	for _, record := range records {
		fmt.Printf("Processing record: %v\n", record)

		if formType == "IQ" && len(record) >= 10 {
			if record[1] == "IQ" { // Match the type field for IQ
				counter := findIQ(record, grade)
				subject = record[1]
				return counter, record, subject
			}
		} else if formType == "A-Level" && len(record) >= 10 {
			if searchRecords(subject, record) {
				counter := findGrade(record, grade)
				subject = record[1]
				return counter, record, subject
			}
		}
	}

	return -1.0, nil, subject
}

func convertToCsvHeight(fileName string, gender string, height string, formType string) (float64, []string, string) {
	// Open the CSV file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", fileName, err)
		return -1.0, nil, "nil"
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV file %s: %v\n", fileName, err)
		return -1.0, nil, gender
	}

	// Print the records for debugging
	fmt.Printf("Records from %s: %v\n", fileName, records)

	for _, record := range records {
		fmt.Printf("Processing record: %v\n", record)

		if searchRecords(gender, record) {
			counter := findHeight(record, gender, height)
			gender = record[2]
			return counter, record, gender
		} else {
			fmt.Println("Not Found")
		}
	}

	return -1.0, nil, gender
}
func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle form submission
	http.HandleFunc("/submit", formHandler)

	// Handle the main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		details := gradeInfo{
			Show: false,
		}

		// Parse and execute the template
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

	// Start the server
	fmt.Println("Server Running at LocalHost")
	http.ListenAndServe(":8080", nil)
}
