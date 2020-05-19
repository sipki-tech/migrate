package fs_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ZergsLaw/zerg-repo/cli/core"
	"github.com/ZergsLaw/zerg-repo/cli/fs"
	"github.com/stretchr/testify/assert"
)

var (
	firstMigrate = core.Migrate{
		Version: 1,
		Query: core.Query{
			Up: `create table first_table
(
    id       serial,
    username text not null,
    unique (username),
    primary key (id)
);`,
			Down: `drop table first_table;`,
		},
	}

	secondMigrate = core.Migrate{
		Version: 2,
		Query: core.Query{
			Up: `create table second_table
(
    id       serial,
    username text not null,
    unique (username),
    primary key (id)
);`,
			Down: `drop table second_table;`,
		},
	}

	thirdMigrate = core.Migrate{
		Version: 3,
		Query: core.Query{
			Up: `create table third_table
(
    id       serial,
    username text not null,
    unique (username),
    primary key (id)
);`,
			Down: `drop table third_table;`,
		},
	}
)

func TestFiles_Smoke(t *testing.T) {
	t.Parallel()

	fourMigrate := core.Migrate{
		Version: 4,
		Query: core.Query{
			Up:   `create table test();`,
			Down: `drop table test;`,
		},
	}

	expected := []core.Migrate{firstMigrate, secondMigrate, thirdMigrate, fourMigrate}
	// for remove "\n" from line.
	for i := range expected {
		expected[i].Query.Up = strings.ReplaceAll(expected[i].Query.Up, "\n", "")
	}

	fileSystem := fs.New()
	err := fileSystem.CreateMigrate("testdata/", "test_table", fourMigrate)
	assert.Nil(t, err)

	defer func() {
		assert.Nil(t, os.Remove("testdata/4_test_table.sql"))
	}()

	res, err := fileSystem.Walk("testdata/")
	assert.Nil(t, err)

	assert.Equal(t, expected, res)
}
