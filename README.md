# QualtrEX

It's crazy to me that Qualtrics doesn't have an option for exporting a batch of survey results as a PDF.
Even if you export them one by one, it still contains *all* of the information about that submission
(including a map of their location, etc.) It doesn't look good and there's no way to customize it. Another
failing is that there is no way to export the results of a survey as JSON unless you are in their developer
program.

This command-line tool allows you to create JSON and PDF files from a CSV file that you've exported from
the Data & Analysis tab of a Qualtrics survey.

## Command-Line Options

There are two command-line options:

```
-i, --csvfile string     CSV input file
-t, --typstfile string   Typst document input file
```

A CSV file from Qualtrics is required. If you don't specify a Typst file, the command will only generate
JSON files from the CSV input. All output files are put in a folder called `exports`, which is created if
it doesn't already exist.

## Customizing the PDF Output

If a Typst file is specified, a PDF file will be made for each survey. The filenames are numbered, starting with `000.pdf`.
The Typst file is where the look of the PDFs is defined. Use Typst syntax in the file, which is similar to Markdown.
More details on how to format a document can be found in the [Typst documentation](https://typst.app/docs/) or the
[Typst Examples Book](https://sitandr.github.io/typst-examples-book/book/).

To access the survey data from within the Typst file, use the syntax `#q.Q1.text` or `#q.Q1.answer`, where `Q1` is the
`importID` from the CSV file. (You can also access `#q.Q1.qualtricsID`, which is the question ID that appears above your
question when you are looking at the survey on the Qualtrics site.)

The repo contains two sample Typst example files ([simple.typ](simple.typ) and [advanced.typ](advanced.typ)>) and a [sample CSV file](test.csv)).

## How to Install It

First, install the latest version of the [Typst](https://typst.app) CLI

Then, install QualtrEX by downloading it from the [Releases][releases] tab to your right, then put the binary somewhere in your PATH.
If you're using Windows and new to the command-line, watch [this video where I explain how to install a command-line program](https://www.youtube.com/watch?v=ZjdqxBTLNTk&t=5s).

If you have Go installed, you can install QualtrEX by running this command instead: `go install github.com/rahji/qualtrex@latest`

If you're using macOS and it works for you, let me know. I've only tested it in Linux and Windows.

