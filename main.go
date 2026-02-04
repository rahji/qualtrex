package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

var (
	CSVFile   string
	TypstFile string
	help      bool
)

func main() {
	parseFlags()

	reader, err := getCSV(CSVFile)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	if len(rows) < 3 {
		log.Fatal("expected at least 3 rows in a qualtrics export")
	}

	// first three rows are metadata, common for all of the following rows
	qualtricsID := rows[0]
	text := rows[1]
	importID := rows[2]

	var importIDJSON struct {
		ImportID string `json:"ImportId"`
	}
	metadata := make([]map[string]string, len(qualtricsID))

	// loop through each column, to get metadata for that col from the first three rows
	for i := range rows[0] {
		if err := json.Unmarshal([]byte(importID[i]), &importIDJSON); err != nil {
			log.Fatal("couldn't get qualtrics import id from " + importID[i])
		}
		// xxx change this to be a map w/ importID as string key -> map w/string keys
		// xxx later, the answer key will be added to the innter map
		metadata[i] = map[string]string{
			"importID":    importIDJSON.ImportID,
			"qualtricsID": qualtricsID[i],
			"text":        text[i],
		}
	}

	// loop through the remaining rows, to get the actual data/answer for each col
	for row := range rows {
		// skip the first three rows
		if row < 3 {
			continue
		}
		// xxx fix according to new structure
		data := make([]map[string]string, len(metadata))
		copy(data, metadata)
		// loop over the copied slice of meta data and complete it with the data from this row
		for i := range data {
			data[i]["answer"] = rows[row][i]
		}
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		JSONFile := fmt.Sprintf("%03d.json", row-3)
		writeFile(JSONFile, jsonBytes)

		// if TypstFile != "" {
		// 	// try to make pdf for this json file, using typst
		// 	if _, err := os.Stat(TypstFile); os.IsNotExist(err) {
		// 		log.Fatal(err)
		// 	}

		// 	cmd := exec.Command("typst", "compile", TypstFile)
		// 	_, err := cmd.Output()
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }
	}
}

func writeFile(fn string, b []byte) {
	err := os.WriteFile(fn, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags() {
	pflag.StringVarP(&CSVFile, "csvfile", "i", "", "CSV input file (if not specified, reads from STDIN)")
	pflag.StringVarP(&TypstFile, "typstfile", "t", "", "Typst document input file, for PDF output")
	pflag.BoolVarP(&help, "help", "h", false, "show help message")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	if help {
		pflag.Usage()
		os.Exit(0)
	}
}

// getCSV reads from an input filename or from STDIN if the filename is empty
// returns a CSV reader
func getCSV(infile string) (*csv.Reader, error) {
	var reader *csv.Reader
	// if file flag is provided, try to open that file
	if infile != "" {
		f, err := os.Open(infile)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %v", err)
		}
		defer f.Close()
		reader = csv.NewReader(f)
	} else {
		// get the csv data from STDIN
		stat, _ := os.Stdin.Stat()
		// unless the STDIN is not piped data
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, errors.New("no input file specified and nothing piped to STDIN")
		}
		reader = csv.NewReader(os.Stdin)
	}
	return reader, nil
}
