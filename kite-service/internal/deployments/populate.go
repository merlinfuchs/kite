package deployments

import (
	"context"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
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
			config := module.ModuleConfig{
				MemoryPagesLimit:   1024,
				TotalTimeLimit:     time.Second * 10,
				ExecutionTimeLimit: time.Millisecond * 10,
			}

			env := host.NewEnv(m.envStores)
			env.DeploymentID = row.ID
			env.GuildID = row.GuildID

			deployments[i] = engine.NewDeployment(row.ID, env, m.compilationCache, row.WasmBytes, row.Manifest, config)
		}

		m.engine.ReplaceGuildDeployments(guildID, deployments)
	}
}
