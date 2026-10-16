import { useQuery } from "@tanstack/react-query";
import Head from "next/head";
import { useRouter } from "next/router";
import { useEffect, useState, type FormEvent } from "react";
import {
  AppWindowIcon,
  LoaderCircleIcon,
  SearchIcon,
  ShieldAlertIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { apiRequest, getApiUrl } from "@/lib/api/client";

type SupportAccess = {
  authorized: boolean;
};

type SupportError = Error & {
  code?: string;
};

async function checkSupportAccess() {
  const response = await apiRequest<SupportAccess>("/v1/support/access");
  if (!response.success) {
    const error = new Error(response.error.message) as SupportError;
    error.code = response.error.code;
    throw error;
  }
  return response.data;
}

export default function SupportPage() {
  const router = useRouter();
  const [appId, setAppId] = useState("");
  const access = useQuery({
    queryKey: ["support", "access"],
    queryFn: checkSupportAccess,
    retry: false,
  });

  useEffect(() => {
    if ((access.error as SupportError | null)?.code === "unauthorized") {
      window.location.assign(getApiUrl("/v1/auth/login?redirect=%2Fsupport"));
    }
  }, [access.error]);

  function openDashboard(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const kiteAppId = appId.trim();
    if (kiteAppId) {
      void router.push(`/support/${encodeURIComponent(kiteAppId)}`);
    }
  }

  if (access.isPending) {
    return (
      <main className="grid min-h-screen place-items-center">
        <LoaderCircleIcon className="size-6 animate-spin text-muted-foreground" />
      </main>
    );
  }

  if (!access.data?.authorized) {
    return (
      <main className="grid min-h-screen place-items-center bg-muted/30 p-6">
        <div className="max-w-md rounded-xl border bg-background p-8 text-center shadow-sm">
          <ShieldAlertIcon className="mx-auto size-9 text-destructive" />
          <h1 className="mt-4 text-xl font-semibold">
            {(access.error as SupportError | null)?.code === "unauthorized"
              ? "Redirecting to login"
              : "Access denied"}
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            {access.error?.message ||
              "You are not authorized to use the support panel."}
          </p>
        </div>
      </main>
    );
  }

  return (
    <>
      <Head>
        <title>Kite Support | Staff Access</title>
        <meta name="robots" content="noindex, nofollow" />
      </Head>
      <main className="grid min-h-screen place-items-center bg-muted/30 p-6">
        <section className="w-full max-w-lg rounded-2xl border bg-background p-7 shadow-sm sm:p-9">
          <div className="mx-auto grid size-12 place-items-center rounded-xl bg-primary text-primary-foreground">
            <ShieldAlertIcon className="size-6" />
          </div>
          <div className="mt-5 text-center">
            <p className="text-sm font-medium text-primary">Kite Support</p>
            <h1 className="mt-1 text-2xl font-semibold tracking-tight">
              Open an app dashboard
            </h1>
          </div>

          <form onSubmit={openDashboard} className="mt-7 space-y-3">
            <label htmlFor="kite-app-id" className="block text-sm font-medium">
              Kite App ID
            </label>
            <Input
              id="kite-app-id"
              value={appId}
              onChange={(event) => setAppId(event.target.value)}
              placeholder="Paste the Kite App ID"
              autoComplete="off"
              required
            />
            <Button type="submit" className="w-full" disabled={!appId.trim()}>
              <SearchIcon className="mr-2 size-4" />
              Open app dashboard
            </Button>
          </form>
          <div className="mt-6 flex items-center justify-center gap-2 text-xs text-muted-foreground">
            <AppWindowIcon className="size-4" />
            Staff access is checked by the server.
          </div>
        </section>
      </main>
    </>
  );
}
