// +build integration

package zergrepo_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	zergrepo "github.com/ZergsLaw/zerg-repo"
	"go.uber.org/zap"
)

var (
	Repo = &zergrepo.Repo{}

	timeout = time.Second * 5
)

func TestMain(m *testing.M) {
	cfg := zergrepo.DefaultConfig()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := zergrepo.ConnectByCfg(ctx, "postgres", cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("connect db: %w", err))
	}

	l, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(fmt.Errorf("init zap: %w", err))
	}

	metric := zergrepo.MustMetric("test", "test")
	mapper := zergrepo.NewMapper()

	Repo = zergrepo.New(db, l.Named("zergrepo").Sugar(), metric, mapper)

	os.Exit(m.Run())
}
