package util

import (
	"bufio"
	"encoding/json"
	"io"
	"sort"
	"fmt"
)

type HeaderPos interface{
	Position(field string) (int, error)
	NumFields() int
}

type Headers struct {
	fields []string
}

func (h *Headers) CsvLine() []byte {

	return sliceToCsvLine(h.fields)
}

func (h *Headers) Position(field string) (int, error) {

	for i, f := range h.fields {
		if f == field {
			return i, nil
		}
	}
	return 0, fmt.Errorf("No header found for field '%s'", field)
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
			iField, err := h.Position(fieldName)
			if err != nil {
				return err
			}
			fieldValues[iField] = fmt.Sprint(value)
		}

		err = addSubValues(fieldValues, h, "", m)
		if err != nil {
			return err
		}

		_, err = w.Write(
			sliceToCsvLine(fieldValues),
		)
		if err != nil {
			return err
		}
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
) error {

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
			iField, err := h.Position(headerName)
			if err != nil {
				return err
			}
			fieldValues[iField] = fmt.Sprint(childValue)
		}
	}
	return nil
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
