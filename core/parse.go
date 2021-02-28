package core

import (
	"bufio"
	"fmt"
	"io/fs"
)

// TODO: put in order.
// TODO: very fragile code, fix it.

const (
	up   = `--up`
	down = `--down`
)

func parse(f fs.File) (*Query, error) {
	scan := bufio.NewScanner(f)
	q := Query{}
	currentParse := 0
	for scan.Scan() {
		str := scan.Text()
		switch str {
		case up:
			currentParse = 1
			continue
		case down:
			currentParse = 2
			continue
		}

		switch currentParse {
		case 1:
			q.Up += str
		case 2:
			q.Down += str
		}
	}

	err := scan.Err()
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return &q, nil
}
