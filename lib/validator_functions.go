package constago

import "regexp"

func isValidRegex(s string) bool {
	_, err := regexp.Compile(s)
	return err == nil
}
