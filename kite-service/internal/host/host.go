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

type HostEnvironmentStores struct {
	bot               *bot.Bot
	deployments       store.DeploymentStore
	deploymentLogs    store.DeploymentLogStore
	deploymentMetrics store.DeploymentMetricStore
	kvStorage         store.KVStorageStore
}

func NewHostEnvironmentStores(bot *bot.Bot, deployments store.DeploymentStore, deploymentLogs store.DeploymentLogStore, deploymentMetrics store.DeploymentMetricStore, kvStorage store.KVStorageStore) HostEnvironmentStores {
	return HostEnvironmentStores{
		bot:               bot,
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
		kvStorage:         kvStorage,
	}
}

type HostEnvironment struct {
	HostEnvironmentStores

	DeploymentID string
	GuildID      string
}

func NewEnv(stores HostEnvironmentStores) HostEnvironment {
	return HostEnvironment{
		HostEnvironmentStores: stores,
	}
}

func (h HostEnvironment) Call(ctx context.Context, req call.Call) (res interface{}, err error) {
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
		res, err = h.callKVKeyGet(ctx, req.Data.(kvmodel.KVKeyGetCall))
	case call.KVKeySet:
		res, err = h.callKVKeySet(ctx, req.Data.(kvmodel.KVKeySetCall))
	case call.KVKeyDelete:
		res, err = h.callKVKeyDelete(ctx, req.Data.(kvmodel.KVKeyDeleteCall))
	case call.KVKeyIncrease:
		res, err = h.callKVKeyIncrease(ctx, req.Data.(kvmodel.KVKeyIncreaseCall))
	case call.DiscordBanList:
		res, err = h.callDiscordBanList(ctx, req.Data.(dismodel.BanListCall))
	case call.DiscordBanGet:
		res, err = h.callDiscordBanGet(ctx, req.Data.(dismodel.BanGetCall))
	case call.DiscordBanCreate:
		res, err = h.callDiscordBanCreate(ctx, req.Data.(dismodel.BanCreateCall))
	case call.DiscordBanRemove:
		res, err = h.callDiscordBanRemove(ctx, req.Data.(dismodel.BanRemoveCall))
	case call.DiscordChannelGet:
		res, err = h.callDiscordChannelGet(ctx, req.Data.(dismodel.ChannelGetCall))
	case call.DiscordChannelList:
		res, err = h.callDiscordChannelList(ctx, req.Data.(dismodel.ChannelListCall))
	case call.DiscordChannelCreate:
		res, err = h.callDiscordChannelCreate(ctx, req.Data.(dismodel.ChannelCreateCall))
	case call.DiscordChannelUpdate:
		res, err = h.callDiscordChannelUpdate(ctx, req.Data.(dismodel.ChannelUpdateCall))
	case call.DiscordChannelUpdatePositions:
		res, err = h.callDiscordChannelUpdatePositions(ctx, req.Data.(dismodel.ChannelUpdatePositionsCall))
	case call.DiscordChannelDelete:
		res, err = h.callDiscordChannelDelete(ctx, req.Data.(dismodel.ChannelDeleteCall))
	case call.DiscordChannelUpdatePermissions:
		res, err = h.callDiscordChannelUpdatePermissions(ctx, req.Data.(dismodel.ChannelUpdatePermissionsCall))
	case call.DiscordChannelDeletePermissions:
		res, err = h.callDiscordChannelDeletePermissions(ctx, req.Data.(dismodel.ChannelDeletePermissionsCall))
	case call.DiscordMessageCreate:
		res, err = h.callDiscordMessageCreate(ctx, req.Data.(dismodel.MessageCreateCall))
	case call.DiscordMessageUpdate:
		res, err = h.callDiscordMessageUpdate(ctx, req.Data.(dismodel.MessageUpdateCall))
	case call.DiscordMessageDelete:
		res, err = h.callDiscordMessageDelete(ctx, req.Data.(dismodel.MessageDeleteCall))
	case call.DiscordMessageGet:
		res, err = h.callDiscordMessageGet(ctx, req.Data.(dismodel.MessageGetCall))
	case call.DiscordInteractionResponseCreate:
		res, err = h.callDiscordInteractionResponseCreate(ctx, req.Data.(dismodel.InteractionResponseCreateCall))
	case call.DiscordGuildGet:
		res, err = h.callDiscordGuildGet(ctx, req.Data.(dismodel.GuildGetCall))
	case call.DiscordRoleList:
		res, err = h.callDiscordRoleList(ctx, req.Data.(dismodel.RoleListCall))
	default:
		return nil, fmt.Errorf("unknown call type: %s", req.Type)
	}

	err = h.deploymentMetrics.CreateDeploymentMetricEntry(ctx, model.DeploymentMetricEntry{
		DeploymentID:  h.DeploymentID,
		Type:          model.DeploymentMetricEntryTypeCallExecuted,
		CallType:      string(req.Type),
		CallSuccess:   err == nil,
		CallTotalTime: time.Since(startTime),
		Timestamp:     time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Error creating deployment metric entry for action")
	}

	if err != nil {
		return nil, modelError(err)
	}

	return res, nil
}
