package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Meat-Hook/migrate/cmd"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var application = &cli.App{
	Name:         filepath.Base(os.Args[0]),
	HelpName:     filepath.Base(os.Args[0]),
	Usage:        "Migration migrate.",
	Commands:     []*cli.Command{cmd.NewMigrate, cmd.Version, cmd.Run},
	BashComplete: cli.DefaultAppComplete,
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)
	go func() { <-signals; cancel() }()
	go forceShutdown(ctx)

	output := zerolog.NewConsoleWriter()
	logger := zerolog.New(output).With().Timestamp().Logger()

	err := application.RunContext(logger.WithContext(ctx), os.Args)
	if err != nil {
		logrus.Fatalf("failed: %s", err)
	}
}

func forceShutdown(ctx context.Context) {
	const shutdownDelay = 9 * time.Second

	<-ctx.Done()
	time.Sleep(shutdownDelay)
	log.Fatal("failed to shutdown")
}
