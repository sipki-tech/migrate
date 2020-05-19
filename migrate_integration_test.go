// +build integration

package zergrepo_test

import (
	"context"
	"testing"

	zergrepo "github.com/ZergsLaw/zerg-repo"
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
	migrateUser = zergrepo.Migrate{
		Version: 1,
		Up:      zergrepo.Query(upUser),
		Down:    zergrepo.Query(downUser),
	}

	migrateProduct = zergrepo.Migrate{
		Version: 2,
		Up:      zergrepo.Query(upProduct),
		Down:    zergrepo.Query(downProduct),
	}
)

func TestRepo_UpAndDownSmoke(t *testing.T) {
	t.Parallel()

	migrates := []zergrepo.Migrate{migrateUser, migrateProduct}
	// Register you migrate.
	err := zergrepo.RegisterMetric(migrates...)
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), timeout*1000)
	defer cancel()

	// Migration to a specific version.
	err = Repo.UpTo(ctx, 1)
	assert.Nil(t, err)
	// Starting migration of the next version.
	err = Repo.UpOne(ctx)
	assert.Nil(t, err)
	// Rollback to a specific version.
	err = Repo.DownTo(ctx, 2)
	assert.Nil(t, err)
	// Rollback current migration.
	err = Repo.Down(ctx)
	assert.Nil(t, err)
	// Up all migration.
	err = Repo.Up(ctx)
	assert.Nil(t, err)
	// Rollback all migration.
	err = Repo.Reset(ctx)
	assert.Nil(t, err)
}
