import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AppCreateRequest,
  AppCreateResponse,
  AppDeleteResponse,
  AppUpdateRequest,
  AppUpdateResponse,
  AuthLogoutResponse,
  CommandCreateRequest,
  CommandCreateResponse,
  CommandDeleteResponse,
  CommandUpdateRequest,
  CommandUpdateResponse,
} from "../types/wire.gen";
import { apiRequest } from "./client";

export function useAuthLogoutMutation() {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<AuthLogoutResponse>(`/v1/auth/logout`, {
        method: "POST",
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: [],
      });
    },
  });
}

export function useAppCreateMutation() {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: AppCreateRequest) =>
      apiRequest<AppCreateResponse>(`/v1/apps`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps"],
      });
    },
  });
}

export function useAppUpdateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: AppUpdateRequest) =>
      apiRequest<AppUpdateResponse>(`/v1/apps/${appId}`, {
        method: "PATCH",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps"],
      });
    },
  });
}

export function useAppDeleteMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<AppDeleteResponse>(`/v1/apps/${appId}`, {
        method: "POST",
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps"],
      });
    },
  });
}

export function useCommandCreateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: CommandCreateRequest) =>
      apiRequest<CommandCreateResponse>(`/v1/apps/${appId}/commands`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "commands"],
      });
    },
  });
}

export function useCommandUpdateMutation(appId: string, cmdId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: CommandUpdateRequest) =>
      apiRequest<CommandUpdateResponse>(`/v1/apps/${appId}/commands/${cmdId}`, {
        method: "PATCH",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "commands"],
      });
    },
  });
}

export function useCommandDeleteMutation(appId: string, cmdId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<CommandDeleteResponse>(`/v1/apps/${appId}/commands/${cmdId}`, {
        method: "DELETE",
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "commands"],
      });
    },
  });
}
