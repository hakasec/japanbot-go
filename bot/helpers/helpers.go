package helpers

import "errors"

// CreateNgrams takes a string and returns a list of grams of a given size
func CreateNgrams(s string, size int) []string {
	runes := []rune(s)
	if size < 1 {
		err := errors.New("size must be 1 or more")
		panic(err)
	} else if size == len(runes) {
		return []string{s}
	}

	var result []string
	for i := 0; i < len(runes); i++ {
		if i+size > len(runes) {
			break
		}
		gram := string(runes[i : i+size])
		if !StringSliceContains(result, gram) {
			result = append(result, gram)
		}
	}
	return result
}

// StringSliceContains scans a slice for an element
func StringSliceContains(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}
