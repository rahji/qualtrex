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
	CSVFile     string
	TypstFile   string
	JSONFolder  string
	TypstFolder string
	PDFFolder   string
	help        bool
)

func main() {
	parseFlags()

	var err error
	// make the output folders (json,typst,pdf)
	err = os.MkdirAll(JSONFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(TypstFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(PDFFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}

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

	numCols := len(rows[0])

	// get metadata for each column and store it in a slice of maps
	// (ImportID is stored as JSON in one of the fields)
	var importIDJSON struct {
		ImportID string `json:"ImportId"`
	}
	metadata := make([]map[string]string, numCols)
	for i := range rows[0] { // i == column slice
		m := make(map[string]string)
		if err := json.Unmarshal([]byte(rows[2][i]), &importIDJSON); err != nil {
			log.Fatal("couldn't get qualtrics import id from " + rows[2][i])
		}
		m["importID"] = importIDJSON.ImportID
		m["text"] = rows[1][i]
		m["qualtricsID"] = rows[0][i]
		metadata[i] = m
	}

	// for each row (each to become a separate json file):
	//   1. loop through each column, adding a new map entry to a map that looks like:
	//      `importid: { qid:str, text:str, answer:str}`
	//      (where importid = current col row 2, qid = current column row 0, text = current col row 1)
	//   2. create a json file based on the map
	//   3. prepend info re: the json file onto the typst document and write the new file to the typst folder
	// run typst against the json files

	for ir, row := range rows {
		// skip the first three rows
		if ir < 3 {
			continue
		}
		// make a map of maps that will be this row's json data
		data := make(map[string]map[string]string, numCols)
		for ic := range len(row) {
			importID := metadata[ic]["importID"]
			m := make(map[string]string)
			m["text"] = metadata[ic]["text"]
			m["qualtricsID"] = metadata[ic]["qualtricsID"]
			m["answer"] = row[ic]
			data[importID] = m
		}
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		JSONFile := fmt.Sprintf("%s/%03d.json", JSONFolder, ir-3)
		writeFile(JSONFile, jsonBytes)
	}

	// 	// if TypstFile != "" {
	// 	// 	// try to make pdf for this json file, using typst
	// 	// 	if _, err := os.Stat(TypstFile); os.IsNotExist(err) {
	// 	// 		log.Fatal(err)
	// 	// 	}

	// 	// 	cmd := exec.Command("typst", "compile", TypstFile)
	// 	// 	_, err := cmd.Output()
	// 	// 	if err != nil {
	// 	// 		log.Fatal(err)
	// 	// 	}
	// 	// }
	// }
}

// writeFile writes a slice of bytes to a named file
func writeFile(fn string, b []byte) {
	err := os.WriteFile(fn, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// parseFlags parses all of the CLI flags and populates the
func parseFlags() {
	pflag.StringVarP(&CSVFile, "csvfile", "i", "", "CSV input file (if not specified, reads from STDIN)")
	pflag.StringVarP(&TypstFile, "typstfile", "t", "./qualtrics.typ", "Typst document input file (default: ./qualtrics.typ)")
	pflag.StringVarP(&JSONFolder, "jsondir", "", "./json", "JSON output folder (default: ./json)")
	pflag.StringVarP(&TypstFolder, "typstdir", "", "./typst", "Typst output folder (default: ./typst)")
	pflag.StringVarP(&PDFFolder, "pdfdir", "", "./pdf", "PDF output folder (default: ./pdf)")
	pflag.BoolVarP(&help, "help", "h", false, "show help message")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	if help {
		pflag.Usage()
		os.Exit(0)
	}
}

// getCSV reads from an input filename or from STDIN if the filename is empty
// returns a CSV reader and an error
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

// makeFolder creates a folder if it doesn't exist
func makeFolder(d string) error {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		return os.Mkdir(d, os.ModeDir|0755)
	}
	return nil
}
