// Package regexp provides a global regexp cache via go-regexpcache
package regexp

import (
	"fmt"
	"regexp"

	"github.com/umisama/go-regexpcache"
)

// Compile wraps regexpcache.Compile
func Compile(s string) (*regexp.Regexp, error) {
	r, err := regexpcache.Compile(s)
	if err != nil {
		return nil, fmt.Errorf("regexp pattern %s: %v", s, err)
	}

	return r, nil
}

// MustCompile wraps regexpcache.MustCompile
func MustCompile(s string) *regexp.Regexp {
	return regexpcache.MustCompile(s)
}
