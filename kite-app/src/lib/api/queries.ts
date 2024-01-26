import { useQuery } from "react-query";
import {
  DeploymentListResponse,
  DeploymentLogEntryListResponse,
  DeploymentLogSummaryGetResponse,
  DeploymentMetricEventsListResponse,
  GuildGetResponse,
  GuildListResponse,
  KVStorageNamespaceKeyListResponse,
  KVStorageNamespaceListResponse,
  QuickAccessItemListResponse,
  WorkspaceGetResponse,
  WorkspaceListResponse,
} from "./wire";

export function useGuildsQuery() {
  return useQuery<GuildListResponse>(["guilds"], () => {
    return fetch(`/api/v1/guilds`).then((res) => res.json());
  });
}

export function useGuildQuery(guildId?: string | null) {
  return useQuery<GuildGetResponse>(
    ["guilds", guildId],
    () => {
      return fetch(`/api/v1/guilds/${guildId}`).then((res) => res.json());
    },
    { enabled: !!guildId }
  );
}

export function useWorkspacesQuery(guildId?: string | null) {
  return useQuery<WorkspaceListResponse>(
    ["guilds", guildId, "workspaces"],
    () => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces`).then((res) =>
        res.json()
      );
    },
    {
      enabled: !!guildId,
    }
  );
}

export function useWorkspaceQuery(
  guildId?: string | null,
  workspaceId?: string | null
) {
  return useQuery<WorkspaceGetResponse>(
    ["guilds", guildId, "workspaces", workspaceId],
    () => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces/${workspaceId}`).then(
        (res) => res.json()
      );
    },
    {
      enabled: !!guildId && !!workspaceId,
    }
  );
}

export function useDeploymentsQuery(guildId?: string | null) {
  return useQuery<DeploymentListResponse>(
    ["guilds", guildId, "deployments"],
    () => {
      return fetch(`/api/v1/guilds/${guildId}/deployments`).then((res) =>
        res.json()
      );
    },
    {
      enabled: !!guildId,
    }
  );
}

export function useDeploymentLogEntriesQuery(
  guildId?: string | null,
  deploymentId?: string | null
) {
  return useQuery<DeploymentLogEntryListResponse>(
    ["guilds", guildId, "deployments", deploymentId, "logs"],
    () => {
      return fetch(
        `/api/v1/guilds/${guildId}/deployments/${deploymentId}/logs`
      ).then((res) => res.json());
    },
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
  return useQuery<DeploymentLogSummaryGetResponse>(
    [
      "guilds",
      guildId,
      "deployments",
      deploymentId,
      "logs",
      "summary",
      timeframe,
    ],
    () => {
      return fetch(
        `/api/v1/guilds/${guildId}/deployments/${deploymentId}/logs/summary?timeframe=${timeframe}`
      ).then((res) => res.json());
    },
    {
      enabled: !!guildId && !!deploymentId,
    }
  );
}

export function useDeploymentEventMetricsQuery(
  guildId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery<DeploymentMetricEventsListResponse>(
    [
      "guilds",
      guildId,
      "deployments",
      deploymentId,
      "metrics",
      "events",
      timeframe,
    ],
    () => {
      return fetch(
        `/api/v1/guilds/${guildId}/deployments/${deploymentId}/metrics/events?timeframe=${timeframe}`
      ).then((res) => res.json());
    },
    {
      enabled: !!guildId && !!deploymentId,
    }
  );
}

export function useKVStorageNamespacesQuery(guildId?: string | null) {
  return useQuery<KVStorageNamespaceListResponse>(
    ["guilds", guildId, "kv-storage", "namespaces"],
    () => {
      return fetch(`/api/v1/guilds/${guildId}/kv-storage/namespaces`).then(
        (res) => res.json()
      );
    },
    {
      enabled: !!guildId,
    }
  );
}

export function useKVStorageKeysQuery(
  guildId?: string | null,
  namespace?: string | null
) {
  return useQuery<KVStorageNamespaceKeyListResponse>(
    ["guilds", guildId, "kv-storage", "namespaces", namespace, "keys"],
    () => {
      return fetch(
        `/api/v1/guilds/${guildId}/kv-storage/namespaces/${namespace}/keys`
      ).then((res) => res.json());
    },
    {
      enabled: !!guildId && !!namespace,
    }
  );
}

export function useQuickAccessItemListQuery(guildId?: string | null) {
  return useQuery<QuickAccessItemListResponse>(
    ["guilds", guildId, "quickAccess"],
    () => {
      return fetch(`/api/v1/guilds/${guildId}/quick-access`).then((res) =>
        res.json()
      );
    },
    {
      enabled: !!guildId,
    }
  );
}
