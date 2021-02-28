package sql

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

// TODO: put in order.
// TODO: very fragile code, fix it.

// Errors.
var (
	ErrMustBeStruct = errors.New("must be struct")
)

// Marshal returns sql encoding v.
func Marshal(val interface{}) ([]byte, error) {
	v := reflect.ValueOf(val)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil, ErrMustBeStruct
	}

	buf := bytes.Buffer{}

	for i := 0; i < v.NumField(); i++ {
		fieldT := t.Field(i)
		fieldV := v.Field(i)

		tagName, ok := fieldT.Tag.Lookup(tag)
		if !ok {
			continue
		}

		str := fmt.Sprintf("--%s\n%s", tagName, fieldV.String())

		buf.WriteString(str + "\n")
	}

	return buf.Bytes(), nil
}
