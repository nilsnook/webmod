package webmod

import (
	"fmt"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	tname := "Random String (length=10)"
	strlen := 10

	var testTools Tools

	s := testTools.RandomString(strlen)
	if len(s) != strlen {
		expected := fmt.Sprintf("Expected length: %d", strlen)
		received := fmt.Sprintf("Received length: %d", len(s))
		printErr(t, tname, "Wrong length random string returned", expected, received)
	}
}
