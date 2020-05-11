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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// UP only one migrate.
	err := Repo.Up(ctx, migrateUser)
	assert.Nil(t, err)

	// Down all.
	err = Repo.Down(ctx, migrates...)
	assert.Nil(t, err)

	err = Repo.Up(ctx, migrates...)
	assert.Nil(t, err)

	err = Repo.Down(ctx, migrates...)
	assert.Nil(t, err)
}
