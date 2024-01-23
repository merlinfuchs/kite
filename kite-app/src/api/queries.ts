import { useQuery } from "react-query";
import {
  DeploymentListResponse,
  GuildGetResponse,
  GuildListResponse,
  WorkspaceGetResponse,
  WorkspaceListResponse,
} from "./wire";

export function useGuildsQuery() {
  return useQuery<GuildListResponse>(["guilds"], () => {
    return fetch(`/api/v1/guilds`).then((res) => res.json());
  });
}

export function useGuildQuery(guildId?: string | null) {
  return useQuery<GuildGetResponse>(["guilds", guildId], () => {
    return fetch(`/api/v1/guilds/${guildId}`).then((res) => res.json());
  });
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
