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

export function useDeploymentCreateMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    (req: DeploymentCreateRequest) => {
      return fetch(`/api/v1/guilds/${guildId}/deployments`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<DeploymentCreateResponse>(res.json()));
    },
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
        }
      },
    }
  );
}

export function useDeploymentDeleteMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({ deploymentId }: { deploymentId: string }) => {
      return fetch(`/api/v1/guilds/${guildId}/deployments/${deploymentId}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<DeploymentDeleteResponse>(res.json()));
    },
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
        }
      },
    }
  );
}

export function useWorkspaceCreateMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    (req: WorkspaceCreateRequest) => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<WorkspaceCreateResponse>(res.json()));
    },
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
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
    }) => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces/${workspaceId}`, {
        method: "PUT",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<WorkspaceUpdateResponse>(res.json()));
    },
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
        }
      },
    }
  );
}

export function useWorkspaceDeleteMutation(guildId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({ workspaceId }: { workspaceId: string }) => {
      return fetch(`/api/v1/guilds/${guildId}/workspaces/${workspaceId}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      }).then((res) => handleApiResponse<WorkspaceDeleteResponse>(res.json()));
    },
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["guilds", guildId, "quickAccess"]);
        }
      },
    }
  );
}
