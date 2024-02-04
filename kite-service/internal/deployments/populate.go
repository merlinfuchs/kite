package deployments

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
	"github.com/merlinfuchs/kite/kite-types/manifest"
)

func (m *DeploymentManager) populateEngineDeployments(ctx context.Context) error {
	guildIDs, err := m.store.GetGuildIDsWithDeployment(ctx)
	if err != nil {
		return fmt.Errorf("error getting guild ids with deployment: %w", err)
	}

	for _, guildID := range guildIDs {
		rows, err := m.store.GetDeploymentsForGuild(ctx, guildID)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error getting deployments for guild")
			continue
		}

		hasChanged := false

		deployments := make([]*engine.Deployment, len(rows))
		for i, row := range rows {
			config := module.ModuleConfig{
				MemoryPagesLimit:   1024,
				TotalTimeLimit:     time.Second * 10,
				ExecutionTimeLimit: time.Millisecond * 100,
				CompilationCache:   m.compilationCache,
			}

			env := host.NewEnv(m.envStores)
			env.DeploymentID = row.ID
			env.GuildID = row.GuildID
			env.Manifest = row.Manifest

			deployments[i] = engine.NewDeployment(row.ID, env, row.WasmBytes, row.Manifest, config)

			if !row.DeployedAt.Valid || row.UpdatedAt.After(row.DeployedAt.Time) {
				hasChanged = true
			}
		}

		m.engine.ReplaceGuildDeployments(ctx, guildID, deployments)

		if hasChanged {
			// TODO: deploy discord commands
			fmt.Println("deploy commands")
		}

		err = m.store.UpdateDeploymentsDeployedAtForGuild(ctx, guildID, time.Now().UTC())
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error updating deployments deployed at for guild")
			continue
		}
	}

	return nil
}

func (m *DeploymentManager) updateEngineDeployments(ctx context.Context) error {
	guildIDs, err := m.store.GetGuildIDsWithDeployment(ctx)
	if err != nil {
		return fmt.Errorf("error getting guild ids with deployment: %w", err)
	}

	for _, guildID := range guildIDs {
		rows, err := m.store.GetDeploymentsForGuild(ctx, guildID)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error getting deployments for guild")
			continue
		}

		hasChanged := false

		deploymentIDs := make([]string, len(rows))
		manifests := make([]*manifest.Manifest, len(rows))
		for i, row := range rows {
			deploymentIDs[i] = row.ID
			manifests[i] = &row.Manifest

			if !row.DeployedAt.Valid || row.UpdatedAt.After(row.DeployedAt.Time) {
				config := module.ModuleConfig{
					MemoryPagesLimit:   1024,
					TotalTimeLimit:     time.Second * 10,
					ExecutionTimeLimit: time.Millisecond * 10,
					CompilationCache:   m.compilationCache,
				}

				env := host.NewEnv(m.envStores)
				env.DeploymentID = row.ID
				env.GuildID = row.GuildID
				env.Manifest = row.Manifest

				deployment := engine.NewDeployment(row.ID, env, row.WasmBytes, row.Manifest, config)
				m.engine.LoadGuildDeployment(ctx, guildID, deployment)

				hasChanged = true
			}
		}

		removed := m.engine.TruncateGuildDeployments(ctx, guildID, deploymentIDs)
		if removed {
			hasChanged = true
		}

		if hasChanged {
			// TODO: deploy discord commands
			fmt.Println("deploy commands 2")
		}

		err = m.store.UpdateDeploymentsDeployedAtForGuild(ctx, guildID, time.Now().UTC())
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error updating deployments deployed at for guild")
			continue
		}
	}

	return nil
}
