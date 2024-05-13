import { useQuery } from "react-query";
import {
  AppEntitlementResolvedGetResponse,
  AppGetResponse,
  AppListResponse,
  AppUsageSummaryGetResponse,
  DeploymentGetResponse,
  DeploymentListResponse,
  DeploymentLogEntryListResponse,
  DeploymentLogSummaryGetResponse,
  DeploymentMetricCallsListResponse,
  DeploymentMetricEventsListResponse,
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

export function useAppsQuery() {
  return useQuery(["apps"], () => apiRequest<AppListResponse>("/v1/apps"));
}

export function useAppQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId],
    () => apiRequest<AppGetResponse>(`/v1/apps/${appId}`),
    { enabled: !!appId }
  );
}

export function useWorkspacesQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId, "workspaces"],
    () => apiRequest<WorkspaceListResponse>(`/v1/apps/${appId}/workspaces`),
    {
      enabled: !!appId,
    }
  );
}

export function useWorkspaceQuery(
  appId?: string | null,
  workspaceId?: string | null
) {
  return useQuery(
    ["apps", appId, "workspaces", workspaceId],
    () =>
      apiRequest<WorkspaceGetResponse>(
        `/v1/apps/${appId}/workspaces/${workspaceId}`
      ),
    {
      enabled: !!appId && !!workspaceId,
    }
  );
}

export function useDeploymentsQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId, "deployments"],
    () => apiRequest<DeploymentListResponse>(`/v1/apps/${appId}/deployments`),
    {
      enabled: !!appId,
    }
  );
}

export function useDeploymentQuery(
  appId?: string | null,
  deploymentId?: string | null
) {
  return useQuery(
    ["apps", appId, "deployments", deploymentId],
    () =>
      apiRequest<DeploymentGetResponse>(
        `/v1/apps/${appId}/deployments/${deploymentId}`
      ),
    {
      enabled: !!appId,
    }
  );
}

export function useDeploymentLogEntriesQuery(
  appId?: string | null,
  deploymentId?: string | null
) {
  return useQuery(
    ["apps", appId, "deployments", deploymentId, "logs"],
    () =>
      apiRequest<DeploymentLogEntryListResponse>(
        `/v1/apps/${appId}/deployments/${deploymentId}/logs`
      ),
    {
      enabled: !!appId && !!deploymentId,
    }
  );
}

export function useDeploymentLogSummaryQuery(
  appId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery(
    ["apps", appId, "deployments", deploymentId, "logs", "summary", timeframe],
    () =>
      apiRequest<DeploymentLogSummaryGetResponse>(
        `/v1/apps/${appId}/deployments/${deploymentId}/logs/summary?timeframe=${timeframe}`
      ),
    {
      enabled: !!appId && !!deploymentId,
    }
  );
}

export function useDeploymentsEventMetricsQuery(
  appId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery(
    [
      "apps",
      appId,
      "deployments",
      deploymentId,
      "metrics",
      "events",
      timeframe,
    ],
    () =>
      apiRequest<DeploymentMetricEventsListResponse>(
        `/v1/apps/${appId}/deployments/${
          deploymentId ? deploymentId + "/" : ""
        }metrics/events?timeframe=${timeframe}`
      ),
    {
      enabled: !!appId,
    }
  );
}

export function useDeploymentsCallMetricsQuery(
  appId?: string | null,
  deploymentId?: string | null,
  timeframe: "hour" | "day" | "week" | "month" = "day"
) {
  return useQuery(
    ["apps", appId, "deployments", deploymentId, "metrics", "calls", timeframe],
    () =>
      apiRequest<DeploymentMetricCallsListResponse>(
        `/v1/apps/${appId}/deployments/${
          deploymentId ? deploymentId + "/" : ""
        }metrics/calls?timeframe=${timeframe}`
      ),
    {
      enabled: !!appId,
    }
  );
}

export function useKVStorageNamespacesQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId, "kv-storage", "namespaces"],
    () =>
      apiRequest<KVStorageNamespaceListResponse>(
        `/v1/apps/${appId}/kv-storage/namespaces`
      ),
    {
      enabled: !!appId,
    }
  );
}

export function useKVStorageKeysQuery(
  appId?: string | null,
  namespace?: string | null
) {
  return useQuery(
    ["apps", appId, "kv-storage", "namespaces", namespace, "keys"],
    () =>
      apiRequest<KVStorageNamespaceKeyListResponse>(
        `/v1/apps/${appId}/kv-storage/namespaces/${namespace}/keys`
      ),
    {
      enabled: !!appId && !!namespace,
    }
  );
}

export function useQuickAccessItemListQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId, "quickAccess"],
    () =>
      apiRequest<QuickAccessItemListResponse>(`/v1/apps/${appId}/quick-access`),
    {
      enabled: !!appId,
    }
  );
}

export function useAppUsageSummaryQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId, "usage", "summary"],
    () =>
      apiRequest<AppUsageSummaryGetResponse>(`/v1/apps/${appId}/usage/summary`),
    {
      enabled: !!appId,
    }
  );
}

export function useAppEntitlementsResolvedQuery(appId?: string | null) {
  return useQuery(
    ["apps", appId, "entitlements", "resolved"],
    () =>
      apiRequest<AppEntitlementResolvedGetResponse>(
        `/v1/apps/${appId}/entitlements/resolved`
      ),
    {
      enabled: !!appId,
    }
  );
}
