package main

import (
	"flag"
	"io"
	"log"
	"github.com/thecodedproject/njson2csv/util"
	"os"
	"strings"
)

const (
	fileFieldName = "__nsjon_file"
)

var outputFile = flag.String("o", "out.csv", "Output csv file")
var filterColumnsString = flag.String("columns", "", "Comma seperated list of columns to keep (e.g. `\"col1,col2\"`)")
var ioBufferSize = flag.Int("io_buffer", 4096, "The maximum line length the io reader will read")

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("usage: njson2csv [-o output_file] [-columns \"columnName,...\"] njson_file [more_njson_files...]")
	}

	assertFileDoesNotExist(*outputFile)

	files, err := openFiles(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		defer f.Close()
	}

	out, err := os.OpenFile(*outputFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	var filterColumns []string
	if *filterColumnsString != "" {
		filterColumns = strings.Split(*filterColumnsString, ",")
	}

	headers, err := createHeaders(
		files,
		filterColumns,
	)
	if err != nil {
		log.Println("Error finding column names:", err)
		logFatalJsonErrorMessage()
	}

	_, err = out.Write(headers.CsvLine())
	if err != nil {
		log.Fatal(err)
	}

	err = seekFilesToStart(files)
	if err != nil {
		log.Fatal(err)
	}

	err = writeValues(out, files, &headers)
	if err != nil {
		log.Println("Error writing values:", err)
		logFatalJsonErrorMessage()
	}
}

func assertFileDoesNotExist(filename string) {

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Fatal(err)
	}
	log.Fatal(filename, " already exists")
}

func openFiles(filenames []string) ([]*os.File, error) {

	files := make([]*os.File, 0, len(filenames))
	for _, filename := range flag.Args() {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		files = append(files, f)
	}
	return files, nil
}

func createHeaders(files []*os.File, filterColumns []string) (util.Headers, error) {

	var h util.Headers
	var err error
	for _, f := range files {
		h, err = util.AddHeaders(h, f, *ioBufferSize)
		if err != nil {
			return util.Headers{}, err
		}
	}

	if len(filterColumns) != 0 {
		h, err = util.FilterHeaders(h, filterColumns)
		if err != nil {
			return util.Headers{}, err
		}
	}

	h.Add(fileFieldName)

	return h, nil
}

func seekFilesToStart(files []*os.File) error {

	for _, f := range files {
		_, err := f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeValues(outFile *os.File, files []*os.File, h util.HeaderPos) error {

	for _, f := range files {
		fileValues := map[string]string{
			fileFieldName: f.Name(),
		}

		err := util.WriteLines(
			outFile,
			f,
			h,
			fileValues,
			*ioBufferSize,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func logFatalJsonErrorMessage() {

		log.Println()
		log.Println("JSON decoding errors may be caused by truncated lines (if the io.Reader buffer is smaller than the max line length)")
		log.Println("Get the max line length with:")
		log.Println("")
		log.Println("\tawk 'length > l {l=length;line=$0} END {print l}'", strings.Join(flag.Args(), " "))
		log.Println("")
		log.Fatal("Then adjust to `max_length+1` with the `--io_buffer` cmd option.")
}
