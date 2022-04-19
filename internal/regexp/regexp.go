// Package regexp provides a global regexp cache via go-regexpcache
package regexp

import (
	"regexp"

	"github.com/umisama/go-regexpcache"
)

// Compile wraps regexpcache.Compile
func Compile(s string) (*regexp.Regexp, error) {
	return regexpcache.Compile(s)
}

// MustCompile wraps regexpcache.MustCompile
func MustCompile(s string) *regexp.Regexp {
	return regexpcache.MustCompile(s)
}
