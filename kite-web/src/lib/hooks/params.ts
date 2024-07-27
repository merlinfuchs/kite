import { useRouter } from "next/router";

export function useAppId() {
  const router = useRouter();
  return router.query.appId as string;
}

export function useCommandId() {
  const router = useRouter();
  return router.query.cmdId as string;
}
