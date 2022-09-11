package webmod

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

var downloadTest = struct {
	name               string
	dir                string
	file               string
	displayName        string
	contentLength      string
	contentDisposition string
	errorExpected      bool
}{
	name:               "Download file",
	dir:                "./testdata",
	file:               "red.jpg",
	displayName:        "wall.jpg",
	contentLength:      "1107051",
	contentDisposition: "attachment; filename=\"wall.jpg\"",
	errorExpected:      false,
}

func TestTools_DownloadStaticFile(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	var testTool Tools
	var dt = downloadTest
	testTool.DownloadStaticFile(w, r, dt.dir, dt.file, dt.displayName)
	res := w.Result()
	defer res.Body.Close()

	contentLength := res.Header["Content-Length"][0]
	if contentLength != dt.contentLength {
		tname := "Content Length"
		msg := "Invalid content length"
		expected := fmt.Sprintf("Expected: %s", dt.contentLength)
		received := fmt.Sprintf("Received: %s", contentLength)
		printErr(t, tname, msg, expected, received)
	}

	contentDisposition := res.Header["Content-Disposition"][0]
	if contentDisposition != dt.contentDisposition {
		tname := "Content Disposition"
		msg := "Wrong content disposition"
		expected := fmt.Sprintf("Expected: %s", dt.contentDisposition)
		received := fmt.Sprintf("Received: %s", contentDisposition)
		printErr(t, tname, msg, expected, received)
	}

	_, err := ioutil.ReadAll(res.Body)
	if !dt.errorExpected && err != nil {
		tname := "Read file"
		printErr(t, tname, err.Error())
	}
}
