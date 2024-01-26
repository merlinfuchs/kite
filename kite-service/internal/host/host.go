package host

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/kvmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/bot"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type HostEnvironment struct {
	bot               *bot.Bot
	deployments       store.DeploymentStore
	deploymentLogs    store.DeploymentLogStore
	deploymentMetrics store.DeploymentMetricStore
	kvStorage         store.KVStorageStore
}

func NewEnv(bot *bot.Bot, deployments store.DeploymentStore, deploymentLogs store.DeploymentLogStore, deploymentMetrics store.DeploymentMetricStore, kvStorage store.KVStorageStore) HostEnvironment {
	return HostEnvironment{
		bot:               bot,
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
		kvStorage:         kvStorage,
	}
}

func (h HostEnvironment) Call(ctx context.Context, deploymentID string, guildID string, req call.Call) (res interface{}, err error) {
	timeout := time.Duration(time.Duration(req.Config.Timeout) * time.Millisecond)
	if timeout == 0 || timeout > 3*time.Second {
		timeout = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()

	switch req.Type {
	case call.Sleep:
		data := req.Data.(call.SleepCall)

		duration := time.Duration(data.Duration) * time.Millisecond
		if duration > 3*time.Second {
			duration = 3 * time.Second
		}

		time.Sleep(duration)
		return call.SleepResponse{}, nil
	case call.KVKeyGet:
		res, err = h.callKVKeyGet(ctx, guildID, req.Data.(kvmodel.KVKeyGetCall))
	case call.KVKeySet:
		res, err = h.callKVKeySet(ctx, guildID, req.Data.(kvmodel.KVKeySetCall))
	case call.KVKeyDelete:
		res, err = h.callKVKeyDelete(ctx, guildID, req.Data.(kvmodel.KVKeyDeleteCall))
	case call.KVKeyIncrease:
		res, err = h.callKVKeyIncrease(ctx, guildID, req.Data.(kvmodel.KVKeyIncreaseCall))
	case call.DiscordBanList:
		res, err = h.callDiscordBanList(ctx, guildID, req.Data.(dismodel.BanListCall))
	case call.DiscordBanGet:
		res, err = h.callDiscordBanGet(ctx, guildID, req.Data.(dismodel.BanGetCall))
	case call.DiscordBanCreate:
		res, err = h.callDiscordBanCreate(ctx, guildID, req.Data.(dismodel.BanCreateCall))
	case call.DiscordBanRemove:
		res, err = h.callDiscordBanRemove(ctx, guildID, req.Data.(dismodel.BanRemoveCall))
	case call.DiscordChannelGet:
		res, err = h.callDiscordChannelGet(ctx, guildID, req.Data.(dismodel.ChannelGetCall))
	case call.DiscordChannelList:
		res, err = h.callDiscordChannelList(ctx, guildID, req.Data.(dismodel.ChannelListCall))
	case call.DiscordChannelCreate:
		res, err = h.callDiscordChannelCreate(ctx, guildID, req.Data.(dismodel.ChannelCreateCall))
	case call.DiscordChannelUpdate:
		res, err = h.callDiscordChannelUpdate(ctx, guildID, req.Data.(dismodel.ChannelUpdateCall))
	case call.DiscordChannelUpdatePositions:
		res, err = h.callDiscordChannelUpdatePositions(ctx, guildID, req.Data.(dismodel.ChannelUpdatePositionsCall))
	case call.DiscordChannelDelete:
		res, err = h.callDiscordChannelDelete(ctx, guildID, req.Data.(dismodel.ChannelDeleteCall))
	case call.DiscordChannelUpdatePermissions:
		res, err = h.callDiscordChannelUpdatePermissions(ctx, guildID, req.Data.(dismodel.ChannelUpdatePermissionsCall))
	case call.DiscordChannelDeletePermissions:
		res, err = h.callDiscordChannelDeletePermissions(ctx, guildID, req.Data.(dismodel.ChannelDeletePermissionsCall))
	case call.DiscordMessageCreate:
		res, err = h.callDiscordMessageCreate(ctx, guildID, req.Data.(dismodel.MessageCreateCall))
	case call.DiscordMessageUpdate:
		res, err = h.callDiscordMessageUpdate(ctx, guildID, req.Data.(dismodel.MessageUpdateCall))
	case call.DiscordMessageDelete:
		res, err = h.callDiscordMessageDelete(ctx, guildID, req.Data.(dismodel.MessageDeleteCall))
	case call.DiscordMessageGet:
		res, err = h.callDiscordMessageGet(ctx, guildID, req.Data.(dismodel.MessageGetCall))
	case call.DiscordInteractionResponseCreate:
		res, err = h.callDiscordInteractionResponseCreate(ctx, guildID, req.Data.(dismodel.InteractionResponseCreateCall))
	case call.DiscordGuildGet:
		res, err = h.callDiscordGuildGet(ctx, guildID, req.Data.(dismodel.GuildGetCall))
	case call.DiscordRoleList:
		res, err = h.callDiscordRoleList(ctx, guildID, req.Data.(dismodel.RoleListCall))
	default:
		return nil, fmt.Errorf("unknown call type: %s", req.Type)
	}

	err = h.deploymentMetrics.CreateDeploymentMetricEntry(ctx, model.DeploymentMetricEntry{
		DeploymentID:  deploymentID,
		Type:          model.DeploymentMetricEntryTypeCallExecuted,
		CallTotalTime: time.Since(startTime),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Error creating deployment metric entry for action")
	}

	if err != nil {
		return nil, modelError(err)
	}

	return res, nil
}
