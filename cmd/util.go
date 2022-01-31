package cmd

import (
	"fmt"
	"regexp"
)

// commandRegex is a regex that all command names and parameters must match to.
var commandRegex = regexp.MustCompile("^[\\w-]{1,32}$")

// mustMatch checks if the provided string matches commandRegex. If this is not the case, the function will panic. This
// is used to check command names and parameters, which must be lowercase.
func mustMatch(param string) {
	if !commandRegex.MatchString(param) {
		panic(fmt.Sprintf("string %s does not match expected pattern", param))
	}
}
