import { useQuery } from "react-query";
import {
  DeploymentGetResponse,
  DeploymentListResponse,
  DeploymentLogEntryListResponse,
  DeploymentLogSummaryGetResponse,
  DeploymentMetricCallsListResponse,
  DeploymentMetricEventsListResponse,
  GuildGetResponse,
  GuildListResponse,
  KVStorageNamespaceKeyListResponse,
  KVStorageNamespaceListResponse,
  QuickAccessItemListResponse,
  UserGetResponse,
  WorkspaceGetResponse,
  WorkspaceListResponse,
} from "../types/wire";
import { apiRequest } from "./client";

export function useUserQuery() {
  return useQuery(["users", "@me"], () =>
    apiRequest<UserGetResponse>(`/v1/users/@me`)
  );
}

export function useGuildsQuery() {
  return useQuery(["guilds"], () =>
    apiRequest<GuildListResponse>("/v1/guilds")
  );
}

export function useGuildQuery(guildId?: string | null) {
  return useQuery(
    ["guilds", guildId],
    () => apiRequest<GuildGetResponse>(`/v1/guilds/${guildId}`),
    { enabled: !!guildId }
  );
}

export function useWorkspacesQuery(guildId?: string | null) {
  return useQuery(
    ["guilds", guildId, "workspaces"],
    () => apiRequest<WorkspaceListResponse>(`/v1/guilds/${guildId}/workspaces`),
    {
      enabled: !!guildId,
    }
  );
}

export function useWorkspaceQuery(
  guildId?: string | null,
  workspaceId?: string | null
) {
  return useQuery(
    ["guilds", guildId, "workspaces", workspaceId],
    () =>
      apiRequest<WorkspaceGetResponse>(
        `/v1/guilds/${guildId}/workspaces/${workspaceId}`
      ),
    {
      enabled: !!guildId && !!workspaceId,
    }
  );
}

export function useDeploymentsQuery(guildId?: string | null) {
  return useQuery(
    ["guilds", guildId, "deployments"],
    () =>
      apiRequest<DeploymentListResponse>(`/v1/guilds/${guildId}/deployments`),
    {
      enabled: !!guildId,
    }
  );
}

export function useDeploymentQuery(
  guildId?: string | null,
  deploymentId?: string | null
) {
  return useQuery(
    ["guilds", guildId, "deployments", deploymentId],
    () =>
      apiRequest<DeploymentGetResponse>(
        `/v1/guilds/${guildId}/deployments/${deploymentId}`
      ),
    {
      enabled: !!guildId,
    }
  );
}

export function useDeploymentLogEntriesQuery(
  guildId?: string | null,
  deploymentId?: string | null
) {
  return useQuery(
    ["guilds", guildId, "deployments", deploymentId, "logs"],
    () =>
      apiRequest<DeploymentLogEntryListResponse>(
        `/v1/guilds/${guildId}/deployments/${deploymentId}/logs`
      ),
    {
      enabled: !!guildId && !!deploymentId,
    }
  );
}

export function useDeploymentLogSummaryQuery(
  guildId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery(
    [
      "guilds",
      guildId,
      "deployments",
      deploymentId,
      "logs",
      "summary",
      timeframe,
    ],
    () =>
      apiRequest<DeploymentLogSummaryGetResponse>(
        `/v1/guilds/${guildId}/deployments/${deploymentId}/logs/summary?timeframe=${timeframe}`
      ),
    {
      enabled: !!guildId && !!deploymentId,
    }
  );
}

export function useDeploymentsEventMetricsQuery(
  guildId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery(
    [
      "guilds",
      guildId,
      "deployments",
      deploymentId,
      "metrics",
      "events",
      timeframe,
    ],
    () =>
      apiRequest<DeploymentMetricEventsListResponse>(
        `/v1/guilds/${guildId}/deployments/${
          deploymentId ? deploymentId + "/" : ""
        }metrics/events?timeframe=${timeframe}`
      ),
    {
      enabled: !!guildId,
    }
  );
}

export function useDeploymentsCallMetricsQuery(
  guildId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery(
    [
      "guilds",
      guildId,
      "deployments",
      deploymentId,
      "metrics",
      "calls",
      timeframe,
    ],
    () =>
      apiRequest<DeploymentMetricCallsListResponse>(
        `/v1/guilds/${guildId}/deployments/${
          deploymentId ? deploymentId + "/" : ""
        }metrics/calls?timeframe=${timeframe}`
      ),
    {
      enabled: !!guildId,
    }
  );
}

export function useKVStorageNamespacesQuery(guildId?: string | null) {
  return useQuery(
    ["guilds", guildId, "kv-storage", "namespaces"],
    () =>
      apiRequest<KVStorageNamespaceListResponse>(
        `/v1/guilds/${guildId}/kv-storage/namespaces`
      ),
    {
      enabled: !!guildId,
    }
  );
}

export function useKVStorageKeysQuery(
  guildId?: string | null,
  namespace?: string | null
) {
  return useQuery(
    ["guilds", guildId, "kv-storage", "namespaces", namespace, "keys"],
    () =>
      apiRequest<KVStorageNamespaceKeyListResponse>(
        `/v1/guilds/${guildId}/kv-storage/namespaces/${namespace}/keys`
      ),
    {
      enabled: !!guildId && !!namespace,
    }
  );
}

export function useQuickAccessItemListQuery(guildId?: string | null) {
  return useQuery(
    ["guilds", guildId, "quickAccess"],
    () =>
      apiRequest<QuickAccessItemListResponse>(
        `/v1/guilds/${guildId}/quick-access`
      ),
    {
      enabled: !!guildId,
    }
  );
}
