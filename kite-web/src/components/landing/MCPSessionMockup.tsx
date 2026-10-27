import { cn } from "@/lib/utils";
import { CSSProperties } from "react";

export interface MCPSession {
  prompt: string;
  // The Kite tools the client calls, with what they found or did.
  calls: { tool: string; result: string }[];
  answer: string;
}

// A simplified terminal of an AI client like Claude Code that builds a bot
// with Kite's tools. It's dark in both themes, like most terminals.
export default function MCPSessionMockup({
  session,
  className,
}: {
  session: MCPSession;
  className?: string;
}) {
  // The lines show up one after the other, like the client is working.
  const line = (i: number): { className: string; style: CSSProperties } => ({
    className:
      "motion-safe:animate-in motion-safe:fade-in motion-safe:slide-in-from-bottom-1 motion-safe:fill-mode-both motion-safe:duration-500",
    style: { animationDelay: `${300 + i * 450}ms` },
  });

  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border border-white/10 bg-[#0f0f10] font-mono text-[13px] leading-relaxed text-zinc-200 shadow-2xl",
        className
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
          <span>{session.prompt}</span>
        </div>
        {session.calls.map((call, i) => (
          <div key={call.tool} {...line(i)}>
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
        ))}
        <div {...line(session.calls.length)}>
          <div className="flex gap-2 text-zinc-100">
            <span>●</span>
            <span>{session.answer}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
