package zergrepo_test

import (
	"database/sql"
	"errors"
	"testing"

	zergrepo "github.com/ZergsLaw/zerg-repo"
	"github.com/stretchr/testify/assert"
)

func TestMapper_Map(t *testing.T) {
	t.Parallel()

	var (
		ErrNotFound = errors.New("not found")
	)

	mapper := zergrepo.NewMapper(zergrepo.NewConvert(ErrNotFound, sql.ErrNoRows))

	testCases := map[string]struct {
		param    error
		expected error
	}{
		"err_not_found": {sql.ErrNoRows, ErrNotFound},
		"err_nil":       {nil, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			err := mapper.Map(tc.param)
			assert.Equal(t, tc.expected, err)
		})
	}
}
