package njson2csv_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/njson2csv"
	"io"
	"strings"
	"testing"
)

func TestGetHeaders(t *testing.T) {

	testCases := []struct{
		name string
		njson string
		expectedHeader string
		expectedOrder []string
		expectedErr error
	}{
		{
			name: "Single njson line - headers sorted alphabetically",
			njson: "{\"a\": 0, \"b\": true}",
			expectedHeader: "a,b,\n",
			expectedOrder: []string{"a", "b"},
		},
		{
			name: "Single njson line with nested fields",
			njson: "{\"a\": 0, \"b\": {\"a\": 1, \"b\": 2}}",
			expectedHeader: "a,b_a,b_b,\n",
			expectedOrder: []string{"a", "b_a", "b_b"},
		},
		{
			name: "Single njson line with triple nested fields",
			njson: "{\"a\": 0, \"b\": {\"a\": {\"a\": 1, \"c\": 0}, \"b\": 2}}",
			expectedHeader: "a,b_a_a,b_a_c,b_b,\n",
			expectedOrder: []string{"a", "b_a_a", "b_a_c", "b_b"},
		},
		{
			name: "Multiple njson lines with same fields",
			njson: "{\"a\": \"abc\", \"b\": 1}\n{\"a\": false, \"b\": 1}\n",
			expectedHeader: "a,b,\n",
			expectedOrder: []string{"a", "b"},
		},
		{
			name: "Multiple njson lines with extra fields",
			njson: "{\"a\": 10, \"b\": 11}\n{\"a\": 0, \"b\": 1, \"c\": 1, \"d\": 0}\n",
			expectedHeader: "a,b,c,d,\n",
			expectedOrder: []string{"a", "b", "c", "d"},
		},
		{
			name: "Multiple njson lines with different fields",
			njson: "{\"a\": 10, \"c\": 11}\n{\"b\": 0, \"f\": 1}\n{\"e\": 1, \"d\": 0}\n",
			expectedHeader: "a,b,c,d,e,f,\n",
			expectedOrder: []string{"a", "b", "c", "d", "e", "f"},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			reader := strings.NewReader(test.njson)

			var h njson2csv.Headers
			h, err := njson2csv.AddHeaders(h, reader)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, test.expectedHeader, string(h.CsvLine()))

			for pos, field := range test.expectedOrder {
				actualPos, err := h.Position(field)
				require.NoError(t, err)
				assert.Equal(t, pos, actualPos, field)
			}
		})
	}
}

type LineWriter struct {
	Lines []string
}

func (l *LineWriter) Write(p []byte) (n int, err error) {

	l.Lines = append(l.Lines, string(p))
	return len(p), nil
}

func TestWriteLines(t *testing.T) {

	testCases := []struct{
		name string
		njson string
		constantFields map[string]string
		expectedLines []string
	}{
		{
			name: "Empty reader write no lines",
		},
		{
			name: "Single line with string fields",
			njson: "{\"field_a\": \"a\", \"field_b\": \"b\"}",
			expectedLines: []string{
				"a,b,\n",
			},
		},
		{
			name: "Single line with string int bool float fields",
			njson: "{\"a\": \"a\", \"b\": 21, \"c\": false, \"d\": true, \"f\": 1.234}",
			expectedLines: []string{
				"a,21,false,true,1.234,\n",
			},
		},
		{
			name: "Single line with double nested fields",
			njson: "{\"a\": 0, \"b\": {\"c\": 1, \"d\": 2}}",
			expectedLines: []string{
				"0,1,2,\n",
			},
		},
		{
			name: "Single line with triple nested fields",
			njson: "{\"a\": 0, \"b\": {\"c\": {\"a\": 11, \"c\": 42}, \"d\": 2}}",
			expectedLines: []string{
				"0,11,42,2,\n",
			},
		},
		{
			name: "Multiple lines with the same fields",
			njson: `{"a": 11,			"b": 21, 		"c": false}
							{"a": false,	"b": true, 	"c": 1.34}
							{"a": 1.3,		"b": "a", 	"c": 43}
							{"a": "14",		"b": 2.4, 	"c": "fds"}`,
			expectedLines: []string{
				"11,21,false,\n",
				"false,true,1.34,\n",
				"1.3,a,43,\n",
				"14,2.4,fds,\n",
			},
		},
		{
			name: "Multiple lines with some fields missing",
			njson: `{"a": 11,	"c": false}
							{"b": true, 	"c": 1.34}
							{"a": 1.3,		"b": "a", 	"c": 43}
							{"c": "fds"}`,
			expectedLines: []string{
				"11,,false,\n",
				",true,1.34,\n",
				"1.3,a,43,\n",
				",,fds,\n",
			},
		},
		{
			name: "Multiple lines with some constant fields",
			njson: `{"a": 11,	"c": false}
							{"b": true, 	"c": 1.34}
							{"a": 1.3,		"b": "a", 	"c": 43}
							{"c": "fds"}`,
			constantFields: map[string]string{
				"hello": "world",
				"aaa": "value",
			},
			expectedLines: []string{
				"11,value,,false,world,\n",
				",value,true,1.34,world,\n",
				"1.3,value,a,43,world,\n",
				",value,,fds,world,\n",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			reader := strings.NewReader(test.njson)
			var writer LineWriter

			var h njson2csv.Headers
			for fieldName, _ := range test.constantFields {
				h.Add(fieldName)
			}
			h, err := njson2csv.AddHeaders(h, reader)
			require.NoError(t, err)
			reader.Seek(0, io.SeekStart)

			err = njson2csv.WriteLines(&writer, reader, &h, test.constantFields)
			require.NoError(t, err)

			assert.Equal(t, test.expectedLines, writer.Lines)
		})
	}
}

func TestWriteLinesWhenHeaderNotFoundReturnsError(t *testing.T) {

	reader := strings.NewReader(
		"{\"a\": 0}",
	)
	var writer LineWriter
	var h njson2csv.Headers
	var cf map[string]string

	err := njson2csv.WriteLines(&writer, reader, &h, cf)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "a")
}
