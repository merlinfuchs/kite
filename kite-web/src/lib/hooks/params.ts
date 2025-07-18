import { useRouter } from "next/router";

export function useAppId() {
  const router = useRouter();
  return router.query.appId as string;
}

export function useCommandId() {
  const router = useRouter();
  return router.query.cmdId as string;
}

export function useVariableId() {
  const router = useRouter();
  return router.query.variableId as string;
}

export function useMessageId() {
  const router = useRouter();
  return router.query.messageId as string;
}

export function useEventId() {
  const router = useRouter();
  return router.query.eventId as string;
}

export function usePluginId() {
  const router = useRouter();
  return router.query.pluginId as string;
}
