package deployments

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
	"github.com/tetratelabs/wazero"
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
		manifests := make([]*manifest.Manifest, len(rows))
		for i, row := range rows {
			manifests[i] = &row.Manifest

			config := deploymentConfigFromLimits(m.limits, m.compilationCache)

			env := host.NewEnv(m.envStores)
			env.DeploymentID = row.ID
			env.GuildID = row.GuildID
			env.Manifest = row.Manifest

			deployments[i] = engine.NewDeployment(row.ID, env, row.WasmBytes, row.Manifest, config)

			if !row.DeployedAt.Valid || row.UpdatedAt.After(row.DeployedAt.Time) {
				hasChanged = true
			}
		}

		m.engine.ReplaceGuildDeployments(ctx, distype.Snowflake(guildID), deployments)

		if hasChanged {
			err := m.deployManifestChanges(guildID, manifests)
			if err != nil {
				slog.With(logattr.Error(err)).Error("Error deploying manifest changes")
				continue
			}
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
				config := deploymentConfigFromLimits(m.limits, m.compilationCache)

				env := host.NewEnv(m.envStores)
				env.DeploymentID = row.ID
				env.GuildID = row.GuildID
				env.Manifest = row.Manifest

				deployment := engine.NewDeployment(row.ID, env, row.WasmBytes, row.Manifest, config)
				m.engine.LoadGuildDeployment(ctx, distype.Snowflake(guildID), deployment)

				hasChanged = true
			}
		}

		removed := m.engine.TruncateGuildDeployments(ctx, distype.Snowflake(guildID), deploymentIDs)
		if removed {
			hasChanged = true
		}

		if hasChanged {
			err := m.deployManifestChanges(guildID, manifests)
			if err != nil {
				slog.With(logattr.Error(err)).Error("Error deploying manifest changes")
				continue
			}
		}

		err = m.store.UpdateDeploymentsDeployedAtForGuild(ctx, guildID, time.Now().UTC())
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error updating deployments deployed at for guild")
			continue
		}
	}

	return nil
}

func deploymentConfigFromLimits(limits config.ServerEngineLimitConfig, compilationCache wazero.CompilationCache) engine.DeploymentConfig {
	return engine.DeploymentConfig{
		ModuleConfig: module.ModuleConfig{
			MemoryPagesLimit:   limits.MaxMemoryPages,
			TotalTimeLimit:     time.Duration(limits.MaxTotalTime) * time.Millisecond,
			ExecutionTimeLimit: time.Duration(limits.MaxExecutionTime) * time.Millisecond,
			CompilationCache:   compilationCache,
		},
		PoolMaxTotal: limits.DeploymentPoolMaxTotal,
		PoolMaxIdle:  limits.DeploymentPoolMaxIdle,
		PoolMinIdle:  limits.DeploymentPoolMinIdle,
	}
}

func (m *DeploymentManager) deployManifestChanges(guildID string, manifests []*manifest.Manifest) error {
	fmt.Println("deploy manifest changes")
	return nil
}
