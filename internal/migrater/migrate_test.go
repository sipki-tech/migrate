package migrater_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Meat-Hook/migrate/internal/core"
	"github.com/Meat-Hook/migrate/internal/migrater"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	upUser = `create table users
(
    id         serial,
    username   text                    not null,
    created_at timestamp default now() not null,

    unique (username),
    primary key (id)
)`
	downUser = `drop table users`

	upProduct = `create table product
(
    id         serial,
    user_id    integer                 not null,
    name 	   text                    not null default '',
    created_at timestamp default now() not null,

    foreign key (user_id) references users on delete cascade,
    primary key (id)
)`
	downProduct = `drop table product`
)

var (
	migrateUser = core.Migrate{
		Version: 1,
		Query: core.Query{
			Up:   upUser,
			Down: downUser,
		},
	}

	migrateProduct = core.Migrate{
		Version: 2,
		Query: core.Query{
			Up:   upProduct,
			Down: downProduct,
		},
	}
)

const timeout = time.Second * 1000

func TestRepo_UpAndDownSmoke(t *testing.T) {
	t.Parallel()

	migrates := []core.Migrate{migrateUser, migrateProduct}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db := start(t)
	m := migrater.New(db, logrus.New())

	// Migration to a specific version.
	err := m.UpTo(ctx, 1, migrates...)
	assert.Nil(t, err)
	// Starting migration of the next version.
	err = m.UpOne(ctx, migrates...)
	assert.Nil(t, err)
	// Rollback to a specific version.
	err = m.DownTo(ctx, 2, migrates...)
	assert.Nil(t, err)
	// Rollback current migration.
	err = m.Down(ctx, migrates...)
	assert.Nil(t, err)
	// Up all migration.
	err = m.Up(ctx, migrates...)
	assert.Nil(t, err)
	// Rollback all migration.
	err = m.Reset(ctx, migrates...)
	assert.Nil(t, err)
}

func start(t *testing.T) *sql.DB {
	pool, err := dockertest.NewPool("")
	assert.Nil(t, err)

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "latest", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_DB=postgres",
		"POSTGRES_PASSWORD=postgres",
	})
	assert.Nil(t, err)

	var db *sql.DB
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		str := fmt.Sprintf("host=localhost port=%s user=postgres "+
			"password=postgres dbname=postgres sslmode=disable", resource.GetPort("5432/tcp"))
		db, err = sql.Open("postgres", str)
		if err != nil {
			return err
		}

		return db.Ping()
	}); err != nil {
		assert.Nil(t, err)
	}

	t.Cleanup(func() {
		err = pool.Purge(resource)
		assert.Nil(t, err)
	})

	return db
}
