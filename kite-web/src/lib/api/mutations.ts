import { useMutation, useQueryClient } from "react-query";
import {
  CompileJSRequest,
  CompileJSResponse,
  DeploymentCreateRequest,
  DeploymentCreateResponse,
  DeploymentDeleteResponse,
  WorkspaceCreateRequest,
  WorkspaceCreateResponse,
  WorkspaceDeleteResponse,
  WorkspaceUpdateResponse,
} from "../types/wire";
import { APIResponse } from "../types/response";
import { apiRequest } from "./client";

export function useCompileJsMutation() {
  return useMutation((req: CompileJSRequest) =>
    apiRequest<CompileJSResponse>(`/v1/compile/js`, {
      method: "POST",
      body: JSON.stringify(req),
      headers: {
        "Content-Type": "application/json",
      },
    })
  );
}

export function useDeploymentCreateMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    (req: DeploymentCreateRequest) =>
      apiRequest<DeploymentCreateResponse>(
        `/v1/guilds/${guildId}/deployments`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
          client.invalidateQueries(["guilds", guildId, "deployments"]);
        }
      },
    }
  );
}

export function useDeploymentDeleteMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({ deploymentId }: { deploymentId: string }) =>
      apiRequest<DeploymentDeleteResponse>(
        `/v1/guilds/${guildId}/deployments/${deploymentId}`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
          client.invalidateQueries(["guilds", guildId, "deployments"]);
        }
      },
    }
  );
}

export function useWorkspaceCreateMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    (req: WorkspaceCreateRequest) =>
      apiRequest<WorkspaceCreateResponse>(`/v1/guilds/${guildId}/workspaces`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
          client.invalidateQueries(["guilds", guildId, "workspaces"]);
        }
      },
    }
  );
}

export function useWorkspaceUpdateMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({
      workspaceId,
      req,
    }: {
      workspaceId: string;
      req: WorkspaceCreateRequest;
    }) =>
      apiRequest<WorkspaceUpdateResponse>(
        `/v1/guilds/${guildId}/workspaces/${workspaceId}`,
        {
          method: "PUT",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
          client.invalidateQueries(["guilds", guildId, "workspaces"]);
        }
      },
    }
  );
}

export function useWorkspaceDeleteMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({ workspaceId }: { workspaceId: string }) =>
      apiRequest<WorkspaceDeleteResponse>(
        `/v1/guilds/${guildId}/workspaces/${workspaceId}`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
          client.invalidateQueries(["guilds", guildId, "workspaces"]);
        }
      },
    }
  );
}
