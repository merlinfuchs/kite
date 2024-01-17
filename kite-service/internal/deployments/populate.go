package deployments

import (
	"context"
	"log/slog"

	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

func (m *DeploymentManager) populateEngineDeployments(ctx context.Context) {
	// TODO: only get guild ids that have a deployment that has updated recently
	guildIDs, err := m.store.GetGuildIDsWithDeployment(ctx)
	if err != nil {
		slog.With(logattr.Error(err)).Error("Error getting guild ids with deployment")
		return
	}

	for _, guildID := range guildIDs {
		rows, err := m.store.GetDeploymentsForGuild(ctx, guildID)
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error getting deployments for guild")
			continue
		}

		deployments := make([]*engine.PluginDeployment, len(rows))
		for i, row := range rows {
			manifest := plugin.Manifest{
				Events: row.ManifestEvents,
				// TODO: Commands: row.ManifestCommands,
			}
			config := plugin.PluginConfig{
				MemoryPagesLimit: 64,
			}

			deployments[i] = engine.NewDeployment(row.WasmBytes, manifest, config)
		}

		m.engine.ReplaceGuildDeployments(guildID, deployments)
	}
}