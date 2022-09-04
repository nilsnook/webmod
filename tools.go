package webmod

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module.
// Any variable of this type will have access to all
// the methods with the receiver *Tools.
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString returns a string of random characters of length n,
// using `randomStringSource` as the source for the string.
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}

// UploadedFile is a struct used to save information
// about an uploaded file
type UploadedFile struct {
	FileName         string
	OriginalFileName string
	FileSize         int64
}

// UploadOneFile is just a convenience method that calls UploadFiles, but expects only one file.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	uploadedFiles, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}

	return uploadedFiles[0], nil
}

// UploadFiles uploads one or more file to a specified directory, and gives the files a random name.
// It returns a slice containing the newly named files, the original file name, the size of the file
// and potentially an error. If the optional last parameter isn't set to true, then we will not rename
// the files, but will use the original file names.
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}
	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("The uploaded file is too big")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, fheader := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile

				infile, err := fheader.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				// Check if the filetype is permitted,
				// In order to do that we have to read first 512 bytes of this file to figure out its mimetype
				// and then subsequently check if the filetype is permitted
				buf := make([]byte, 512)
				_, err = infile.Read(buf)
				if err != nil {
					return nil, err
				}
				// check to see if the file type is permitted
				allowed := false
				filetype := http.DetectContentType(buf)
				if len(t.AllowedFileTypes) > 0 {
					for _, aType := range t.AllowedFileTypes {
						if strings.EqualFold(filetype, aType) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}
				if !allowed {
					return nil, errors.New("The uploaded file type is not permitted")
				}

				// If filetype is allowed, go back to beginning of the file
				// because now we need to WRITE it
				// since we moved ahead by 512 bytes to read filetype
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}
				// rename file if opted for
				// here we generate a random string of 25 chars
				// with same extension as the original one.
				if renameFile {
					uploadedFile.FileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(fheader.Filename))
				} else {
					uploadedFile.FileName = fheader.Filename
				}
				uploadedFile.OriginalFileName = fheader.Filename
				// write file
				var outfile *os.File
				defer outfile.Close()
				if outfile, err := os.Create(filepath.Join(uploadDir, uploadedFile.FileName)); err != nil {
					return nil, err
				} else {
					filesize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = filesize
				}

				// Append to the list of uploaded files
				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	// fmt.Printf("%d files uploaded\n", len(uploadedFiles))
	return uploadedFiles, nil
}
