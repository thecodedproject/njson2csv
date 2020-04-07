package njson2csv

import (
	"bufio"
	"encoding/json"
	"io"
	"sort"
	"fmt"
)

type HeaderPos interface{
	Position(field string) int
	NumFields() int
}

type Headers struct {
	fields []string
}

func (h *Headers) CsvLine() []byte {

	return sliceToCsvLine(h.fields)
}

func (h *Headers) Position(field string) int {

	for i, f := range h.fields {
		if f == field {
			return i
		}
	}
	return 0
}

func (h *Headers) NumFields() int {

	return len(h.fields)
}

func (h *Headers) Add(field string) {

	if !contains(h.fields, field) {
		h.fields = append(h.fields, field)
		sort.Strings(h.fields)
	}
}

func AddHeaders(h Headers, r io.Reader) (Headers, error) {

	bufReader := bufio.NewReader(r)
	iLine := 0
	for {
		line, _, err := bufReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return Headers{}, err
		}

		m := make(map[string]interface{})
		err = json.Unmarshal(line, &m)

		addSubFields(&h, "", m)

		iLine++
	}

	return h, nil
}

func WriteLines(
	w io.Writer,
	r io.Reader,
	h HeaderPos,
	constantFields map[string]string,
) error {

	fieldValues := make([]string, h.NumFields())

	bufReader := bufio.NewReader(r)
	for {
		line, _, err := bufReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		m := make(map[string]interface{})
		err = json.Unmarshal(line, &m)
		if err != nil {
			return err
		}

		for i := range fieldValues {
			fieldValues[i] = ""
		}

		for fieldName, value := range constantFields {
			iField := h.Position(fieldName)
			fieldValues[iField] = fmt.Sprint(value)
		}

		addSubValues(fieldValues, h, "", m)

		w.Write(
			sliceToCsvLine(fieldValues),
		)
	}

	return nil
}

func addSubFields(h *Headers, parentName string, m map[string]interface{}) {

	for childName := range m {
		headerName := fieldName(parentName, childName)
		childDict, isDict := m[childName].(map[string]interface{})
		if isDict {
			addSubFields(
				h,
				headerName,
				childDict,
			)
		} else {
			h.Add(headerName)
		}
	}
}

func fieldName(parentName, childName string) string {

	if parentName == "" {
		return childName
	}

	return fmt.Sprint(parentName, "_", childName)
}

func addSubValues(
	fieldValues []string,
	h HeaderPos,
	parentName string,
	m map[string]interface{},
) {

	for childName, childValue := range m {

		headerName := fieldName(parentName, childName)

		childDict, isDict := childValue.(map[string]interface{})
		if isDict {
			addSubValues(
				fieldValues,
				h,
				headerName,
				childDict)
		} else {
			iField := h.Position(headerName)
			fieldValues[iField] = fmt.Sprint(childValue)
		}
	}
}

func contains(s []string, v string) bool {

	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

func sliceToCsvLine(s []string) []byte {

	var str string
	for _, value := range s {
		str += value + ","
	}
	str += "\n"
	return []byte(str)
}
