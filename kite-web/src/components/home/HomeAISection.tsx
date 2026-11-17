import MCPSessionMockup, {
  MCPSession,
} from "@/components/landing/MCPSessionMockup";
import { cn } from "@/lib/utils";
import { CheckIcon, SparklesIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";

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
  const [mode, setMode] = useState<"flow-ai" | "mcp">("flow-ai");
  const sectionRef = useRef<HTMLElement>(null);
  const stageRef = useRef<HTMLDivElement>(null);
  const [progress, setProgress] = useState(1);
  const [pinState, setPinState] = useState<"before" | "pinned" | "after">(
    "before"
  );
  const [reducedMotion, setReducedMotion] = useState(false);

  useEffect(() => {
    const section = sectionRef.current;
    const stage = stageRef.current;
    if (!section || !stage) return;

    const prefersReduced = !window.matchMedia(
      "(prefers-reduced-motion: no-preference)"
    ).matches;
    setReducedMotion(prefersReduced);
    if (prefersReduced) {
      setProgress(1);
      setPinState("before");
      return;
    }

    let rafId = 0;
    const measure = () => {
      rafId = 0;
      const rect = section.getBoundingClientRect();

      if (window.innerWidth < 1024) {
        setPinState("before");
        const start = window.innerHeight * 0.9;
        const end = window.innerHeight * 0.35;
        setProgress(clamp01((start - rect.top) / Math.max(1, start - end)));
        return;
      }

      const stageH = stage.offsetHeight;
      const pinTop = 56;
      const travel = Math.max(1, rect.height - stageH);

      if (rect.top > pinTop) {
        setPinState("before");
        setProgress(0);
      } else if (rect.bottom < pinTop + stageH) {
        setPinState("after");
        setProgress(1);
      } else {
        setPinState("pinned");
        setProgress(clamp01((pinTop - rect.top) / travel));
      }
    };

    const schedule = () => {
      if (!rafId) rafId = window.requestAnimationFrame(measure);
    };

    measure();
    window.addEventListener("scroll", schedule, { passive: true });
    window.addEventListener("resize", schedule, { passive: true });
    return () => {
      window.removeEventListener("scroll", schedule);
      window.removeEventListener("resize", schedule);
      if (rafId) window.cancelAnimationFrame(rafId);
    };
  }, []);

  const showPrompt = reducedMotion || progress >= 0.08;
  const showReply = reducedMotion || progress >= 0.38;
  const showActions = reducedMotion || progress >= 0.68;

  return (
    <section
      id="ai"
      ref={sectionRef}
      className={cn("relative scroll-mt-16", !reducedMotion && "lg:h-[165vh]")}
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
              <div
                role="tablist"
                aria-label="AI workflow previews"
                className="inline-flex items-center gap-1 rounded-lg border bg-muted/40 p-1"
              >
                <button
                  type="button"
                  role="tab"
                  aria-selected={mode === "flow-ai"}
                  onClick={() => setMode("flow-ai")}
                  className={cn(
                    "rounded-md px-3 py-1 text-xs font-medium transition-colors cursor-pointer",
                    mode === "flow-ai"
                      ? "bg-background text-foreground shadow-sm border border-border"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  Flow AI
                </button>
                <button
                  type="button"
                  role="tab"
                  aria-selected={mode === "mcp"}
                  onClick={() => setMode("mcp")}
                  className={cn(
                    "rounded-md px-3 py-1 text-xs font-medium transition-colors cursor-pointer",
                    mode === "mcp"
                      ? "bg-background text-foreground shadow-sm border border-border"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  Claude &amp; ChatGPT
                </button>
              </div>

              {mode === "flow-ai" ? (
                /* Simplified version of the Flow AI panel in the editor, scrubbed by scroll progress. */
                <div className="rounded-xl border bg-card shadow-sm p-5 space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 font-semibold">
                      <SparklesIcon className="w-5 h-5 text-primary" />
                      Flow AI
                    </div>
                    {!reducedMotion && (
                      <div className="hidden lg:block w-24 h-1 rounded-full bg-border overflow-hidden">
                        <div
                          className="h-full w-full rounded-full bg-primary origin-left transition-transform duration-150 ease-out"
                          style={{ transform: `scaleX(${progress})` }}
                        />
                      </div>
                    )}
                  </div>

                  <div
                    className={cn(
                      "flex justify-end transition-all duration-300",
                      showPrompt
                        ? "opacity-100 translate-y-0"
                        : "opacity-0 translate-y-2"
                    )}
                  >
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
              ) : (
                <MCPSessionMockup session={mcpSession} />
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
