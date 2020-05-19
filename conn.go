package zergrepo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Config contains basic data for connecting to the database.
// The basic tags for convenient data collection are placed.
type Config struct {
	Host     string `json:"host" yaml:"host" toml:"host" dsn:"host"`
	Port     int    `json:"port" yaml:"port" toml:"port" dsn:"port"`
	User     string `json:"user" yaml:"user" toml:"user" dsn:"user"`
	Password string `json:"password" yaml:"password" toml:"password" dsn:"password"`
	DBName   string `json:"db_name" yaml:"db_name" toml:"db_name" dsn:"dbname"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode" toml:"ssl_mode" dsn:"sslmode"`
}

// Default values.
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "postgres"
	DBPassword = "postgres"
	DBName     = "postgres"
	DBSSLMode  = "disable"
)

// DefaultConfig create instance Config by default data.
func DefaultConfig() *Config {
	return &Config{
		Host:     DBHost,
		Port:     DBPort,
		User:     DBUser,
		Password: DBPassword,
		DBName:   DBName,
		SSLMode:  DBSSLMode,
	}
}

// Cfg returns a formatted string in Cfg format.
func (c Config) DSN() string {
	const tagName = `dsn`

	v := reflect.ValueOf(c)
	t := reflect.TypeOf(c)
	dsn := make([]string, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		fieldInfo := t.Field(i)
		fieldVal := v.Field(i)

		value := fieldInfo.Tag.Get(tagName) + "="
		kind := fieldVal.Kind()
		switch kind {
		case reflect.String:
			value += fieldVal.String()
		case reflect.Int:
			value += strconv.Itoa(int(fieldVal.Int()))
		}

		dsn[i] = value
	}

	return strings.Join(dsn, " ")
}

// Connect to connect to the database using default values.
func Connect(ctx context.Context, driver string, options ...Option) (*sqlx.DB, error) {
	cfg := DefaultConfig()
	for i := range options {
		options[i](cfg)
	}

	return ConnectByCfg(ctx, driver, *cfg)
}

// Cfg for convent replace for custom config.
type Cfg interface {
	DSN() string
}

// ConnectByCfg connect to database by config.
func ConnectByCfg(ctx context.Context, driver string, cfg Cfg) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	err = db.PingContext(ctx)
	for err != nil {
		nextErr := db.PingContext(ctx)
		if errors.Is(nextErr, context.DeadlineExceeded) || errors.Is(nextErr, context.Canceled) {
			return nil, fmt.Errorf("db ping: %w", err)
		}

		err = nextErr
	}

	return db, nil
}
