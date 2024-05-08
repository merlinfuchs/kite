package deployments

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/rest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/tetratelabs/wazero"
)

func (m *DeploymentManager) populateEngineDeployments(ctx context.Context) error {
	appIDs, err := m.store.GetAppIDsWithDeployment(ctx)
	if err != nil {
		return fmt.Errorf("error getting app ids with deployment: %w", err)
	}

	for _, appID := range appIDs {
		rows, err := m.store.GetDeploymentsForApp(ctx, appID)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error getting deployments for app")
			continue
		}

		hasChanged := false

		deployments := make([]*engine.Deployment, len(rows))
		manifests := make([]manifest.Manifest, len(rows))
		for i, row := range rows {
			for _, command := range row.Manifest.DiscordApplicationCommands() {
				fmt.Println(command.Name)
			}

			manifests[i] = row.Manifest

			config := deploymentConfigFromLimits(m.limits, m.compilationCache)

			env := host.NewEnv(m.envStores, row.AppID, row.ID)
			env.Manifest = row.Manifest

			deployments[i] = engine.NewDeployment(row.ID, env, row.WasmBytes, row.Manifest, config)

			if !row.DeployedAt.Valid || row.UpdatedAt.After(row.DeployedAt.Time) {
				hasChanged = true
			}
		}

		m.engine.ReplaceAppDeployments(ctx, distype.Snowflake(appID), deployments)

		if hasChanged {
			err := m.deployManifestChanges(ctx, appID, manifests)
			if err != nil {
				slog.With(logattr.Error(err)).Error("Error deploying manifest changes")
				continue
			}
		}

		err = m.store.UpdateDeploymentsDeployedAtForApp(ctx, appID, time.Now().UTC())
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error updating deployments deployed at for app")
			continue
		}
	}

	return nil
}

func (m *DeploymentManager) updateEngineDeployments(ctx context.Context) error {
	appIDs, err := m.store.GetAppIDsWithDeployment(ctx)
	if err != nil {
		return fmt.Errorf("error getting app ids with deployment: %w", err)
	}

	for _, appID := range appIDs {
		rows, err := m.store.GetDeploymentsForApp(ctx, appID)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error getting deployments for app")
			continue
		}

		hasChanged := false

		deploymentIDs := make([]string, len(rows))
		manifests := make([]manifest.Manifest, len(rows))
		for i, row := range rows {
			deploymentIDs[i] = row.ID
			manifests[i] = row.Manifest

			if !row.DeployedAt.Valid || row.UpdatedAt.After(row.DeployedAt.Time) {
				config := deploymentConfigFromLimits(m.limits, m.compilationCache)

				// TODO: get right app state from manager
				env := host.NewEnv(m.envStores, row.AppID, row.ID)
				env.Manifest = row.Manifest

				deployment := engine.NewDeployment(row.ID, env, row.WasmBytes, row.Manifest, config)
				m.engine.LoadAppDeployment(ctx, distype.Snowflake(appID), deployment)

				hasChanged = true
			}
		}

		removed := m.engine.TruncateAppDeployments(ctx, distype.Snowflake(appID), deploymentIDs)
		if removed {
			hasChanged = true
		}

		if hasChanged {
			err := m.deployManifestChanges(ctx, appID, manifests)
			if err != nil {
				slog.With(logattr.Error(err)).Error("Error deploying manifest changes")
				continue
			}
		}

		err = m.store.UpdateDeploymentsDeployedAtForApp(ctx, appID, time.Now().UTC())
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error updating deployments deployed at for app")
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
			HostCallLimit:      limits.MaxHostCalls,
			CompilationCache:   compilationCache,
		},
		PoolMaxTotal: limits.DeploymentPoolMaxTotal,
		PoolMaxIdle:  limits.DeploymentPoolMaxIdle,
		PoolMinIdle:  limits.DeploymentPoolMinIdle,
	}
}

func (m *DeploymentManager) deployManifestChanges(ctx context.Context, appID string, manifests []manifest.Manifest) error {
	fmt.Println("deploy manifest changes")

	discordCommands := []distype.ApplicationCommandCreateRequest{}
	for _, manifest := range manifests {
		discordCommands = append(discordCommands, manifest.DiscordApplicationCommands()...)
	}

	app, err := m.apps.AppState(distype.Snowflake(appID))
	if err != nil {
		if err == store.ErrNotFound {
			return fmt.Errorf("app state not found for app id: %s", appID)
		}
		return fmt.Errorf("error getting app state: %w", err)
	}

	err = app.Client().Request(
		rest.SetGlobalCommands.Compile(nil, appID), discordCommands, nil,
		rest.WithCtx(ctx),
	)
	if err != nil {
		derr := err.(rest.Error)
		fmt.Println(string(derr.RsBody))
		return fmt.Errorf("error setting app commands: %w", err)
	}

	return nil
}
