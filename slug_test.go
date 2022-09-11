package webmod

import (
	"fmt"
	"testing"
)

var slugTests = []struct {
	name          string
	str           string
	expected      string
	errorExpected bool
}{
	{
		name:          "Valid string",
		str:           "now is the time!",
		expected:      "now-is-the-time",
		errorExpected: false,
	},
	{
		name:          "Empty string",
		str:           "",
		expected:      "",
		errorExpected: true,
	},
	{
		name:          "Complex string",
		str:           "yo! ^now^ is the f**king time. 123... Go!",
		expected:      "yo-now-is-the-f-king-time-123-go",
		errorExpected: false,
	},
	{
		name:          "Japanese string",
		str:           "今がその時だ",
		expected:      "",
		errorExpected: true,
	},
	{
		name:          "Japanese string and Roamn characters",
		str:           "今がその時だ! GO GET THEM-->",
		expected:      "go-get-them",
		errorExpected: false,
	},
}

func TestTools_Slugify(t *testing.T) {
	var testTool Tools

	for _, e := range slugTests {
		slug, err := testTool.Slugify(e.str)
		if !e.errorExpected && err != nil {
			info := fmt.Sprintf("Error: %s", err.Error())
			printErr(t, e.name, "Error received when none expected!", info)
		}

		if !e.errorExpected && slug != e.expected {
			expected := fmt.Sprintf("Expected: %s", e.expected)
			received := fmt.Sprintf("Received: %s", slug)
			printErr(t, e.name, "Wrong slug returned!", expected, received)
		}

		if e.errorExpected && err == nil {
			printErr(t, e.name, "Error expected, but none received")
		}
	}
}
