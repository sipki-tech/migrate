package sql

import (
	"errors"
	"reflect"
	"strings"
)

// TODO: put in order.
// TODO: very fragile code, fix it.

// Errors.
var (
	ErrBlockNotFound = errors.New("sql block not found")
)

// Unmarshal parses the sql-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, val interface{}) error {
	fields := strings.Split(string(data), separator)
	// block-name -> sql query
	maps := make(map[string]string)

	for _, field := range fields {
		if field == "" {
			continue
		}
		slice := strings.Split(strings.Trim(field, "\n"), "\n")
		maps[slice[0]] = slice[1]
	}

	v := reflect.ValueOf(val)
	elem := v.Elem()
	t := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		tagName, ok := t.Field(i).Tag.Lookup(tag)
		if !ok {
			continue
		}

		value, ok := maps[tagName]
		if !ok {
			return ErrBlockNotFound
		}

		elem.Field(i).SetString(value)
	}

	return nil
}
