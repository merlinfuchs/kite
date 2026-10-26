import { keepPreviousData, useQueries, useQuery } from "@tanstack/react-query";
import { apiRequest } from "./client";
import {
  AppCollaboratorInviteListResponse,
  AppCollaboratorListResponse,
  AppEmojiListResponse,
  AppEntityListResponse,
  AppGetResponse,
  AppListResponse,
  AppSetupGetResponse,
  AssetGetResponse,
  BillingPlanListResponse,
  CommandGetResponse,
  CommandListResponse,
  OpenBundleReportListResponse,
  PendingBundleReviewGetResponse,
  PendingBundleReviewListResponse,
  EventListenerGetResponse,
  EventListenerListResponse,
  FeaturesGetResponse,
  FlowAIUsageGetResponse,
  LogEntry,
  LogEntryListResponse,
  LogSummaryGetResponse,
  MessageGetResponse,
  BundleGetResponse,
  BundleInstallListResponse,
  BundleListResponse,
  UserBundleListResponse,
  MessageInstanceListResponse,
  MessageListResponse,
  PluginInstanceGetResponse,
  PluginInstanceListResponse,
  PluginListResponse,
  StateGuildChannelListResponse,
  StateGuildRoleListResponse,
  StateEmojiListResponse,
  StateGuildListResponse,
  StateStatusGetResponse,
  StateUserGetResponse,
  SubscriptionListResponse,
  UsageByDayListResponse,
  UsageByTypeListResponse,
  UsageCreditsGetResponse,
  UserGetResponse,
  VariableGetResponse,
  VariableListResponse,
  VariableValueListResponse,
  AppSecretListResponse,
  AppIntegrationListResponse,
  ItemVersionGetResponse,
  ItemVersionListResponse,
  StatsGetResponse,
  APITokenListResponse,
} from "../types/wire.gen";

export function useUserQuery(userId = "@me") {
  return useQuery({
    queryKey: ["users", userId],
    queryFn: () => apiRequest<UserGetResponse>(`/v1/users/${userId}`),
    enabled: !!userId,
  });
}

export function useStatsQuery(enabled = true) {
  return useQuery({
    queryKey: ["stats"],
    queryFn: () => apiRequest<StatsGetResponse>(`/v1/stats`),
    enabled,
    // The API only updates them once an hour.
    staleTime: 60 * 60 * 1000,
  });
}

export function useAppsQuery(enabled = true) {
  return useQuery({
    queryKey: ["apps"],
    queryFn: () => apiRequest<AppListResponse>(`/v1/apps`),
    enabled,
  });
}

export function useAppQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId],
    queryFn: () => apiRequest<AppGetResponse>(`/v1/apps/${appId}`),
    enabled: !!appId,
    // Changed commands are deployed to Discord in the background.
    refetchInterval: (query) => {
      const app = query.state.data?.success ? query.state.data.data : null;
      return app?.commands_deploy_pending && app.enabled ? 3000 : false;
    },
  });
}

export function useAppStateUserQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "state", "user"],
    queryFn: () =>
      apiRequest<StateUserGetResponse>(`/v1/apps/${appId}/state/user`),
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
    enabled?: boolean;
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
    enabled: !!appId && (args?.enabled ?? true),
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

// The marketplace, with only what the cards show. The official bundles only
// change when Kite is deployed, the listed ones of users when a review
// approves them.
export function useBundlesQuery() {
  return useQuery({
    queryKey: ["bundles"],
    queryFn: () => apiRequest<BundleListResponse>(`/v1/bundles`),
    staleTime: 5 * 60 * 1000,
  });
}

// usePendingBundleReviewsQuery are the bundles waiting for a review, without
// their content, which only admins can see.
export function usePendingBundleReviewsQuery() {
  return useQuery({
    queryKey: ["admin", "bundle-reviews"],
    queryFn: () =>
      apiRequest<PendingBundleReviewListResponse>(`/v1/admin/bundle-reviews`),
  });
}

// usePendingBundleReviewQuery is a review waiting for a decision with the
// content of the version it reviews.
export function usePendingBundleReviewQuery(reviewId: string) {
  return useQuery({
    queryKey: ["admin", "bundle-reviews", reviewId],
    queryFn: () =>
      apiRequest<PendingBundleReviewGetResponse>(
        `/v1/admin/bundle-reviews/${reviewId}`
      ),
    enabled: !!reviewId,
  });
}

// useOpenBundleReportsQuery are the reports of bundles waiting for a decision,
// which only admins can see.
export function useOpenBundleReportsQuery() {
  return useQuery({
    queryKey: ["admin", "bundle-reports"],
    queryFn: () =>
      apiRequest<OpenBundleReportListResponse>(`/v1/admin/bundle-reports`),
  });
}

// useUserBundlesQuery are the bundles the user made, from all of their apps.
export function useUserBundlesQuery() {
  return useQuery({
    queryKey: ["users", "@me", "bundles"],
    queryFn: () => apiRequest<UserBundleListResponse>(`/v1/users/@me/bundles`),
  });
}

// useAPITokensQuery are the user's API tokens, without the tokens themselves.
export function useAPITokensQuery(enabled: boolean) {
  return useQuery({
    queryKey: ["users", "@me", "api-tokens"],
    queryFn: () => apiRequest<APITokenListResponse>(`/v1/users/@me/api-tokens`),
    enabled,
  });
}

// The placeholder is shown until the bundle loads, like the one the page was
// rendered with on the server.
export function useBundleQuery(
  bundleId: string,
  placeholder?: BundleGetResponse | null
) {
  // The list only has what the cards show, so the bundle is always loaded.
  return useQuery({
    queryKey: ["bundles", bundleId],
    queryFn: () => apiRequest<BundleGetResponse>(`/v1/bundles/${bundleId}`),
    enabled: !!bundleId,
    // Users' bundles get new versions or are deleted at any time.
    staleTime: 60_000,
    placeholderData: placeholder
      ? { success: true as const, data: placeholder }
      : undefined,
  });
}

export function useBundleInstallsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "bundle-installs"],
    queryFn: () =>
      apiRequest<BundleInstallListResponse>(
        `/v1/apps/${appId}/bundle-installs`
      ),
    enabled: !!appId,
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

// The command, event listener or message whose version history is shown.
export interface ItemVersionTarget {
  appId: string;
  type: "command" | "event_listener" | "message";
  id: string;
}

const itemVersionPathSegments = {
  command: "commands",
  event_listener: "event-listeners",
  message: "messages",
};

function itemVersionsPath(target: ItemVersionTarget) {
  return `/v1/apps/${target.appId}/${itemVersionPathSegments[target.type]}/${
    target.id
  }/versions`;
}

// Under the key of the item, so saving it refreshes the versions too.
export function useItemVersionsQuery(
  target: ItemVersionTarget,
  enabled: boolean
) {
  return useQuery({
    queryKey: [
      "apps",
      target.appId,
      itemVersionPathSegments[target.type],
      target.id,
      "versions",
    ],
    queryFn: () =>
      apiRequest<ItemVersionListResponse>(itemVersionsPath(target)),
    enabled: enabled && !!target.appId,
    // Every save adds a version, so the list is always refetched.
    staleTime: 0,
  });
}

export function getItemVersion(target: ItemVersionTarget, versionId: string) {
  return apiRequest<ItemVersionGetResponse>(
    `${itemVersionsPath(target)}/${versionId}`
  );
}

export function useVariablesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "variables"],
    queryFn: () =>
      apiRequest<VariableListResponse>(`/v1/apps/${appId}/variables`),
    enabled: !!appId,
  });
}

export function useAppIntegrationsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "integrations"],
    queryFn: () =>
      apiRequest<AppIntegrationListResponse>(`/v1/apps/${appId}/integrations`),
    enabled: !!appId,
  });
}

export function useAppSecretsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "secrets"],
    queryFn: () =>
      apiRequest<AppSecretListResponse>(`/v1/apps/${appId}/secrets`),
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

export function useVariableValuesQuery(
  appId: string,
  variableId: string,
  params: { search: string; limit: number; offset: number }
) {
  return useQuery({
    queryKey: ["apps", appId, "variables", variableId, "values", params],
    queryFn: () => {
      const query = new URLSearchParams({
        search: params.search,
        limit: params.limit.toString(),
        offset: params.offset.toString(),
      });
      return apiRequest<VariableValueListResponse>(
        `/v1/apps/${appId}/variables/${variableId}/values?${query}`
      );
    },
    enabled: !!appId && !!variableId,
    // Keeps the current page visible while the next one or a search loads.
    placeholderData: keepPreviousData,
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

// Polls until every step is done, so steps tick as the user does them.
export function useAppSetupQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "setup"],
    queryFn: () => apiRequest<AppSetupGetResponse>(`/v1/apps/${appId}/setup`),
    enabled: !!appId,
    refetchInterval: (query) => {
      const setup = query.state.data?.success ? query.state.data.data : null;
      const done = setup?.invited && setup.has_command && setup.ran_command;
      return done ? false : 5000;
    },
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

export function useAppStateEmojisQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "state", "emojis"],
    queryFn: () =>
      apiRequest<StateEmojiListResponse>(`/v1/apps/${appId}/state/emojis`),
    enabled: !!appId,
  });
}

export function useAppStateGuildRolesQuery(
  appId: string,
  guildId: string | null
) {
  return useQuery({
    queryKey: ["apps", appId, "state", "guilds", guildId, "roles"],
    queryFn: () =>
      apiRequest<StateGuildRoleListResponse>(
        `/v1/apps/${appId}/state/guilds/${guildId}/roles`
      ),
    enabled: !!appId && !!guildId,
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

export function useAppCollaboratorInvitesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "collaborators", "invites"],
    queryFn: () =>
      apiRequest<AppCollaboratorInviteListResponse>(
        `/v1/apps/${appId}/collaborators/invites`
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

export function useBillingPlansQuery(enabled = true) {
  return useQuery({
    queryKey: ["billing", "plans"],
    queryFn: () => apiRequest<BillingPlanListResponse>(`/v1/billing/plans`),
    enabled,
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
