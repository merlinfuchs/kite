import { cn } from "@/lib/utils";
import { ReactNode } from "react";

// A drawing of the Discord Developer Portal that points at what to click, for
// the steps of creating a bot. Not a screenshot, so it doesn't go out of date
// with every small change Discord makes.
export default function DeveloperPortalMock({
  page,
}: {
  page: "applications" | "bot";
}) {
  return (
    <div
      aria-hidden
      className="w-full min-w-0 rounded-lg border overflow-hidden select-none text-[11px] leading-tight"
    >
      <div className="flex items-center gap-1.5 bg-[#1e1f22] px-3 py-2">
        <div className="h-2 w-2 rounded-full bg-white/20" />
        <div className="h-2 w-2 rounded-full bg-white/20" />
        <div className="h-2 w-2 rounded-full bg-white/20" />
        <div className="ml-2 flex-auto rounded bg-white/10 px-2 py-0.5 text-white/50 truncate">
          discord.com/developers/applications
        </div>
      </div>
      <div className="flex h-44 bg-[#313338] text-white/80">
        <div className="w-24 sm:w-28 flex-none bg-[#2b2d31] p-2 space-y-1">
          {page === "applications" ? (
            <>
              <SidebarItem active>Applications</SidebarItem>
              <SidebarItem>Teams</SidebarItem>
              <SidebarItem>Embed Debugger</SidebarItem>
            </>
          ) : (
            <>
              <div className="px-2 pb-1 text-white/50 truncate">My Bot</div>
              <SidebarItem>General Information</SidebarItem>
              <SidebarItem>Installation</SidebarItem>
              <SidebarItem>OAuth2</SidebarItem>
              <SidebarItem active highlight>
                Bot
              </SidebarItem>
              <SidebarItem>Emojis</SidebarItem>
            </>
          )}
        </div>
        <div className="flex-auto p-3 min-w-0">
          {page === "applications" ? (
            <>
              <div className="flex items-center justify-between gap-2">
                <div className="text-sm font-semibold text-white truncate">
                  Applications
                </div>
                <Highlight>
                  <div className="rounded bg-[#5865f2] px-2 py-1 font-medium text-white whitespace-nowrap">
                    New Application
                  </div>
                </Highlight>
              </div>
              <div className="mt-4 grid grid-cols-3 gap-2">
                {[0, 1, 2].map((i) => (
                  <div key={i} className="aspect-square rounded bg-white/5" />
                ))}
              </div>
            </>
          ) : (
            <>
              <div className="text-sm font-semibold text-white">Bot</div>
              <div className="mt-3 text-white/50 uppercase tracking-wide text-[9px] font-semibold">
                Token
              </div>
              <div className="mt-1 truncate rounded bg-[#1e1f22] px-2 py-1.5 font-mono text-white/60">
                MTA5ODc2••••••.••••••.••••••••••••
              </div>
              <div className="mt-2 flex gap-2">
                <Highlight label="1">
                  <div className="rounded bg-[#5865f2] px-2 py-1 font-medium text-white whitespace-nowrap">
                    Reset Token
                  </div>
                </Highlight>
                <Highlight label="2">
                  <div className="rounded bg-white/10 px-2 py-1 font-medium text-white">
                    Copy
                  </div>
                </Highlight>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function SidebarItem({
  children,
  active,
  highlight,
}: {
  children: ReactNode;
  active?: boolean;
  highlight?: boolean;
}) {
  const item = (
    <div
      className={cn(
        "rounded px-2 py-1 truncate",
        active ? "bg-white/10 text-white" : "text-white/50"
      )}
    >
      {children}
    </div>
  );
  return highlight ? <Highlight>{item}</Highlight> : item;
}

function Highlight({
  children,
  label,
}: {
  children: ReactNode;
  label?: string;
}) {
  return (
    <div className="relative rounded ring-2 ring-primary ring-offset-1 ring-offset-[#313338]">
      {children}
      {label && (
        <div className="absolute -top-2.5 -right-2.5 h-4 w-4 rounded-full bg-primary text-primary-foreground text-[9px] font-bold flex items-center justify-center">
          {label}
        </div>
      )}
    </div>
  );
}
