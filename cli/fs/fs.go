// Package fs is responsible for migrating files.
package fs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ZergsLaw/zerg-repo/cli/core"
)

type fs struct{}

// New create new instance core.Fs.
func New() core.Fs {
	return &fs{}
}

const ext = `.sql`

// Walk searches for all migration files located in the specified directory.
func (fs *fs) Walk(dir string) (migrates []core.Migrate, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ext != filepath.Ext(info.Name()) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer func() {
			if closerErr := f.Close(); closerErr != nil {
				err = fmt.Errorf("%w, err close file: %s", err, closerErr)
			}
		}()

		v, err := version(info.Name())
		if err != nil {
			return err
		}

		q, err := parse(f)
		if err != nil {
			return err
		}

		migrates = append(migrates, core.Migrate{Version: v, Query: *q})
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return migrates, nil
}

// CreateMigrate creates a migration file in the specified directory.
func (fs *fs) CreateMigrate(dir, name string, m core.Migrate) error {
	name = strings.Join([]string{strconv.Itoa(int(m.Version)), name + ext}, "_")

	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	for _, str := range []string{up, m.Query.Up, "\n", down, m.Query.Down} {
		_, err = f.WriteString(str + "\n")
		if err != nil {
			return fmt.Errorf("write string '%s': %w", str, err)
		}
	}

	return nil
}

const (
	up   = `--up`
	down = `--down`
)

func parse(f io.Reader) (*core.Query, error) {
	scan := bufio.NewScanner(f)
	q := core.Query{}
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

func version(name string) (uint, error) {
	slice := strings.Split(name, "_")

	version, err := strconv.Atoi(slice[0])
	if err != nil {
		return 0, fmt.Errorf("parse version: %w", err)
	}

	return uint(version), nil
}
