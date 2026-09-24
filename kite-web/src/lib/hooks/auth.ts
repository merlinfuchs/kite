import { useAuthLogoutMutation } from "@/lib/api/mutations";
import { useRouter } from "next/router";
import { useCallback } from "react";

export function useLogout() {
  const router = useRouter();
  const logoutMutation = useAuthLogoutMutation();

  return useCallback(() => {
    router.push("/");
    setTimeout(() => logoutMutation.mutate(), 500);
  }, [logoutMutation, router]);
}
