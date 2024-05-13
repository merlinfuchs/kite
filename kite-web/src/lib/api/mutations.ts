import { useMutation, useQueryClient } from "react-query";
import {
  CompileRequest,
  CompileResponse,
  DeploymentCreateRequest,
  DeploymentCreateResponse,
  DeploymentDeleteResponse,
  WorkspaceCreateRequest,
  WorkspaceCreateResponse,
  WorkspaceDeleteResponse,
  WorkspaceUpdateRequest,
  WorkspaceUpdateResponse,
} from "../types/wire";
import { apiRequest } from "./client";

export function useCompileMutation() {
  return useMutation((req: CompileRequest) =>
    apiRequest<CompileResponse>(`/v1/compile`, {
      method: "POST",
      body: JSON.stringify(req),
      headers: {
        "Content-Type": "application/json",
      },
    })
  );
}

export function useDeploymentCreateMutation(appId: string | null) {
  const client = useQueryClient();

  return useMutation(
    (req: DeploymentCreateRequest) =>
      apiRequest<DeploymentCreateResponse>(`/v1/apps/${appId}/deployments`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["apps", appId, "quickAccess"]);
          client.invalidateQueries(["apps", appId, "deployments"]);
        }
      },
    }
  );
}

export function useDeploymentDeleteMutation(appId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({ deploymentId }: { deploymentId: string }) =>
      apiRequest<DeploymentDeleteResponse>(
        `/v1/apps/${appId}/deployments/${deploymentId}`,
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
          client.invalidateQueries(["apps", appId, "quickAccess"]);
          client.invalidateQueries(["apps", appId, "deployments"]);
        }
      },
    }
  );
}

export function useWorkspaceCreateMutation(appId: string | null) {
  const client = useQueryClient();

  return useMutation(
    (req: WorkspaceCreateRequest) =>
      apiRequest<WorkspaceCreateResponse>(`/v1/apps/${appId}/workspaces`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    {
      onSuccess: (res) => {
        if (res.success) {
          client.invalidateQueries(["apps", appId, "quickAccess"]);
          client.invalidateQueries(["apps", appId, "workspaces"]);
        }
      },
    }
  );
}

export function useWorkspaceUpdateMutation(appId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({
      workspaceId,
      req,
    }: {
      workspaceId: string;
      req: WorkspaceUpdateRequest;
    }) =>
      apiRequest<WorkspaceUpdateResponse>(
        `/v1/apps/${appId}/workspaces/${workspaceId}`,
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
          client.invalidateQueries(["apps", appId, "quickAccess"]);
          client.invalidateQueries(["apps", appId, "workspaces"]);
        }
      },
    }
  );
}

export function useWorkspaceDeleteMutation(appId: string | null) {
  const client = useQueryClient();

  return useMutation(
    ({ workspaceId }: { workspaceId: string }) =>
      apiRequest<WorkspaceDeleteResponse>(
        `/v1/apps/${appId}/workspaces/${workspaceId}`,
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
          client.invalidateQueries(["apps", appId, "quickAccess"]);
          client.invalidateQueries(["apps", appId, "workspaces"]);
        }
      },
    }
  );
}
