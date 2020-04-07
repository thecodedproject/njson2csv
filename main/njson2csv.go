package main

import (
	"flag"
	"log"
	//"io"
	"github.com/thecodedproject/njson2csv"
	"os"
)

var outputFile = flag.String("o", "out.csv", "Output csv file")

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("usage: njson2csv [-o output_file] njson_file [...]")
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

	out.Write(headers.CsvLine())

/*
	for _, f := range files {

		var h njson2csv.Headers
		for fieldName, _ := range test.constantFields {
			h.Add(fieldName)
		}
		h, err := njson2csv.AddHeaders(h, reader)
		require.NoError(t, err)
		reader.Seek(0, io.SeekStart)


	}
*/

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

func createHeaders(files []*os.File) (njson2csv.Headers, error) {

	var h njson2csv.Headers
	h.Add("__njson_file")
	for _, f := range files {
		var err error
		h, err = njson2csv.AddHeaders(h, f)
		if err != nil {
			return njson2csv.Headers{}, err
		}
	}
	return h, nil
}
