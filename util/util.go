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
	FieldWasRemoved(field string) bool
	NumFields() int
}

type Headers struct {
	fields []string
	removedFields []string
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

func (h *Headers) FieldWasRemoved(field string) bool {

	return contains(h.removedFields, field)
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

func (h *Headers) remove(field string) {

	i, _ := h.Position(field)
	copy(h.fields[i:], h.fields[i+1:])
	h.fields[len(h.fields)-1] = ""
	h.fields = h.fields[:len(h.fields)-1]
	h.removedFields = append(h.removedFields, field)
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

func FilterHeaders(h Headers, filterColumns []string) (Headers, error) {

	for _, col := range filterColumns {
		if !contains(h.fields, col) {
			return Headers{}, fmt.Errorf(
				"Cannot filter column `%s`; no such column",
				col,
			)
		}
	}

	headersToRemove := make([]string, 0, len(h.fields))

	for _, header := range h.fields {
		if !contains(filterColumns, header) {
			headersToRemove = append(headersToRemove, header)
		}
	}

	for _, header := range headersToRemove {
		h.remove(header)
	}

	return h, nil
}

func WriteLines(
	w io.Writer,
	r io.Reader,
	h HeaderPos,
	constantFields map[string]string,
	maxReaderLength int,
) error {

	fieldValues := make([]string, h.NumFields())

	bufReader := bufio.NewReaderSize(r, maxReaderLength)
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
		} else if !h.FieldWasRemoved(headerName) {
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
