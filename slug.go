package webmod

import (
	"errors"
	"regexp"
	"strings"
)

// Slugify is a simple means of creating a slug from a string
func (t *Tools) Slugify(s string) (string, error) {
	// If string empty, return error
	if s == "" {
		return "", errors.New("Empty string not permitted")
	}

	var re = regexp.MustCompile(`[^a-z\d]+`)
	slug := strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	// If the string has no characters or digits,
	// the slug length will be zero.
	if len(slug) == 0 {
		return "", errors.New("String contains no letters or digits, slug length is zero")
	}

	return slug, nil
}
