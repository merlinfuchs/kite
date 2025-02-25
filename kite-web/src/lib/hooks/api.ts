import { useRouter } from "next/router";
import { useEffect } from "react";
import {
  useAppCollaboratorsQuery,
  useAppEmojisQuery,
  useAppEntitiesQuery,
  useAppQuery,
  useAppsQuery,
  useAppStateGuildChannelsQuery,
  useAppStateGuildsQuery,
  useAppSubscriptionsQuery,
  useBillingPlansQuery,
  useCommandQuery,
  useCommandsQuery,
  useEventListenerQuery,
  useEventListenersQuery,
  useAppFeaturesQuery,
  useLogSummaryQuery,
  useMessageInstancesQuery,
  useMessageQuery,
  useMessagesQuery,
  useUsageCreditsByDayQuery,
  useUsageCreditsByTypeQuery,
  useUsageCreditsQuery,
  useUserQuery,
  useVariableQuery,
  useVariablesQuery,
} from "../api/queries";
import { APIResponse } from "../api/response";
import {
  AppCollaboratorListResponse,
  AppEmojiListResponse,
  AppEntityListResponse,
  AppGetResponse,
  AppListResponse,
  BillingPlanListResponse,
  CommandGetResponse,
  CommandListResponse,
  EventListenerGetResponse,
  EventListenerListResponse,
  FeaturesGetResponse,
  LogSummaryGetResponse,
  MessageGetResponse,
  MessageInstanceListResponse,
  MessageListResponse,
  StateGuildChannelListResponse,
  StateGuildListResponse,
  SubscriptionListResponse,
  UsageByDayListResponse,
  UsageByTypeListResponse,
  UsageCreditsGetResponse,
  UserGetResponse,
  VariableGetResponse,
  VariableListResponse,
} from "../types/wire.gen";

export function useResponseData<T>(
  {
    data,
  }: {
    data?: APIResponse<T>;
  },
  callback?: (res: APIResponse<T>) => void
): T | undefined {
  useEffect(() => {
    if (data !== undefined && callback) {
      callback(data);
    }
  }, [data, callback]);

  return data?.success ? data.data : undefined;
}

export function useUser(
  callback?: (res: APIResponse<UserGetResponse>) => void
) {
  const query = useUserQuery();
  return useResponseData(query, callback);
}

export function useApps(
  callback?: (res: APIResponse<AppListResponse>) => void
) {
  const query = useAppsQuery();
  return useResponseData(query, callback);
}

export function useApp(callback?: (res: APIResponse<AppGetResponse>) => void) {
  const router = useRouter();

  const query = useAppQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useAppEntities(
  callback?: (res: APIResponse<AppEntityListResponse>) => void
) {
  const router = useRouter();

  const query = useAppEntitiesQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useAppEmojis(
  callback?: (res: APIResponse<AppEmojiListResponse>) => void
) {
  const router = useRouter();

  const query = useAppEmojisQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useCommands(
  callback?: (res: APIResponse<CommandListResponse>) => void
) {
  const router = useRouter();

  const query = useCommandsQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useCommand(
  callback?: (res: APIResponse<CommandGetResponse>) => void
) {
  const router = useRouter();

  const query = useCommandQuery(
    router.query.appId as string,
    router.query.cmdId as string
  );
  return useResponseData(query, callback);
}

export function useEventListeners(
  callback?: (res: APIResponse<EventListenerListResponse>) => void
) {
  const router = useRouter();

  const query = useEventListenersQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useEventListener(
  callback?: (res: APIResponse<EventListenerGetResponse>) => void
) {
  const router = useRouter();

  const query = useEventListenerQuery(
    router.query.appId as string,
    router.query.eventId as string
  );
  return useResponseData(query, callback);
}

export function useVariables(
  callback?: (res: APIResponse<VariableListResponse>) => void
) {
  const router = useRouter();

  const query = useVariablesQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useVariable(
  callback?: (res: APIResponse<VariableGetResponse>) => void
) {
  const router = useRouter();

  const query = useVariableQuery(
    router.query.appId as string,
    router.query.variableId as string
  );
  return useResponseData(query, callback);
}

export function useMessages(
  callback?: (res: APIResponse<MessageListResponse>) => void
) {
  const router = useRouter();

  const query = useMessagesQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useMessage(
  callback?: (res: APIResponse<MessageGetResponse>) => void
) {
  const router = useRouter();

  const query = useMessageQuery(
    router.query.appId as string,
    router.query.messageId as string
  );
  return useResponseData(query, callback);
}

export function useMessageInstances(
  callback?: (res: APIResponse<MessageInstanceListResponse>) => void
) {
  const router = useRouter();

  const query = useMessageInstancesQuery(
    router.query.appId as string,
    router.query.messageId as string
  );
  return useResponseData(query, callback);
}

export function useLogSummary(
  callback?: (res: APIResponse<LogSummaryGetResponse>) => void
) {
  const router = useRouter();

  const query = useLogSummaryQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useUsageCredits(
  callback?: (res: APIResponse<UsageCreditsGetResponse>) => void
) {
  const router = useRouter();

  const query = useUsageCreditsQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useUsageCreditsByDay(
  callback?: (res: APIResponse<UsageByDayListResponse>) => void
) {
  const router = useRouter();

  const query = useUsageCreditsByDayQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useUsageCreditsByType(
  callback?: (res: APIResponse<UsageByTypeListResponse>) => void
) {
  const router = useRouter();

  const query = useUsageCreditsByTypeQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useAppStateGuilds(
  callback?: (res: APIResponse<StateGuildListResponse>) => void
) {
  const router = useRouter();

  const query = useAppStateGuildsQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useAppStateGuild(guildId: string | null) {
  const router = useRouter();

  const query = useAppStateGuildsQuery(router.query.appId as string);
  const data = useResponseData(query);

  return data?.find((g) => g!.id === guildId);
}

export function useAppStateGuildChannels(
  guildId: string | null,
  callback?: (res: APIResponse<StateGuildChannelListResponse>) => void
) {
  const router = useRouter();

  const query = useAppStateGuildChannelsQuery(
    router.query.appId as string,
    guildId
  );
  return useResponseData(query, callback);
}

export function useAppStateGuildChannel(
  guildId: string | null,
  channelId: string | null
) {
  const router = useRouter();

  const query = useAppStateGuildChannelsQuery(
    router.query.appId as string,
    guildId
  );
  const data = useResponseData(query);

  return data?.find((c) => c!.id === channelId);
}

export function useAppCollaborators(
  callback?: (res: APIResponse<AppCollaboratorListResponse>) => void
) {
  const router = useRouter();

  const query = useAppCollaboratorsQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useAppSubscriptions(
  callback?: (res: APIResponse<SubscriptionListResponse>) => void
) {
  const router = useRouter();

  const query = useAppSubscriptionsQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useBillingPlans(
  callback?: (res: APIResponse<BillingPlanListResponse>) => void
) {
  const query = useBillingPlansQuery();
  return useResponseData(query, callback);
}

export function useAppFeatures(
  callback?: (res: APIResponse<FeaturesGetResponse>) => void
) {
  const router = useRouter();

  const query = useAppFeaturesQuery(router.query.appId as string);
  return useResponseData(query, callback);
}

export function useAppFeature<T>(
  accessor: (features: FeaturesGetResponse) => T
): T | undefined {
  const features = useAppFeatures();
  return features ? accessor(features) : undefined;
}
