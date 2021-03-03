package core

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NewMigrate create and save new migrate file.
func (c *Core) NewMigrate(_ context.Context, dir, name string) error {
	if strings.Trim(dir, " ") == "" {
		dir = "."
	}

	t := time.Now()

	year, month, day := t.Date()

	dirPath := filepath.Join(dir, strconv.Itoa(year), month.String(), strconv.Itoa(day))

	err := c.fs.Mkdir(dirPath)
	if err != nil {
		return fmt.Errorf("path [%s]: fs make dir err: %w", dirPath, err)
	}

	currentVersion := uint(0)
	// collect migrates
	var migrates []Migrate
	err = c.fs.Walk(dir, func(path string, info fs.FileInfo) (err error) {
		if info.IsDir() {
			return nil
		}

		m, err := c.walkCb(path, info)
		if err != nil {
			return fmt.Errorf("path [%s]: call walk callback: %w", path, err)
		}

		if currentVersion < m.Version {
			currentVersion = m.Version
		}

		migrates = append(migrates, *m)

		return nil
	})
	if err != nil {
		return fmt.Errorf("path [%s]: fs walk: %w", dirPath, err)
	}

	// build new migrate data
	sort.Slice(migrates, func(i, j int) bool {
		return migrates[i].Version < migrates[j].Version
	})

	const (
		exampleUp   = `create table test();`
		exampleDown = `drop table test;`
	)

	m := Migrate{
		Version: currentVersion + 1,
		Query: Query{
			Up:   exampleUp,
			Down: exampleDown,
		},
	}

	// save new migrate file
	const ext = `.sql`
	migrateName := strings.Join([]string{strconv.Itoa(int(m.Version)), name + ext}, "_")

	buf, err := marshal(m.Query)
	if err != nil {
		return fmt.Errorf("marshal query: %w", err)
	}

	path := filepath.Join(dirPath, migrateName)
	err = c.fs.SaveFile(path, buf)
	if err != nil {
		return fmt.Errorf("path [%s]: save migrate file: %w", path, err)
	}

	c.logger.Info().Msgf("create migrate file by path '%s'", path)

	return nil
}

func (c *Core) walkCb(path string, info fs.FileInfo) (*Migrate, error) {
	file, err := c.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("path [%s]: fs open: %w", path, err)
	}
	// TODO: add error handler.
	defer file.Close()

	q, err := parse(file)
	if err != nil {
		return nil, fmt.Errorf("path [%s]: unmrshal query; %w", path, err)
	}

	version, err := version(info.Name())
	if err != nil {
		return nil, fmt.Errorf("path [%s]: get version: %w", path, err)
	}

	return &Migrate{
		Version: version,
		Query:   *q,
	}, nil
}

func version(name string) (uint, error) {
	slice := strings.Split(name, "_")

	version, err := strconv.Atoi(slice[0])
	if err != nil {
		return 0, fmt.Errorf("parse version: %w", err)
	}

	return uint(version), nil
}
