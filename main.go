package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"
)

var (
	CSVFile   string
	TypstFile string
	help      bool
)

const exportFolder = "./exports"

func main() {
	parseFlags()

	var err error
	// make an output folder for json and (maybe) typst and pdf files
	err = os.MkdirAll(exportFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// read the csv file into a slice (rows) of slices (cols)
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
	for c := range rows[0] { // c == column slice
		m := make(map[string]string)
		if err := json.Unmarshal([]byte(rows[2][c]), &importIDJSON); err != nil {
			log.Fatal("couldn't get qualtrics import id from " + rows[2][c])
		}
		m["importID"] = importIDJSON.ImportID
		m["text"] = rows[1][c]
		m["qualtricsID"] = rows[0][c]
		metadata[c] = m
	}

	// for each row (each to become a separate json file):
	//   1. loop through each column, adding a new map entry to a map that looks like:
	//      `importid: { qid:str, text:str, answer:str}`
	//      (where importid = current_col:row2, qid = current_col:row0, text = current_col:row1)
	//   2. create a json file based on the map
	//   3. prepend info re: the json file onto the typst document and write the new file to the typst folder

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
		JSONFile := fmt.Sprintf("%03d.json", ir-3)
		writeFile(exportFolder+"/"+JSONFile, jsonBytes)

		// if --typstfile is specified, then make a copy of the typstfile that points to the json file
		// and use the typst command to compile that new json-referencing typstfile to create a pdf
		if TypstFile != "" {
			if _, err := os.Stat(TypstFile); os.IsNotExist(err) {
				log.Fatal(err)
			}
			thisTypstFile, err := typstFromTemplate(TypstFile, JSONFile)
			if err != nil {
				log.Fatal(err)
			}
			cmd := exec.Command("typst", "compile", thisTypstFile)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Dir = exportFolder
			err = cmd.Run()
			if err != nil {
				log.Fatalf("Failed to run typst: %s", stderr.String())
			}
		}
	}
}

// writeFile writes a slice of bytes to a named file.
func writeFile(fn string, b []byte) {
	err := os.WriteFile(fn, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// typstFromTemplate makes a copy of the typist file, prepends a reference to the json data file,
// and saves it in the exports folder with a name that matches the json file
func typstFromTemplate(tf, jf string) (string, error) {
	base, _ := strings.CutSuffix(jf, ".json")
	tempfn := base + ".tmp"
	tempFile, err := os.Create(tempfn)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// open original file for reading
	f, err := os.Open(tf)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// prepend the text
	str := fmt.Sprintf("#let q = json(\"%s\")\n\n", jf)
	_, err = tempFile.WriteString(str)
	if err != nil {
		return "", err
	}

	// read the original typst file and append its text to the temp file
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		_, err = tempFile.WriteString(scanner.Text())
		_, err = tempFile.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "nil", err
	}
	tempFile.Sync() // flush the writes

	newFile := base + ".typ"
	newFileWithPath := fmt.Sprintf("%s/%s", exportFolder, newFile)
	err = os.Rename(tempfn, newFileWithPath)
	if err != nil {
		return "", err
	}

	return newFile, nil
}

// parseFlags parses all of the CLI flags and populates the
func parseFlags() {
	pflag.StringVarP(&CSVFile, "csvfile", "i", "", "CSV input file (if not specified, reads from STDIN)")
	pflag.StringVarP(&TypstFile, "typstfile", "t", "", "Typst document input file (default is not to use typst)")
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
