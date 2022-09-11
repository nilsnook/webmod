package webmod

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{
		name:          "Allowed no rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    false,
		errorExpected: false,
	},
	{
		name:          "Allowed rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    true,
		errorExpected: false,
	},
	{
		name:          "Not allowed",
		allowedTypes:  []string{"image/jpeg"},
		renameFile:    false,
		errorExpected: true,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// set up a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)

		// writing multipart file data to pipe concurrently
		// simulating a file upload from http form
		go func(tname string) {
			defer writer.Close()
			defer wg.Done()

			// create http form data field "file"
			testFile := "./testdata/img.png"
			part, err := writer.CreateFormFile("file", testFile)
			if err != nil {
				printErr(t, tname, err.Error())
			}

			f, err := os.Open(testFile)
			if err != nil {
				printErr(t, tname, err.Error())
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				info := fmt.Sprintf("Error: %s", err.Error())
				printErr(t, tname, "Error decoding image", info)
			}

			// writing png to pipe
			err = png.Encode(part, img)
			if err != nil {
				printErr(t, tname, err.Error())
			}
		}(e.name)

		// read from the pipe
		r := httptest.NewRequest(http.MethodPost, "/", pr)
		r.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadDir := "./testdata/uploads/"
		uploadedFiles, err := testTools.UploadFiles(r, uploadDir, e.renameFile)

		// There is an error but we don't expect one!
		if err != nil && !e.errorExpected {
			printErr(t, e.name, err.Error())
		}

		// There was no error and we don't expect one
		// so check if the file exists at the uploaded file path
		if !e.errorExpected {
			uploadedFilePath := fmt.Sprintf("%s%s", uploadDir, uploadedFiles[0].FileName)

			// check if the uploaded file actually exists
			if _, err := os.Stat(uploadedFilePath); os.IsNotExist(err) {
				msg := fmt.Sprintf("Expected file %s to exist at %s dir", e.name, uploadDir)
				info := fmt.Sprintf("Error: %s", err.Error())
				printErr(t, e.name, msg, info)
			}

			// clean up
			os.Remove(uploadedFilePath)
		}

		// We expect an error, but none received!
		if e.errorExpected && err == nil {
			printErr(t, e.name, "Expected error, but none received")
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	tname := "Upload one file"

	// set up a pipe to avoid buffering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// writing multipart file data to pipe concurrently
	// simulating a file upload from http form
	go func() {
		defer writer.Close()

		// create http form data field "file"
		testFile := "./testdata/img.png"
		part, err := writer.CreateFormFile("file", testFile)
		if err != nil {
			printErr(t, tname, err.Error())
		}

		f, err := os.Open(testFile)
		if err != nil {
			printErr(t, tname, err.Error())
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			info := fmt.Sprintf("Error: %s", err.Error())
			printErr(t, tname, "Error decoding image", info)
		}

		// writing png to pipe
		err = png.Encode(part, img)
		if err != nil {
			printErr(t, tname, err.Error())
		}
	}()

	// read from the pipe
	r := httptest.NewRequest(http.MethodPost, "/", pr)
	r.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools
	// testTools.AllowedFileTypes = e.allowedTypes

	uploadDir := "./testdata/uploads/"
	uploadedFile, err := testTools.UploadOneFile(r, uploadDir, true)

	// There is an error but we don't expect one!
	if err != nil {
		printErr(t, tname, err.Error())
	}

	// There was no error
	// so check if the file exists at the uploaded file path
	uploadedFilePath := fmt.Sprintf("%s%s", uploadDir, uploadedFile.FileName)
	// check if the uploaded file actually exists
	if _, err := os.Stat(uploadedFilePath); os.IsNotExist(err) {
		msg := fmt.Sprintf("Expected file %s to exist at %s dir", uploadedFile.FileName, uploadDir)
		info := fmt.Sprintf("Error: %s", err.Error())
		printErr(t, tname, msg, info)
	}
	// clean up
	os.Remove(uploadedFilePath)
}
