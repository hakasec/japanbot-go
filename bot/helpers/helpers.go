package helpers

import (
	"errors"
	"reflect"
	"strings"
)

const SqliteDateFormat = "2006-01-02 15:04:05.999999999-07:00"

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

// GetSQLType converts a reflect Type to a SQL type string
func GetSQLType(T reflect.Type) string {
	switch T.Kind() {
	case reflect.String:
		return "TEXT"
	case reflect.Int:
		return "INTEGER"
	case reflect.Float64:
		fallthrough
	case reflect.Float32:
		return "REAL"
	default:
		return "TEXT"
	}
}

// GetModelFields returns a list of StructFields with the model tag
func GetModelFields(T reflect.Type) []*reflect.StructField {
	var fields []*reflect.StructField

	for i := 0; i < T.NumField(); i++ {
		field := T.Field(i)
		_, ok := field.Tag.Lookup("model")
		if ok {
			fields = append(fields, &field)
		}
	}
	return fields
}

func GetNumDigits(num int) int {
	digits := 0
	for num > 0 {
		digits++
		num = int(num / 10)
	}
	return digits
}

func IsDigits(s string) bool {
	for _, c := range s {
		if !strings.ContainsAny(string(c), "0123456789") {
			return false
		}
	}
	return true
}
