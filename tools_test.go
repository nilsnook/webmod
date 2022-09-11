package webmod

import (
	"fmt"
	"testing"
)

func printErr(t *testing.T, test, msg string, more ...string) {
	estr := fmt.Sprintf("\nTest: %s:\n\t%s", test, msg)
	for _, m := range more {
		estr = fmt.Sprintf("%s\n%s", estr, m)
	}
	t.Error(estr)
}
