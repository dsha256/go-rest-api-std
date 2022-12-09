package main

import "regexp"

// match returns true if path matches the regex pattern, and binds any
// capturing groups in pattern to the vars.
func match(path string, pattern *regexp.Regexp, vars ...*string) bool {
	matches := pattern.FindStringSubmatch(path)
	if len(matches) <= 0 {
		return false
	}
	for i, match := range matches[1:] {
		*vars[i] = match
	}
	return true
}
