import { MCPSession } from "@/components/landing/MCPSessionMockup";
import { cn } from "@/lib/utils";
import { CheckIcon, SparklesIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { usePinnedScrollProgress } from "./usePinnedScrollProgress";

const points = [
  "Click Ask AI in the editor and describe what the command should do.",
  "Flow AI adds and connects the blocks. You check the changes and save.",
  "Your bot can use AI too. It answers questions and searches the web without an API key.",
];

const mcpSession: MCPSession = {
  prompt:
    "Add a /ticket command that creates a private support channel and pings staff.",
  calls: [
    {
      tool: "create_command",
      result: "Created /ticket command flow with 4 connected blocks",
    },
    {
      tool: "deploy_command",
      result: "Registered /ticket in Discord and verified permissions",
    },
  ],
  answer:
    "Added /ticket with a private channel step and staff mention reply. Ready to test in Discord.",
};

function clamp01(n: number) {
  return Math.min(1, Math.max(0, n));
}

export default function HomeAISection() {
  const {
    sectionRef,
    stageRef,
    progress,
    pinState,
    isDesktop,
    reducedMotion,
    scrollToProgress,
  } = usePinnedScrollProgress();
  const [mobileMode, setMobileMode] = useState<"flow-ai" | "mcp">("flow-ai");

  const scrollMode: "flow-ai" | "mcp" = progress < 0.5 ? "flow-ai" : "mcp";
  const activeMode = !isDesktop || reducedMotion ? mobileMode : scrollMode;

  // Stage 1 (0 -> 0.5): initializing prompt is visible immediately at 0 so the
  // card never looks empty before scrolling into the pinned runway.
  const stage1 = clamp01(progress / 0.5);
  const showReply = !isDesktop || reducedMotion || stage1 >= 0.28;
  const showActions = !isDesktop || reducedMotion || stage1 >= 0.62;

  // Stage 2 (0.5 -> 1.0): MCP terminal lines reveal progressively with scroll.
  const stage2 = clamp01((progress - 0.5) / 0.5);
  const showCall1 = !isDesktop || reducedMotion || stage2 >= 0.22;
  const showCall2 = !isDesktop || reducedMotion || stage2 >= 0.48;
  const showAnswer = !isDesktop || reducedMotion || stage2 >= 0.74;

  function jumpToMode(target: "flow-ai" | "mcp") {
    setMobileMode(target);
    scrollToProgress(target === "flow-ai" ? 0.35 : 0.9);
  }

  return (
    <section
      id="ai"
      ref={sectionRef}
      className={cn("relative scroll-mt-16", !reducedMotion && "lg:h-[260vh]")}
    >
      <div
        ref={stageRef}
        className={cn(
          "w-full py-16 sm:py-24 lg:py-0 lg:h-[calc(100vh-3.5rem)] lg:flex lg:items-center",
          !reducedMotion &&
            pinState === "before" &&
            "lg:absolute lg:top-0 lg:inset-x-0",
          !reducedMotion &&
            pinState === "pinned" &&
            "lg:fixed lg:top-14 lg:inset-x-0 lg:z-20",
          !reducedMotion &&
            pinState === "after" &&
            "lg:absolute lg:bottom-0 lg:inset-x-0"
        )}
      >
        <div className="container">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-bold">
                Describe it, and AI builds it
              </h2>
              <p className="text-xl text-muted-foreground mt-4">
                Flow AI is the AI assistant in Kite&apos;s editor. Tell it what
                your command should do in your own words.
              </p>
              <ul className="mt-6 space-y-3">
                {points.map((point) => (
                  <li key={point} className="flex gap-3">
                    <CheckIcon className="flex-none w-5 h-5 mt-0.5 text-primary" />
                    <span>{point}</span>
                  </li>
                ))}
              </ul>
              <p className="mt-6 text-muted-foreground">
                Using Claude, ChatGPT, or Claude Code?{" "}
                <Link
                  href="/discord-mcp-server"
                  className="text-primary hover:underline"
                >
                  Connect them to Kite
                </Link>{" "}
                and build your bot from there.
              </p>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between gap-4">
                <div
                  role="tablist"
                  aria-label="AI workflow previews"
                  className="inline-flex items-center gap-1 rounded-lg border bg-muted/40 p-1"
                >
                  <button
                    type="button"
                    role="tab"
                    aria-selected={activeMode === "flow-ai"}
                    onClick={() => jumpToMode("flow-ai")}
                    className={cn(
                      "rounded-md px-3 py-1 text-xs font-medium transition-colors cursor-pointer",
                      activeMode === "flow-ai"
                        ? "bg-background text-foreground shadow-sm border border-border"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    Flow AI
                  </button>
                  <button
                    type="button"
                    role="tab"
                    aria-selected={activeMode === "mcp"}
                    onClick={() => jumpToMode("mcp")}
                    className={cn(
                      "rounded-md px-3 py-1 text-xs font-medium transition-colors cursor-pointer",
                      activeMode === "mcp"
                        ? "bg-background text-foreground shadow-sm border border-border"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    Claude &amp; ChatGPT
                  </button>
                </div>

                {!reducedMotion && (
                  <div className="hidden lg:flex items-center gap-2.5">
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {activeMode === "flow-ai" ? "1 / 2" : "2 / 2"}
                    </span>
                    <div className="w-24 h-1 rounded-full bg-border overflow-hidden">
                      <div
                        className="h-full w-full rounded-full bg-primary origin-left transition-transform duration-150 ease-out"
                        style={{ transform: `scaleX(${progress})` }}
                      />
                    </div>
                  </div>
                )}
              </div>

              <div className="grid grid-cols-1 items-start">
                {/* Stage 1: Flow AI */}
                <div
                  className={cn(
                    "col-start-1 row-start-1 rounded-xl border bg-card shadow-sm p-5 space-y-4 transition-opacity duration-300 ease-out",
                    activeMode === "flow-ai"
                      ? "opacity-100 z-10 pointer-events-auto"
                      : "opacity-0 z-0 pointer-events-none"
                  )}
                >
                  <div className="flex items-center gap-2 font-semibold">
                    <SparklesIcon className="w-5 h-5 text-primary" />
                    Flow AI
                  </div>

                  <div className="flex justify-end">
                    <div className="rounded-lg bg-primary text-primary-foreground px-4 py-2 max-w-[85%]">
                      Make a /purge command that deletes the last messages and
                      says how many were deleted.
                    </div>
                  </div>

                  <div
                    className={cn(
                      "rounded-lg bg-muted px-4 py-2 max-w-[85%] transition-all duration-300",
                      showReply
                        ? "opacity-100 translate-y-0"
                        : "opacity-0 translate-y-2"
                    )}
                  >
                    I added a number argument for how many messages to delete, a
                    Bulk delete messages block and a reply with the count. Check
                    the highlighted blocks and save when you&apos;re happy.
                  </div>

                  <div
                    className={cn(
                      "flex gap-2 transition-all duration-300",
                      showActions
                        ? "opacity-100 translate-y-0"
                        : "opacity-0 translate-y-2"
                    )}
                  >
                    <div className="rounded-md bg-primary text-primary-foreground text-sm px-3 py-1.5">
                      Save changes
                    </div>
                    <div className="rounded-md border text-sm px-3 py-1.5">
                      Undo
                    </div>
                  </div>
                </div>

                {/* Stage 2: Scroll-progressed MCP terminal */}
                <div
                  className={cn(
                    "col-start-1 row-start-1 overflow-hidden rounded-xl border border-white/10 bg-[#0f0f10] font-mono text-[13px] leading-relaxed text-zinc-200 shadow-2xl transition-opacity duration-300 ease-out",
                    activeMode === "mcp"
                      ? "opacity-100 z-10 pointer-events-auto"
                      : "opacity-0 z-0 pointer-events-none"
                  )}
                >
                  <div className="flex items-center gap-1.5 border-b border-white/10 px-4 py-3">
                    <div className="size-3 rounded-full bg-[#ff5f57]" />
                    <div className="size-3 rounded-full bg-[#febc2e]" />
                    <div className="size-3 rounded-full bg-[#28c840]" />
                    <div className="flex-1 text-center text-xs text-zinc-500">
                      claude · my-bot
                    </div>
                    <div className="w-12" />
                  </div>
                  <div className="space-y-4 p-5">
                    <div className="flex gap-2 rounded-md bg-white/5 px-3 py-2 text-zinc-100">
                      <span className="text-primary">&gt;</span>
                      <span>{mcpSession.prompt}</span>
                    </div>

                    {mcpSession.calls.map((call, i) => {
                      const visible = i === 0 ? showCall1 : showCall2;
                      return (
                        <div
                          key={call.tool}
                          className={cn(
                            "transition-all duration-300",
                            visible
                              ? "opacity-100 translate-y-0"
                              : "opacity-0 translate-y-2"
                          )}
                        >
                          <div className="flex gap-2">
                            <span className="text-emerald-400">●</span>
                            <span>
                              <span className="text-zinc-500">kite · </span>
                              <span className="text-sky-300">{call.tool}</span>
                            </span>
                          </div>
                          <div className="flex gap-2 pl-1 text-zinc-500">
                            <span>⎿</span>
                            <span>{call.result}</span>
                          </div>
                        </div>
                      );
                    })}

                    <div
                      className={cn(
                        "transition-all duration-300",
                        showAnswer
                          ? "opacity-100 translate-y-0"
                          : "opacity-0 translate-y-2"
                      )}
                    >
                      <div className="flex gap-2 text-zinc-100">
                        <span>●</span>
                        <span>{mcpSession.answer}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
