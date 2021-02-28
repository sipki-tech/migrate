package core

import (
	"bytes"
	"fmt"
)

// TODO: put in order.
// TODO: very fragile code, fix it.

func marshal(q Query) ([]byte, error) {
	buf := bytes.Buffer{}

	for _, str := range []string{up, q.Up, "\n", down, q.Down} {
		_, err := buf.WriteString(str + "\n")
		if err != nil {
			return nil, fmt.Errorf("write string '%s': %w", str, err)
		}
	}

	return buf.Bytes(), nil
}
