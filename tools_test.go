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

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("Wrong length random string returned")
	}
}

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
		go func() {
			defer writer.Close()
			defer wg.Done()

			// create http form data field "file"
			testFile := "./testdata/img.png"
			part, err := writer.CreateFormFile("file", testFile)
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open(testFile)
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("Error decoding image", err)
			}

			// writing png to pipe
			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		// read from the pipe
		r := httptest.NewRequest(http.MethodPost, "/", pr)
		r.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadDir := "./testdata/uploads/"
		uploadedFiles, err := testTools.UploadFiles(r, uploadDir, e.renameFile)

		// There is an error but we don't expect one!
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		// There was no error and we don't expect one
		// so check if the file exists at the uploaded file path
		if !e.errorExpected {
			uploadedFilePath := fmt.Sprintf("%s%s", uploadDir, uploadedFiles[0].FileName)

			// check if the uploaded file actually exists
			if _, err := os.Stat(uploadedFilePath); os.IsNotExist(err) {
				t.Errorf("Expected file %s to exist at %s dir", e.name, uploadDir)
			}

			// clean up
			os.Remove(uploadedFilePath)
		}

		// We expect an error, but none received!
		if !e.errorExpected && err != nil {
			t.Errorf("Expected error, but none received for test -> %s", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
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
			t.Error(err)
		}

		f, err := os.Open(testFile)
		if err != nil {
			t.Error(err)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Error("Error decoding image", err)
		}

		// writing png to pipe
		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
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
		t.Error(err)
	}

	// There was no error
	// so check if the file exists at the uploaded file path
	uploadedFilePath := fmt.Sprintf("%s%s", uploadDir, uploadedFile.FileName)
	// check if the uploaded file actually exists
	if _, err := os.Stat(uploadedFilePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist at %s dir", e.name, uploadDir)
	}
	// clean up
	os.Remove(uploadedFilePath)
}
