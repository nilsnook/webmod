package webmod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var jsonTests = []struct {
	name          string
	json          string
	maxSize       int
	allowUnknown  bool
	errorExpected bool
}{
	{
		name:          "Good JSON",
		json:          `{"foo": "bar"}`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: false,
	},
	{
		name:          "Invalid JSON - Not JSON",
		json:          `foo=bar`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "Invalid JSON - missing value",
		json:          `{"foo": }`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "Invlaid JSON - Missing field name",
		json:          `{bar: 404}`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "Invalid JSON - Syntax error",
		json:          `{"foo": bar"}`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "Incorrect type",
		json:          `{"foo": 404}`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "More than one JSON",
		json:          `{"foo": "1"}{"bar", "2"}`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "Unknown field",
		json:          `{"bar": "soda"}`,
		maxSize:       1024,
		allowUnknown:  false,
		errorExpected: true,
	},
	{
		name:          "Allow unknown field",
		json:          `{"bar": 404}`,
		maxSize:       1024,
		allowUnknown:  true,
		errorExpected: false,
	},
	{
		name:          "File too Large",
		json:          `{"foo": "bar"}`,
		maxSize:       12,
		allowUnknown:  false,
		errorExpected: true,
	},
}

func TestTools_ReadJSON(t *testing.T) {
	var testTool Tools

	for _, e := range jsonTests {
		testTool.MaxJSONSize = e.maxSize
		testTool.AllowUnknownFields = e.allowUnknown

		// creating a struct type for decoding json data into
		var jdata struct {
			Foo string `json:"foo"`
		}

		// creating a request
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(e.json)))
		defer r.Body.Close()
		// creating a response recorder
		rec := httptest.NewRecorder()
		// call ReadJSON
		err := testTool.ReadJSON(rec, r, &jdata)

		if !e.errorExpected && err != nil {
			info := fmt.Sprintf("Error: %s", err.Error())
			printErr(t, e.name, "Error not expected, but one received", info)
		}

		if e.errorExpected && err == nil {
			printErr(t, e.name, "Error expected, but none received")
		}
	}
}

func TestTools_WriteJSON(t *testing.T) {
	tname := "Write JSON"
	var testTool Tools

	rec := httptest.NewRecorder()
	msg := "high hopes"
	jdata := JSONResponse{
		Error:   false,
		Message: msg,
	}

	headers := make(http.Header)
	headers.Add("Foo", "1")
	headers.Add("Foo", "2")
	headers.Add("Bar", "3")

	err := testTool.WriteJSON(rec, http.StatusOK, jdata, headers)
	if err != nil {
		msg := "Failed to write JSON"
		info := fmt.Sprintf("Error: %s", err.Error())
		printErr(t, tname, msg, info)
	}

	// decode JSON and check for correctness of values received
	err = json.NewDecoder(rec.Body).Decode(&jdata)
	// unable to decode JSON
	if err != nil {
		msg := "Invalid JSON format of response"
		info := fmt.Sprintf("Error: %s", err.Error())
		printErr(t, tname, msg, info)
	}

	// error value 'true'
	if jdata.Error {
		msg := `Invalid "error" value`
		expected := `Expected: "false"`
		received := `Received: "true"`
		printErr(t, tname, msg, expected, received)
	}

	// message value does not match
	if jdata.Message != msg {
		msg := `Invalid "message" value`
		expected := fmt.Sprintf("Expected: %s", msg)
		received := fmt.Sprintf("Received: %s", jdata.Message)
		printErr(t, tname, msg, expected, received)
	}
}

func TestTools_ErrorJSON(t *testing.T) {
	tname := "Error JSON"
	var testTool Tools

	rec := httptest.NewRecorder()
	err := testTool.ErrorJSON(rec, errors.New("What you seek is not here!"), http.StatusServiceUnavailable)
	// unable to write JSON
	if err != nil {
		msg := "Failed to write JSON"
		info := fmt.Sprintf("Error: %s", err.Error())
		printErr(t, tname, msg, info)
	}

	// decode JSON and check for correctness of values received
	var jdata JSONResponse
	err = json.NewDecoder(rec.Body).Decode(&jdata)
	// unable to decode JSON
	if err != nil {
		msg := "Invalid JSON format of response"
		info := fmt.Sprintf("Error: %s", err.Error())
		printErr(t, tname, msg, info)
	}

	// error value 'false'
	if !jdata.Error {
		msg := `Invalid "error" value`
		expected := `Expected: {"error": "true"}`
		received := `Received: {"error": "false"}`
		printErr(t, tname, msg, expected, received)
	}

	// wrong status code
	if rec.Code != http.StatusServiceUnavailable {
		msg := "Wrong status code"
		expected := fmt.Sprintf("Expected: %d", http.StatusServiceUnavailable)
		received := fmt.Sprintf("Received: %d", rec.Code)
		printErr(t, tname, msg, expected, received)
	}
}
