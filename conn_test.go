package zergrepo_test

import (
	"fmt"
	"testing"

	zergrepo "github.com/ZergsLaw/zerg-repo"
	"github.com/stretchr/testify/assert"
)

func TestConfig_DSN(t *testing.T) {
	t.Parallel()

	cfg := zergrepo.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DBName:   "postgres",
		SSLMode:  "disable",
	}

	expected := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	dsn := cfg.DSN()

	assert.Equal(t, expected, dsn)
}
