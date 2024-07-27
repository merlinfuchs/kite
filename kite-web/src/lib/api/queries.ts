import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "./client";
import {
  AppGetResponse,
  AppListResponse,
  CommandGetResponse,
  CommandListResponse,
  LogEntryListResponse,
  UserGetResponse,
} from "../types/wire.gen";

export function useUserQuery() {
  return useQuery({
    queryKey: ["users", "@me"],
    queryFn: () => apiRequest<UserGetResponse>(`/v1/users/@me`),
  });
}

export function useAppsQuery() {
  return useQuery({
    queryKey: ["apps"],
    queryFn: () => apiRequest<AppListResponse>(`/v1/apps`),
  });
}

export function useAppQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId],
    queryFn: () => apiRequest<AppGetResponse>(`/v1/apps/${appId}`),
    enabled: !!appId,
  });
}

export function useLogEntriesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "logs"],
    queryFn: () => apiRequest<LogEntryListResponse>(`/v1/apps/${appId}/logs`),
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
