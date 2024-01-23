import { useMutation } from "react-query";
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
} from "./wire";
import { APIResponse } from "./base";

function handleApiResponse<T extends APIResponse<any>>(
  resp: Promise<T>
): Promise<T> {
  return resp;
}

export function useCompileJsMutation() {
  return useMutation((req: CompileJSRequest) => {
    return fetch(`/api/v1/compile/js`, {
      method: "POST",
      body: JSON.stringify(req),
      headers: {
        "Content-Type": "application/json",
      },
    }).then((res) => handleApiResponse<CompileJSResponse>(res.json()));
  });
}

export function useDeploymentCreateMutation() {
  return useMutation(
    ({ guildId, req }: { guildId: string; req: DeploymentCreateRequest }) => {
      return fetch(`/api/v1/guilds/${guildId}/deployments`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<DeploymentCreateResponse>(res.json()));
    }
  );
}

export function useDeploymentDeleteMutation() {
  return useMutation(
    ({ guildId, deploymentId }: { guildId: string; deploymentId: string }) => {
      return fetch(`/api/v1/guilds/${guildId}/deployments/${deploymentId}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<DeploymentDeleteResponse>(res.json()));
    }
  );
}

export function useWorkspaceCreateMutation() {
  return useMutation(
    ({ guildId, req }: { guildId: string; req: WorkspaceCreateRequest }) => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<WorkspaceCreateResponse>(res.json()));
    }
  );
}

export function useWorkspaceUpdateMutation() {
  return useMutation(
    ({
      guildId,
      workspaceId,
      req,
    }: {
      guildId: string;
      workspaceId: string;
      req: WorkspaceCreateRequest;
    }) => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces/${workspaceId}`, {
        method: "PUT",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<WorkspaceUpdateResponse>(res.json()));
    }
  );
}

export function useWorkspaceDeleteMutation() {
  return useMutation(
    ({ guildId, workspaceId }: { guildId: string; workspaceId: string }) => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces/${workspaceId}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<WorkspaceDeleteResponse>(res.json()));
    }
  );
}
