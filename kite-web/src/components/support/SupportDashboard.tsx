import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useEffect, useState, type ReactNode } from "react";
import Head from "next/head";
import { useRouter } from "next/router";
import {
  ActivityIcon,
  ClipboardListIcon,
  FileTextIcon,
  LayoutDashboardIcon,
  LoaderCircleIcon,
  ShieldAlertIcon,
  XIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import SupportFlowViewer from "@/components/support/SupportFlowViewer";
import { apiRequest, getApiUrl } from "@/lib/api/client";
import type { FlowContextType } from "@/lib/flow/context";
import type { FlowData } from "@/lib/flow/dataSchema";
import type { App, Command, LogEntry } from "@/lib/types/wire.gen";

type Section = "overview" | "logs" | "commands" | "events";

type EventListener = {
  id: string;
  source: string;
  type: string;
  description: string;
  enabled: boolean;
  app_id: string;
  module_id: string | null;
  creator_user_id: string;
  flow_source: FlowData;
  created_at: string;
  updated_at: string;
  last_run_at: string | null;
};

type SupportLogPage = {
  entries: LogEntry[];
  next_before: number;
};

class SupportApiError extends Error {
  constructor(readonly code: string, message: string) {
    super(message);
  }
}

async function supportRequest<T>(path: string): Promise<T> {
  const response = await apiRequest<T>(path);
  if (!response.success) {
    throw new SupportApiError(response.error.code, response.error.message);
  }
  return response.data;
}

const sections: {
  id: Section;
  label: string;
  icon: typeof LayoutDashboardIcon;
}[] = [
  { id: "overview", label: "Overview", icon: LayoutDashboardIcon },
  { id: "logs", label: "Logs", icon: FileTextIcon },
  { id: "commands", label: "Commands", icon: ClipboardListIcon },
  { id: "events", label: "Event Listeners", icon: ActivityIcon },
];

function dateLabel(value: string | null | undefined) {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.valueOf()) ? value : date.toLocaleString();
}

function errorMessage(error: unknown) {
  return error instanceof Error
    ? error.message
    : "An unexpected error occurred.";
}

function Info({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </dt>
      <dd className="mt-1 break-words text-sm">{children || "—"}</dd>
    </div>
  );
}

function JsonDetails({ value, label }: { value: unknown; label: string }) {
  return (
    <details className="mt-4 rounded-lg border bg-muted/30">
      <summary className="cursor-pointer px-4 py-3 text-sm font-medium">
        {label}
      </summary>
      <pre className="max-h-[32rem] overflow-auto border-t p-4 text-xs leading-relaxed">
        {JSON.stringify(value, null, 2)}
      </pre>
    </details>
  );
}

export default function SupportDashboard({ appId }: { appId: string }) {
  const router = useRouter();
  const [section, setSection] = useState<Section>("overview");
  const [logLevel, setLogLevel] = useState("");
  const [logSearch, setLogSearch] = useState("");
  const access = useQuery({
    queryKey: ["support", "access"],
    queryFn: () =>
      supportRequest<{ authorized: boolean }>("/v1/support/access"),
    retry: false,
  });
  const authorized = access.data?.authorized === true;

  useEffect(() => {
    if (
      access.error instanceof SupportApiError &&
      access.error.code === "unauthorized"
    ) {
      window.location.assign(
        getApiUrl(
          `/v1/auth/login?redirect=${encodeURIComponent(router.asPath)}`
        )
      );
    }
  }, [access.error, router.asPath]);

  const app = useQuery({
    queryKey: ["support", "apps", appId],
    queryFn: () =>
      supportRequest<App>(`/v1/support/apps/${encodeURIComponent(appId)}`),
    enabled: authorized && !!appId,
  });
  const commands = useQuery({
    queryKey: ["support", "apps", appId, "commands"],
    queryFn: () =>
      supportRequest<Command[]>(
        `/v1/support/apps/${encodeURIComponent(appId)}/commands`
      ),
    enabled: authorized && !!appId && section === "commands",
  });
  const listeners = useQuery({
    queryKey: ["support", "apps", appId, "event-listeners"],
    queryFn: () =>
      supportRequest<EventListener[]>(
        `/v1/support/apps/${encodeURIComponent(appId)}/event-listeners`
      ),
    enabled: authorized && !!appId && section === "events",
  });
  const logs = useInfiniteQuery({
    queryKey: ["support", "apps", appId, "logs", logLevel, logSearch],
    queryFn: ({ pageParam }) => {
      const query = new URLSearchParams({ limit: "100" });
      if (pageParam) query.set("before", String(pageParam));
      if (logLevel) query.set("level", logLevel);
      if (logSearch.trim()) query.set("query", logSearch.trim());
      return supportRequest<SupportLogPage>(
        `/v1/support/apps/${encodeURIComponent(appId)}/logs?${query}`
      );
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage) => lastPage.next_before || undefined,
    enabled: authorized && !!appId && section === "logs",
  });

  if (access.isPending) {
    return (
      <main className="grid min-h-screen place-items-center">
        <LoaderCircleIcon className="size-6 animate-spin text-muted-foreground" />
      </main>
    );
  }

  if (!authorized) {
    const unauthenticated =
      access.error instanceof SupportApiError &&
      access.error.code === "unauthorized";
    return (
      <main className="grid min-h-screen place-items-center bg-muted/30 p-6">
        <div className="max-w-md rounded-xl border bg-background p-8 text-center shadow-sm">
          <ShieldAlertIcon className="mx-auto size-9 text-destructive" />
          <h1 className="mt-4 text-xl font-semibold">
            {unauthenticated ? "Redirecting to login" : "Access denied"}
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            {unauthenticated
              ? "Sign in with an authorized support account to continue."
              : access.error
              ? errorMessage(access.error)
              : "You are not authorized to use the support panel."}
          </p>
        </div>
      </main>
    );
  }

  return (
    <>
      <Head>
        <title>Support Staff Panel | Kite</title>
        <meta name="robots" content="noindex, nofollow" />
      </Head>
      <div className="min-h-screen bg-muted/30">
        <header className="flex h-16 items-center justify-between border-b bg-background px-5 md:px-8">
          <div className="flex items-center gap-3">
            <div className="grid size-9 place-items-center rounded-lg bg-primary text-primary-foreground">
              <ShieldAlertIcon className="size-5" />
            </div>
            <div>
              <p className="font-semibold leading-tight">Kite Support</p>
            </div>
          </div>
          <Button variant="outline" onClick={() => router.push("/support")}>
            <XIcon className="mr-2 size-4" />
            Quit app dashboard
          </Button>
        </header>

        <div className="mx-auto flex max-w-[1600px] flex-col md:min-h-[calc(100vh-4rem)] md:flex-row">
          <aside className="border-b bg-background p-4 md:w-64 md:shrink-0 md:border-b-0 md:border-r md:p-5">
            <nav
              aria-label="Support sections"
              className="flex gap-2 overflow-x-auto md:flex-col"
            >
              {sections.map(({ id, label, icon: Icon }) => (
                <button
                  key={id}
                  type="button"
                  onClick={() => setSection(id)}
                  aria-current={section === id ? "page" : undefined}
                  className={`flex shrink-0 items-center gap-3 rounded-lg px-3 py-2.5 text-left text-sm font-medium transition-colors ${
                    section === id
                      ? "bg-primary text-primary-foreground"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  }`}
                >
                  <Icon className="size-4" />
                  {label}
                </button>
              ))}
            </nav>
          </aside>

          <main className="min-w-0 flex-1 p-5 md:p-8">
            <div className="mb-7">
              <h1 className="mt-1 text-2xl font-semibold tracking-tight">
                {sections.find((item) => item.id === section)?.label}
              </h1>
              <p className="mt-1 text-sm text-muted-foreground"></p>
            </div>
            {appId && app.isPending && (
              <div className="flex items-center gap-2 rounded-xl border bg-background p-5 text-sm text-muted-foreground">
                <LoaderCircleIcon className="size-4 animate-spin" />
                Looking up Kite App ID {appId}…
              </div>
            )}
            {appId && app.error && (
              <div
                role="alert"
                className="rounded-xl border border-destructive/40 bg-background p-5 text-sm text-destructive"
              >
                {errorMessage(app.error)}
              </div>
            )}

            {app.data && section === "overview" && (
              <section className="rounded-xl border bg-background p-5 shadow-sm">
                <div className="flex flex-wrap items-start justify-between gap-4 border-b pb-4">
                  <div>
                    <h2 className="text-lg font-semibold">{app.data.name}</h2>
                    <p className="mt-1 break-all font-mono text-xs text-muted-foreground">
                      Kite App ID: {app.data.id}
                    </p>
                  </div>
                </div>
                <dl className="grid gap-x-8 gap-y-5 pt-5 sm:grid-cols-2 lg:grid-cols-3">
                  <Info label="Description">{app.data.description}</Info>
                  <Info label="Discord Bot ID">{app.data.discord_id}</Info>
                  <Info label="Owner User ID">{app.data.owner_user_id}</Info>
                  <Info label="Creator User ID">
                    {app.data.creator_user_id}
                  </Info>
                  <Info label="Created">{dateLabel(app.data.created_at)}</Info>
                  <Info label="Last updated">
                    {dateLabel(app.data.updated_at)}
                  </Info>
                  {app.data.disabled_reason && (
                    <Info label="Disabled reason">
                      {app.data.disabled_reason}
                    </Info>
                  )}
                </dl>
              </section>
            )}

            {app.data && section === "logs" && (
              <section className="space-y-4">
                <div className="flex flex-col gap-3 rounded-xl border bg-background p-4 sm:flex-row">
                  <div className="flex-1">
                    <label
                      htmlFor="log-search"
                      className="mb-2 block text-sm font-medium"
                    >
                      Search log messages
                    </label>
                    <Input
                      id="log-search"
                      value={logSearch}
                      onChange={(event) => setLogSearch(event.target.value)}
                      placeholder="Search message text"
                    />
                  </div>
                  <div className="sm:w-48">
                    <label
                      htmlFor="log-level"
                      className="mb-2 block text-sm font-medium"
                    >
                      Level
                    </label>
                    <select
                      id="log-level"
                      value={logLevel}
                      onChange={(event) => setLogLevel(event.target.value)}
                      className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
                    >
                      <option value="">All levels</option>
                      {["debug", "info", "warn", "error"].map((level) => (
                        <option key={level} value={level}>
                          {level[0].toUpperCase() + level.slice(1)}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
                {logs.isPending ? (
                  <LoadingMessage />
                ) : logs.error ? (
                  <ErrorMessage message={errorMessage(logs.error)} />
                ) : (
                  <>
                    <div className="space-y-2">
                      {logs.data.pages
                        .flatMap((page) => page.entries)
                        .map((entry) => (
                          <article
                            key={entry.id}
                            className="rounded-lg border bg-background p-4"
                          >
                            <div className="flex flex-wrap items-center justify-between gap-2">
                              <span className="rounded-full bg-muted px-2.5 py-1 text-xs font-medium uppercase">
                                {entry.level}
                              </span>
                              <time className="text-xs text-muted-foreground">
                                {dateLabel(entry.created_at)}
                              </time>
                            </div>
                            <p className="mt-3 whitespace-pre-wrap break-words text-sm">
                              {entry.message}
                            </p>
                            {(entry.command_id || entry.event_listener_id) && (
                              <p className="mt-3 break-all font-mono text-xs text-muted-foreground">
                                {entry.command_id
                                  ? `Command ID: ${entry.command_id}`
                                  : `Event Listener ID: ${entry.event_listener_id}`}
                              </p>
                            )}
                          </article>
                        ))}
                      {!logs.data.pages.some((page) => page.entries.length) && (
                        <p className="rounded-lg border bg-background p-6 text-center text-sm text-muted-foreground">
                          No matching log entries.
                        </p>
                      )}
                    </div>
                    {logs.hasNextPage && (
                      <Button
                        variant="outline"
                        onClick={() => logs.fetchNextPage()}
                        disabled={logs.isFetchingNextPage}
                      >
                        {logs.isFetchingNextPage && (
                          <LoaderCircleIcon className="mr-2 size-4 animate-spin" />
                        )}
                        Load older logs
                      </Button>
                    )}
                  </>
                )}
              </section>
            )}

            {app.data && section === "commands" && (
              <section className="space-y-3">
                {commands.isPending ? (
                  <LoadingMessage />
                ) : commands.error ? (
                  <ErrorMessage message={errorMessage(commands.error)} />
                ) : commands.data?.length ? (
                  commands.data.map((command) => (
                    <article
                      key={command.id}
                      className="rounded-xl border bg-background p-5 shadow-sm"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <h2 className="font-semibold">{command.name}</h2>
                          <p className="mt-1 text-sm text-muted-foreground">
                            {command.description || "No description"}
                          </p>
                        </div>
                        <span className="rounded-full bg-muted px-3 py-1 text-xs font-medium">
                          {command.enabled ? "Enabled" : "Disabled"}
                        </span>
                      </div>
                      <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2">
                        <Info label="Command ID">{command.id}</Info>
                        <Info label="Created">
                          {dateLabel(command.created_at)}
                        </Info>
                        <Info label="Last updated">
                          {dateLabel(command.updated_at)}
                        </Info>
                      </dl>
                      <SupportFlowViewer
                        flowData={command.flow_source}
                        context="command"
                      />
                    </article>
                  ))
                ) : (
                  <EmptyMessage message="This app has no commands." />
                )}
              </section>
            )}

            {app.data && section === "events" && (
              <section className="space-y-3">
                {listeners.isPending ? (
                  <LoadingMessage />
                ) : listeners.error ? (
                  <ErrorMessage message={errorMessage(listeners.error)} />
                ) : listeners.data?.length ? (
                  listeners.data.map((listener) => (
                    <article
                      key={listener.id}
                      className="rounded-xl border bg-background p-5 shadow-sm"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <h2 className="font-semibold">
                            {listener.description || listener.type}
                          </h2>
                          <p className="mt-1 text-sm text-muted-foreground">
                            {listener.source} · {listener.type}
                          </p>
                        </div>
                        <span className="rounded-full bg-muted px-3 py-1 text-xs font-medium">
                          {listener.enabled ? "Enabled" : "Disabled"}
                        </span>
                      </div>
                      <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2 lg:grid-cols-3">
                        <Info label="Listener ID">{listener.id}</Info>
                        <Info label="Created">
                          {dateLabel(listener.created_at)}
                        </Info>
                        <Info label="Last updated">
                          {dateLabel(listener.updated_at)}
                        </Info>
                      </dl>
                      <SupportFlowViewer
                        flowData={listener.flow_source}
                        context={
                          (listener.source === "schedule"
                            ? "event_schedule"
                            : listener.source === "webhook"
                            ? "event_webhook"
                            : "event_discord") satisfies FlowContextType
                        }
                      />
                    </article>
                  ))
                ) : (
                  <EmptyMessage message="This app has no event listeners." />
                )}
              </section>
            )}
          </main>
        </div>
      </div>
    </>
  );
}

function LoadingMessage() {
  return (
    <div className="flex items-center gap-2 rounded-xl border bg-background p-5 text-sm text-muted-foreground">
      <LoaderCircleIcon className="size-4 animate-spin" />
      Loading application data…
    </div>
  );
}

function ErrorMessage({ message }: { message: string }) {
  return (
    <div
      role="alert"
      className="rounded-xl border border-destructive/40 bg-background p-5 text-sm text-destructive"
    >
      {message}
    </div>
  );
}

function EmptyMessage({ message }: { message: string }) {
  return (
    <div className="rounded-xl border border-dashed bg-background p-10 text-center text-sm text-muted-foreground">
      {message}
    </div>
  );
}
