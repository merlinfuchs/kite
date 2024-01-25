import { useRouter } from "next/router";

export function useRouteParams() {
  const router = useRouter();

  return {
    guildId: router.query.gid as string,
    deploymentId: router.query.did as string,
    workspaceId: router.query.wid as string,
  };
}
