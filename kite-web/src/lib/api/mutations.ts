import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AppCollaboratorCreateRequest,
  AppCollaboratorCreateResponse,
  AppCollaboratorDeleteResponse,
  AppCreateRequest,
  AppCreateResponse,
  AppDeleteResponse,
  AppStatusUpdateRequest,
  AppStatusUpdateResponse,
  AppTokenUpdateRequest,
  AppTokenUpdateResponse,
  AppUpdateRequest,
  AppUpdateResponse,
  AssetCreateResponse,
  AuthLogoutResponse,
  BillingCheckoutRequest,
  BillingCheckoutResponse,
  CommandCreateRequest,
  CommandCreateResponse,
  CommandDeleteResponse,
  CommandsImportRequest,
  CommandsImportResponse,
  CommandUpdateEnabledRequest,
  CommandUpdateEnabledResponse,
  CommandUpdateRequest,
  CommandUpdateResponse,
  EventListenerCreateRequest,
  EventListenerCreateResponse,
  EventListenerDeleteResponse,
  EventListenersImportRequest,
  EventListenersImportResponse,
  EventListenerUpdateEnabledRequest,
  EventListenerUpdateEnabledResponse,
  EventListenerUpdateRequest,
  EventListenerUpdateResponse,
  MessageCreateRequest,
  MessageCreateResponse,
  MessageDeleteResponse,
  MessageInstanceCreateRequest,
  MessageInstanceCreateResponse,
  MessageInstanceDeleteResponse,
  MessageInstanceUpdateRequest,
  MessageInstanceUpdateResponse,
  MessagesImportRequest,
  MessagesImportResponse,
  MessageUpdateRequest,
  MessageUpdateResponse,
  StateGuildLeaveResponse,
  SubscriptionManageResponse,
  VariableCreateRequest,
  VariableCreateResponse,
  VariableDeleteResponse,
  VariablesImportRequest,
  VariablesImportResponse,
  VariableUpdateRequest,
  VariableUpdateResponse,
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
        method: "PUT",
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

export function useAppStatusUpdateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: AppStatusUpdateRequest) =>
      apiRequest<AppStatusUpdateResponse>(`/v1/apps/${appId}/status`, {
        method: "PUT",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId],
      });
    },
  });
}

export function useAppTokenUpdateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: AppTokenUpdateRequest) =>
      apiRequest<AppTokenUpdateResponse>(`/v1/apps/${appId}/token`, {
        method: "PUT",
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
        method: "DELETE",
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

export function useCommandsImportMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: CommandsImportRequest) =>
      apiRequest<CommandsImportResponse>(`/v1/apps/${appId}/commands/import`, {
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

export function useCommandUpdateEnabledMutation(appId: string, cmdId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: CommandUpdateEnabledRequest) =>
      apiRequest<CommandUpdateEnabledResponse>(
        `/v1/apps/${appId}/commands/${cmdId}/enabled`,
        {
          method: "PUT",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
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

export function useEventListenerCreateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: EventListenerCreateRequest) =>
      apiRequest<EventListenerCreateResponse>(
        `/v1/apps/${appId}/event-listeners`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "event-listeners"],
      });
    },
  });
}

export function useEventListenersImportMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: EventListenersImportRequest) =>
      apiRequest<EventListenersImportResponse>(
        `/v1/apps/${appId}/event-listeners/import`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "event-listeners"],
      });
    },
  });
}

export function useEventListenerUpdateMutation(appId: string, eventId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: EventListenerUpdateRequest) =>
      apiRequest<EventListenerUpdateResponse>(
        `/v1/apps/${appId}/event-listeners/${eventId}`,
        {
          method: "PATCH",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "event-listeners"],
      });
    },
  });
}

export function useEventListenerUpdateEnabledMutation(
  appId: string,
  eventId: string
) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: EventListenerUpdateEnabledRequest) =>
      apiRequest<EventListenerUpdateEnabledResponse>(
        `/v1/apps/${appId}/event-listeners/${eventId}/enabled`,
        {
          method: "PUT",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "event-listeners"],
      });
    },
  });
}

export function useEventListenerDeleteMutation(appId: string, eventId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<EventListenerDeleteResponse>(
        `/v1/apps/${appId}/event-listeners/${eventId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "event-listeners"],
      });
    },
  });
}

export function useVariableCreateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: VariableCreateRequest) =>
      apiRequest<VariableCreateResponse>(`/v1/apps/${appId}/variables`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "variables"],
      });
    },
  });
}

export function useVariablesImportMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: VariablesImportRequest) =>
      apiRequest<VariablesImportResponse>(
        `/v1/apps/${appId}/variables/import`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "variables"],
      });
    },
  });
}

export function useVariableUpdateMutation(appId: string, variableId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: VariableUpdateRequest) =>
      apiRequest<VariableUpdateResponse>(
        `/v1/apps/${appId}/variables/${variableId}`,
        {
          method: "PATCH",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "variables"],
      });
    },
  });
}

export function useVariableDeleteMutation(appId: string, variableId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<VariableDeleteResponse>(
        `/v1/apps/${appId}/variables/${variableId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "variables"],
      });
    },
  });
}

export function useMessageCreateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: MessageCreateRequest) =>
      apiRequest<MessageCreateResponse>(`/v1/apps/${appId}/messages`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages"],
      });
    },
  });
}

export function useMessagesImportMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: MessagesImportRequest) =>
      apiRequest<MessagesImportResponse>(`/v1/apps/${appId}/messages/import`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages"],
      });
    },
  });
}

export function useMessageUpdateMutation(appId: string, messageId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: MessageUpdateRequest) =>
      apiRequest<MessageUpdateResponse>(
        `/v1/apps/${appId}/messages/${messageId}`,
        {
          method: "PATCH",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages"],
      });
    },
  });
}

export function useMessageDeleteMutation(appId: string, messageId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<MessageDeleteResponse>(
        `/v1/apps/${appId}/messages/${messageId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages"],
      });
    },
  });
}

export function useMessageInstanceCreateMutation(
  appId: string,
  messageId: string
) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: MessageInstanceCreateRequest) =>
      apiRequest<MessageInstanceCreateResponse>(
        `/v1/apps/${appId}/messages/${messageId}/instances`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages", messageId, "instances"],
      });
    },
  });
}

export function useMessageInstanceUpdateMutation(
  appId: string,
  messageId: string,
  instanceId: number
) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: MessageInstanceUpdateRequest) =>
      apiRequest<MessageInstanceUpdateResponse>(
        `/v1/apps/${appId}/messages/${messageId}/instances/${instanceId}`,
        {
          method: "PUT",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages", messageId, "instances"],
      });
    },
  });
}

export function useMessageInstanceDeleteMutation(
  appId: string,
  messageId: string,
  instanceId: number
) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiRequest<MessageInstanceDeleteResponse>(
        `/v1/apps/${appId}/messages/${messageId}/instances/${instanceId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "messages", messageId, "instances"],
      });
    },
  });
}

export function useAppStateGuildLeaveMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (guildId: string) =>
      apiRequest<StateGuildLeaveResponse>(
        `/v1/apps/${appId}/state/guilds/${guildId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "state", "guilds"],
      });
    },
  });
}

export function useAssetCreateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (file: File) => {
      const body = new FormData();
      body.append("file", file);

      return apiRequest<AssetCreateResponse>(`/v1/apps/${appId}/assets`, {
        method: "POST",
        body,
      });
    },
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "assets"],
      });
    },
  });
}

export function useAppCollaboratorCreateMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: AppCollaboratorCreateRequest) =>
      apiRequest<AppCollaboratorCreateResponse>(
        `/v1/apps/${appId}/collaborators`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(req),
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "collaborators"],
      });
    },
  });
}

export function useAppCollaboratorDeleteMutation(appId: string) {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (collaboratorId: string) =>
      apiRequest<AppCollaboratorDeleteResponse>(
        `/v1/apps/${appId}/collaborators/${collaboratorId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["apps", appId, "collaborators"],
      });
    },
  });
}

export function useCheckoutCreateMutation(appId: string) {
  return useMutation({
    mutationFn: (req: BillingCheckoutRequest) => {
      return apiRequest<BillingCheckoutResponse>(
        `/v1/apps/${appId}/billing/checkout`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
    },
  });
}

export function useAppSubscriptionManageMutation(subscriptionId: string) {
  return useMutation({
    mutationFn: () =>
      apiRequest<SubscriptionManageResponse>(
        `/v1/billing/subscriptions/${subscriptionId}/manage`,
        {
          method: "POST",
        }
      ),
  });
}
