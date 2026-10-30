import { cn } from "@/lib/utils";
import { ReactNode, useEffect, useState } from "react";

// Simplified Discord UI for the status and profile features, drawn with the
// site's colors like the Ask AI panel on the homepage.

const botName = "Kite Helper";

export const customStatus = "🎫 Open a ticket with /ticket";
export const rotatingStatuses = [
  "🎮 Playing with 1,204 members",
  "👀 Watching #support",
  "🎉 Giveaway ends Friday",
];

function AppTag() {
  return (
    <span className="shrink-0 rounded bg-indigo-500 px-1 text-[10px] font-semibold leading-4 text-white">
      APP
    </span>
  );
}

// Avatar shows the emoji, or Kite's logo for the bot. Its tint sits on the
// card's color, so a banner behind it doesn't show through, and padding in
// className becomes a ring like the one on Discord's profiles.
function Avatar({
  emoji,
  animated,
  online,
  className,
}: {
  emoji?: string;
  animated?: boolean;
  online?: boolean;
  className?: string;
}) {
  return (
    <div className={cn("flex-none rounded-full bg-card", className)}>
      <div
        className={cn(
          "relative flex h-full w-full items-center justify-center rounded-full bg-primary/15",
          animated &&
            "bg-gradient-to-br from-primary/40 via-fuchsia-500/30 to-indigo-500/40 bg-[length:200%_200%] motion-safe:animate-banner-shift"
        )}
      >
        {emoji ? (
          <span aria-hidden>{emoji}</span>
        ) : (
          <img
            src="/logo.png"
            alt=""
            className={cn(
              "h-3/5 w-3/5",
              animated && "motion-safe:animate-float"
            )}
          />
        )}
        {online && (
          <span className="absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-[3px] border-card bg-green-500" />
        )}
      </div>
    </div>
  );
}

// useRotating returns the item that's shown now, moving to the next one every
// few seconds like a rotating status does every minute.
function useRotating<T>(items: T[]) {
  const [index, setIndex] = useState(0);
  useEffect(() => {
    if (items.length < 2) return;
    const interval = setInterval(
      () => setIndex((i) => (i + 1) % items.length),
      2500
    );
    return () => clearInterval(interval);
  }, [items.length]);
  return items[index];
}

function Status({ statuses }: { statuses: string[] }) {
  const status = useRotating(statuses);
  return (
    <div
      key={status}
      className="truncate text-xs text-muted-foreground motion-safe:animate-in motion-safe:fade-in motion-safe:slide-in-from-bottom-1 motion-safe:duration-500"
    >
      {status}
    </div>
  );
}

function MemberRow({
  name,
  emoji,
  app,
  statuses,
}: {
  name: string;
  emoji?: string;
  app?: boolean;
  statuses?: string[];
}) {
  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-md px-2 py-1.5",
        app && "bg-muted"
      )}
    >
      <Avatar emoji={emoji} online className="h-8 w-8 text-sm" />
      <div className="min-w-0">
        <div className="flex items-center gap-1.5 text-sm font-medium">
          <span className="truncate">{name}</span>
          {app && <AppTag />}
        </div>
        {statuses && <Status statuses={statuses} />}
      </div>
    </div>
  );
}

// StatusMockup is the member list of a server with the bot's status, which
// rotates when there are several.
export function StatusMockup({ statuses }: { statuses: string[] }) {
  return (
    <div className="space-y-1 rounded-xl border bg-card p-3 shadow-sm">
      <div className="px-2 pb-1 text-xs font-semibold uppercase text-muted-foreground">
        Online — 3
      </div>
      <MemberRow name={botName} app statuses={statuses} />
      <MemberRow name="Alex" emoji="🦊" />
      <MemberRow name="Sam" emoji="🐼" />
    </div>
  );
}

// ProfileMockup is the bot's profile with an animated banner and avatar.
export function ProfileMockup({ statuses }: { statuses?: string[] }) {
  return (
    <div className="overflow-hidden rounded-xl border bg-card shadow-sm">
      <div className="h-20 bg-gradient-to-r from-primary via-fuchsia-500 to-indigo-500 bg-[length:200%_200%] motion-safe:animate-banner-shift" />
      <div className="px-4 pb-4">
        <Avatar animated online className="-mt-9 h-16 w-16 p-[5px]" />
        <div className="mt-2 flex items-center gap-1.5 font-semibold">
          {botName}
          <AppTag />
        </div>
        <div className="text-xs text-muted-foreground">kitehelper</div>
        {statuses && (
          <div className="mt-3 rounded-md bg-muted px-3 py-2">
            <Status statuses={statuses} />
          </div>
        )}
      </div>
    </div>
  );
}

function ChatMessage({
  name,
  emoji,
  children,
}: {
  name: string;
  emoji?: string;
  children: ReactNode;
}) {
  return (
    <div className="flex gap-3">
      <Avatar emoji={emoji} className="h-9 w-9 text-base" />
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-x-1.5">
          <span className="font-medium">{name}</span>
          <AppTag />
          <span className="text-xs text-muted-foreground">Today at 20:02</span>
        </div>
        <div className="text-sm">{children}</div>
      </div>
    </div>
  );
}

const serverProfiles = [
  {
    server: "🎮 Gaming Hub",
    nick: "Quest Master",
    emoji: "🐉",
    message: "Your daily quest is ready!",
  },
  {
    server: "📚 Study Room",
    nick: "Study Buddy",
    emoji: "🦉",
    message: "Focus timer started, 25 minutes.",
  },
];

// ServerProfileMockup is the same bot with another nickname and avatar in two
// servers.
export function ServerProfileMockup() {
  return (
    <div className="space-y-3">
      {serverProfiles.map((p) => (
        <div key={p.server} className="rounded-xl border bg-card p-3 shadow-sm">
          <div className="mb-2 text-xs font-semibold text-muted-foreground">
            {p.server}
          </div>
          <ChatMessage name={p.nick} emoji={p.emoji}>
            {p.message}
          </ChatMessage>
        </div>
      ))}
    </div>
  );
}

// PersonaMockup is a message of the bot next to one it sent with the name and
// avatar of a user.
export function PersonaMockup() {
  return (
    <div className="space-y-4 rounded-xl border bg-card p-4 shadow-sm">
      <ChatMessage name={botName}>Welcome to the server, Alex!</ChatMessage>
      <div>
        <ChatMessage name="Alex" emoji="🦊">
          Movie night starts at 8, bring snacks! 🍿
        </ChatMessage>
        <div className="mt-1 pl-12 text-xs text-muted-foreground">
          Sent by {botName} as Alex
        </div>
      </div>
    </div>
  );
}
