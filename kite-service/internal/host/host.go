package host

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-types/call"
	"github.com/merlinfuchs/kite/kite-types/kvmodel"
	"github.com/merlinfuchs/kite/kite-types/manifest"
)

type HostEnvironmentStores struct {
	deployments       store.DeploymentStore
	deploymentLogs    store.DeploymentLogStore
	deploymentMetrics store.DeploymentMetricStore
	kvStorage         store.KVStorageStore
	discordState      store.DiscordStateStore
	discordClient     *disrest.Client
}

func NewHostEnvironmentStores(
	deployments store.DeploymentStore,
	deploymentLogs store.DeploymentLogStore,
	deploymentMetrics store.DeploymentMetricStore,
	kvStorage store.KVStorageStore,
	discordState store.DiscordStateStore,
	discordClient *disrest.Client,
) HostEnvironmentStores {
	return HostEnvironmentStores{
		deployments:       deployments,
		deploymentLogs:    deploymentLogs,
		deploymentMetrics: deploymentMetrics,
		kvStorage:         kvStorage,
		discordState:      discordState,
		discordClient:     discordClient,
	}
}

type HostEnvironment struct {
	HostEnvironmentStores

	DeploymentID string
	GuildID      string
	Manifest     manifest.Manifest
	Config       map[string]interface{}
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
	case call.DiscordChannelGet:
		res, err = h.callDiscordChannelGet(ctx, req.Data.(distype.ChannelGetRequest))
	case call.DiscordChannelList:
		res, err = h.callDiscordChannelList(ctx, req.Data.(distype.GuildChannelListRequest))
	case call.DiscordMessageCreate:
		res, err = h.callDiscordMessageCreate(ctx, req.Data.(distype.MessageCreateRequest))
	case call.DiscordMessageUpdate:
		res, err = h.callDiscordMessageUpdate(ctx, req.Data.(distype.MessageEditRequest))
	case call.DiscordMessageDelete:
		res, err = h.callDiscordMessageDelete(ctx, req.Data.(distype.MessageDeleteRequest))
	case call.DiscordMessageGet:
		res, err = h.callDiscordMessageGet(ctx, req.Data.(distype.MessageGetRequest))
	case call.DiscordInteractionResponseCreate:
		res, err = h.callDiscordInteractionResponseCreate(ctx, req.Data.(distype.InteractionResponseCreateRequest))
	case call.DiscordGuildGet:
		res, err = h.callDiscordGuildGet(ctx, req.Data.(distype.GuildGetRequest))
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
