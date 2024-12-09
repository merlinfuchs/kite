import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "./client";
import {
  AppGetResponse,
  AppListResponse,
  AssetGetResponse,
  CommandGetResponse,
  CommandListResponse,
  EventListenerGetResponse,
  EventListenerListResponse,
  LogEntryListResponse,
  MessageGetResponse,
  MessageInstanceListResponse,
  MessageListResponse,
  StateGuildChannelListResponse,
  StateGuildListResponse,
  StateStatusGetResponse,
  UserGetResponse,
  VariableGetResponse,
  VariableListResponse,
} from "../types/wire.gen";

export function useUserQuery(userId = "@me") {
  return useQuery({
    queryKey: ["users", userId],
    queryFn: () => apiRequest<UserGetResponse>(`/v1/users/${userId}`),
    enabled: !!userId,
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

export function useEventListenersQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "event-listeners"],
    queryFn: () =>
      apiRequest<EventListenerListResponse>(
        `/v1/apps/${appId}/event-listeners`
      ),
    enabled: !!appId,
  });
}

export function useEventListenerQuery(appId: string, eventId: string) {
  return useQuery({
    queryKey: ["apps", appId, "event-listeners", eventId],
    queryFn: () =>
      apiRequest<EventListenerGetResponse>(
        `/v1/apps/${appId}/event-listeners/${eventId}`
      ),
    enabled: !!appId && !!eventId,
  });
}

export function useVariablesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "variables"],
    queryFn: () =>
      apiRequest<VariableListResponse>(`/v1/apps/${appId}/variables`),
    enabled: !!appId,
  });
}

export function useVariableQuery(appId: string, variableId: string) {
  return useQuery({
    queryKey: ["apps", appId, "variables", variableId],
    queryFn: () =>
      apiRequest<VariableGetResponse>(
        `/v1/apps/${appId}/variables/${variableId}`
      ),
    enabled: !!appId && !!variableId,
  });
}

export function useMessagesQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "messages"],
    queryFn: () =>
      apiRequest<MessageListResponse>(`/v1/apps/${appId}/messages`),
    enabled: !!appId,
  });
}

export function useMessageQuery(appId: string, messageId: string) {
  return useQuery({
    queryKey: ["apps", appId, "messages", messageId],
    queryFn: () =>
      apiRequest<MessageGetResponse>(`/v1/apps/${appId}/messages/${messageId}`),
    enabled: !!appId && !!messageId,
  });
}

export function useMessageInstancesQuery(appId: string, messageId: string) {
  return useQuery({
    queryKey: ["apps", appId, "messages", messageId, "instances"],
    queryFn: () =>
      apiRequest<MessageInstanceListResponse>(
        `/v1/apps/${appId}/messages/${messageId}/instances`
      ),
    enabled: !!appId && !!messageId,
  });
}

export function useAssetQuery(appId: string, assetId: string) {
  return useQuery({
    queryKey: ["apps", appId, "assets", assetId],
    queryFn: () =>
      apiRequest<AssetGetResponse>(`/v1/apps/${appId}/assets/${assetId}`),
    enabled: !!appId && !!assetId,
  });
}

export function useAppStateStatusQuery(appId?: string) {
  return useQuery({
    queryKey: ["apps", appId, "state", "status"],
    queryFn: () =>
      apiRequest<StateStatusGetResponse>(`/v1/apps/${appId}/state`),
    enabled: !!appId,
  });
}

export function useAppStateGuildsQuery(appId: string) {
  return useQuery({
    queryKey: ["apps", appId, "state", "guilds"],
    queryFn: () =>
      apiRequest<StateGuildListResponse>(`/v1/apps/${appId}/state/guilds`),
    enabled: !!appId,
  });
}

export function useAppStateGuildChannelsQuery(
  appId: string,
  guildId: string | null
) {
  return useQuery({
    queryKey: ["apps", appId, "state", "guilds", guildId, "channels"],
    queryFn: () =>
      apiRequest<StateGuildChannelListResponse>(
        `/v1/apps/${appId}/state/guilds/${guildId}/channels`
      ),
    enabled: !!appId && !!guildId,
  });
}
