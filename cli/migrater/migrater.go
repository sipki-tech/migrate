// Package migrater contains logic for migrate data in database.
package migrater

import (
	"context"
	"fmt"

	zergrepo "github.com/ZergsLaw/zerg-repo"
	"github.com/ZergsLaw/zerg-repo/cli/core"
)

// Migrater is responsible for data migration to the database.
type Migrater struct {
	repo *zergrepo.Repo
}

// New create new instance migrater.
func New(r *zergrepo.Repo) core.Migrater {
	return &Migrater{r}
}

// Migrate sql requests.
func (m *Migrater) Migrate(ctx context.Context, cfg core.Config, migrates []core.Migrate) error {
	zergMigrates := make([]zergrepo.Migrate, len(migrates))
	for i := range migrates {
		zergMigrates[i] = zergrepo.Migrate{
			Version: migrates[i].Version,
			Up:      zergrepo.Query(migrates[i].Query.Up),
			Down:    zergrepo.Query(migrates[i].Query.Down),
		}
	}

	err := zergrepo.RegisterMetric(zergMigrates...)
	if err != nil {
		return err
	}

	switch cfg.Cmd {
	case core.Up:
		return m.repo.Up(ctx)
	case core.UpOne:
		return m.repo.UpOne(ctx)
	case core.UpTo:
		return m.repo.UpTo(ctx, cfg.To)
	case core.Down:
		return m.repo.Down(ctx)
	case core.DownTo:
		return m.repo.DownTo(ctx, cfg.To)
	case core.Reset:
		return m.repo.Reset(ctx)
	default:
		panic(fmt.Sprintf("unknown cmd: %d", cfg.Cmd))
	}
}
