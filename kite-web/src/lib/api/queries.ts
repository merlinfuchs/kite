import { useQueries, useQuery } from "@tanstack/react-query";
import { apiRequest } from "./client";
import {
  AppCollaboratorListResponse,
  AppEmojiListResponse,
  AppEntityListResponse,
  AppGetResponse,
  AppListResponse,
  AssetGetResponse,
  BillingPlanListResponse,
  CommandGetResponse,
  CommandListResponse,
  EventListenerGetResponse,
  EventListenerListResponse,
  FeaturesGetResponse,
  FlowAIUsageGetResponse,
  LogEntry,
  LogEntryListResponse,
  LogSummaryGetResponse,
  MessageGetResponse,
  MessageInstanceListResponse,
  MessageListResponse,
  PluginInstanceGetResponse,
  PluginInstanceListResponse,
  PluginListResponse,
  StateGuildChannelListResponse,
  StateGuildListResponse,
  StateStatusGetResponse,
  SubscriptionListResponse,
  UsageByDayListResponse,
  UsageByTypeListResponse,
  UsageCreditsGetResponse,
  UserGetResponse,
  VariableGetResponse,
  VariableListResponse,
} from "../types/wire.gen";

export function useUserQuery(userId = "@me") {
  return useQuery({
    queryKey: ["users", userId],
    queryFn: () => apiRequest<UserGetResponse>(`/v1/users/${userId}`),
    enabled: !!userId,
  });
}

export function useAppsQuery() {
  return useQuery({
    queryKey: ["apps"],
    queryFn: () => apiRequest<AppListResponse>(`/v1/apps`),
  });
}

export function useAppQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId],
    queryFn: () => apiRequest<AppGetResponse>(`/v1/apps/${appId}`),
    enabled: !!appId,
  });
}

export function useAppEmojisQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "emojis"],
    queryFn: () => apiRequest<AppEmojiListResponse>(`/v1/apps/${appId}/emojis`),
    enabled: !!appId,
  });
}

export function useAppEntitiesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "entities"],
    queryFn: () =>
      apiRequest<AppEntityListResponse>(`/v1/apps/${appId}/entities`),
    enabled: !!appId,
  });
}

export function useLogEntriesQuery(
  appId: string,
  args?: {
    limit?: number;
    commandId?: string;
    eventId?: string;
    messageId?: string;
    refetchInterval?: number;
  }
) {
  const query = new URLSearchParams();
  if (args?.commandId) {
    query.set("command_id", args.commandId);
  }
  if (args?.eventId) {
    query.set("event_id", args.eventId);
  }
  if (args?.messageId) {
    query.set("message_id", args.messageId);
  }
  if (args?.limit) {
    query.set("limit", args.limit.toString());
  }

  return useQuery({
    queryKey: ["apps", appId, "logs", args],
    queryFn: () =>
      apiRequest<LogEntry[]>(`/v1/apps/${appId}/logs?${query.toString()}`),
    staleTime: 1000 * 60,
    enabled: !!appId,
    refetchInterval: args?.refetchInterval,
  });
}

export function useLogSummaryQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "logs", "summary"],
    queryFn: () =>
      apiRequest<LogSummaryGetResponse>(`/v1/apps/${appId}/logs/summary`),
    enabled: !!appId,
  });
}

export function useUsageCreditsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "usage", "credits"],
    queryFn: () =>
      apiRequest<UsageCreditsGetResponse>(`/v1/apps/${appId}/usage/credits`),
    enabled: !!appId,
  });
}

export function useUsageCreditsByDayQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "usage", "by-day"],
    queryFn: () =>
      apiRequest<UsageByDayListResponse>(`/v1/apps/${appId}/usage/by-day`),
    enabled: !!appId,
  });
}

export function useUsageCreditsByTypeQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "usage", "by-type"],
    queryFn: () =>
      apiRequest<UsageByTypeListResponse>(`/v1/apps/${appId}/usage/by-type`),
    enabled: !!appId,
  });
}

export function useCommandsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "commands"],
    queryFn: () =>
      apiRequest<CommandListResponse>(`/v1/apps/${appId}/commands`),
    enabled: !!appId,
  });
}

export function useCommandQuery(appId: string, cmdId: string) {
  return useQuery({
    queryKey: ["apps", appId, "commands", cmdId],
    queryFn: () =>
      apiRequest<CommandGetResponse>(`/v1/apps/${appId}/commands/${cmdId}`),
    enabled: !!appId && !!cmdId,
  });
}

export function useEventListenersQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "event-listeners"],
    queryFn: () =>
      apiRequest<EventListenerListResponse>(
        `/v1/apps/${appId}/event-listeners`
      ),
    enabled: !!appId,
  });
}

export function useEventListenerQuery(appId: string, eventId: string) {
  return useQuery({
    queryKey: ["apps", appId, "event-listeners", eventId],
    queryFn: () =>
      apiRequest<EventListenerGetResponse>(
        `/v1/apps/${appId}/event-listeners/${eventId}`
      ),
    enabled: !!appId && !!eventId,
  });
}

export function useVariablesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "variables"],
    queryFn: () =>
      apiRequest<VariableListResponse>(`/v1/apps/${appId}/variables`),
    enabled: !!appId,
  });
}

export function useVariableQuery(appId: string, variableId: string) {
  return useQuery({
    queryKey: ["apps", appId, "variables", variableId],
    queryFn: () =>
      apiRequest<VariableGetResponse>(
        `/v1/apps/${appId}/variables/${variableId}`
      ),
    enabled: !!appId && !!variableId,
  });
}

export function useMessagesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "messages"],
    queryFn: () =>
      apiRequest<MessageListResponse>(`/v1/apps/${appId}/messages`),
    enabled: !!appId,
  });
}

export function useMessageQuery(appId: string, messageId: string) {
  return useQuery({
    queryKey: ["apps", appId, "messages", messageId],
    queryFn: () =>
      apiRequest<MessageGetResponse>(`/v1/apps/${appId}/messages/${messageId}`),
    enabled: !!appId && !!messageId,
  });
}

export function useMessageInstancesQuery(
  appId: string,
  messageId: string,
  sentBy?: "flow",
  enabled = true
) {
  return useQuery({
    queryKey: [
      "apps",
      appId,
      "messages",
      messageId,
      "instances",
      ...(sentBy ? [sentBy] : []),
    ],
    queryFn: () =>
      apiRequest<MessageInstanceListResponse>(
        `/v1/apps/${appId}/messages/${messageId}/instances` +
          (sentBy ? `?sent_by=${sentBy}` : "")
      ),
    enabled: enabled && !!appId && !!messageId,
  });
}

function assetQueryOptions(appId: string, assetId: string) {
  return {
    queryKey: ["apps", appId, "assets", assetId],
    queryFn: () =>
      apiRequest<AssetGetResponse>(`/v1/apps/${appId}/assets/${assetId}`),
    enabled: !!appId && !!assetId,
  };
}

export function useAssetQuery(appId: string, assetId: string) {
  return useQuery(assetQueryOptions(appId, assetId));
}

// Module level so react-query can keep the combined result stable between renders.
function loadedAssets(
  results: { data?: Awaited<ReturnType<typeof apiRequest<AssetGetResponse>>> }[]
) {
  return results.flatMap((q) => (q.data?.success ? [q.data.data] : []));
}

/** The assets that have loaded so far, in the order of `assetIds`. */
export function useAssetQueries(appId: string, assetIds: string[]) {
  return useQueries({
    queries: assetIds.map((assetId) => assetQueryOptions(appId, assetId)),
    combine: loadedAssets,
  });
}

export function useAppStateStatusQuery(appId?: string) {
  return useQuery({
    queryKey: ["apps", appId, "state", "status"],
    queryFn: () =>
      apiRequest<StateStatusGetResponse>(`/v1/apps/${appId}/state`),
    enabled: !!appId,
  });
}

export function useAppStateGuildsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "state", "guilds"],
    queryFn: () =>
      apiRequest<StateGuildListResponse>(`/v1/apps/${appId}/state/guilds`),
    enabled: !!appId,
  });
}

export function useAppStateGuildChannelsQuery(
  appId: string,
  guildId: string | null
) {
  return useQuery({
    queryKey: ["apps", appId, "state", "guilds", guildId, "channels"],
    queryFn: () =>
      apiRequest<StateGuildChannelListResponse>(
        `/v1/apps/${appId}/state/guilds/${guildId}/channels`
      ),
    enabled: !!appId && !!guildId,
  });
}

export function useAppCollaboratorsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "collaborators"],
    queryFn: () =>
      apiRequest<AppCollaboratorListResponse>(
        `/v1/apps/${appId}/collaborators`
      ),
    enabled: !!appId,
  });
}

export function useAppSubscriptionsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "subscriptions"],
    queryFn: () =>
      apiRequest<SubscriptionListResponse>(
        `/v1/apps/${appId}/billing/subscriptions`
      ),
    enabled: !!appId,
  });
}

export function useBillingPlansQuery() {
  return useQuery({
    queryKey: ["billing", "plans"],
    queryFn: () => apiRequest<BillingPlanListResponse>(`/v1/billing/plans`),
  });
}

export const flowAIUsageQueryKey = (appId: string) => [
  "apps",
  appId,
  "flow-ai",
  "usage",
];

export function useFlowAIUsageQuery(appId: string) {
  return useQuery({
    queryKey: flowAIUsageQueryKey(appId),
    queryFn: () =>
      apiRequest<FlowAIUsageGetResponse>(`/v1/apps/${appId}/flow-ai/usage`),
    enabled: !!appId,
  });
}

export function useAppFeaturesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "billing", "features"],
    queryFn: () =>
      apiRequest<FeaturesGetResponse>(`/v1/apps/${appId}/billing/features`),
    enabled: !!appId,
  });
}

export function usePluginsQuery() {
  return useQuery({
    queryKey: ["plugins"],
    queryFn: () => apiRequest<PluginListResponse>(`/v1/plugins`),
  });
}

export function usePluginInstancesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "plugins"],
    queryFn: () =>
      apiRequest<PluginInstanceListResponse>(`/v1/apps/${appId}/plugins`),
    enabled: !!appId,
  });
}

export function usePluginInstanceQuery(appId: string, pluginId: string) {
  return useQuery({
    queryKey: ["apps", appId, "plugins", pluginId],
    queryFn: () =>
      apiRequest<PluginInstanceGetResponse>(
        `/v1/apps/${appId}/plugins/${pluginId}`
      ),
    enabled: !!appId && !!pluginId,
  });
}
