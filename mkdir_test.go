package webmod

import (
	"os"
	"testing"
)

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTool Tools
	testdir := "./testdata/testdir"

	// Creating a dir that does not exists
	err := testTool.CreateDirIfNotExists(testdir)
	if err != nil {
		printErr(t, "Create Dir", err.Error())
	}

	// Following dir already exists, nothing should be done
	err = testTool.CreateDirIfNotExists(testdir)
	if err != nil {
		printErr(t, "Create Dir which already exists", err.Error())
	}

	// Clean up
	os.Remove(testdir)
}
