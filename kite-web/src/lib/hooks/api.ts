import { useEffect } from "react";
import { APIResponse } from "../api/response";
import {
  useCommandsQuery,
  useAppQuery,
  useAppsQuery,
  useUserQuery,
  useCommandQuery,
  useVariablesQuery,
  useVariableQuery,
} from "../api/queries";
import {
  AppGetResponse,
  AppListResponse,
  CommandGetResponse,
  CommandListResponse,
  UserGetResponse,
  VariableGetResponse,
  VariableListResponse,
} from "../types/wire.gen";
import { useRouter } from "next/router";

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
