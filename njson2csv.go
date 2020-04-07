package main

import (
	"flag"
	"io"
	"log"
	"github.com/thecodedproject/njson2csv/util"
	"os"
)

const (
	fileFieldName = "__nsjon_file"
)

var outputFile = flag.String("o", "out.csv", "Output csv file")

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("usage: njson2csv [-o output_file] njson_file [more_njson_files...]")
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

	headers, err := createHeaders(files)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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

func createHeaders(files []*os.File) (util.Headers, error) {

	var h util.Headers
	h.Add(fileFieldName)
	for _, f := range files {
		var err error
		h, err = util.AddHeaders(h, f)
		if err != nil {
			return util.Headers{}, err
		}
	}
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
		)
		if err != nil {
			return err
		}
	}
	return nil
}
