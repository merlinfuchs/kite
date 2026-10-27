import { CheckIcon, LockIcon, SparklesIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { useInView } from "./Reveal";

const prompt =
  "Make a /close command that saves a transcript and deletes the channel";

// Two things Kite does for Cookie API users, each with a small demo.
export default function CookieApiPerks() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
      <div className="rounded-2xl border bg-muted/50 p-6 transition-transform duration-300 hover:-translate-y-1">
        <h3 className="text-xl font-semibold">No token pasting</h3>
        <p className="text-muted-foreground mt-2">
          Transcripts need your bot&apos;s token. Kite adds it to the request
          itself, so it&apos;s never in a block for anyone to see.
        </p>
        <pre className="mt-5 overflow-x-auto rounded-lg bg-[#1e1f22] p-4 text-sm leading-relaxed text-[#dbdee1]">
          <code>
            <span className="text-[#f9ad15]">POST</span> /api/transcript
            {"\n"}Authorization:{" "}
            <span className="text-[#949ba4]">sk_••••••••</span>
            {"\n\n"}
            {'{ "transcript_title": "Ticket 42", "bot_token": '}
            <span className="inline-flex items-center gap-1 rounded bg-emerald-500/20 px-1.5 text-emerald-400">
              <LockIcon className="h-3 w-3" />
              added by Kite
            </span>
            {" }"}
          </code>
        </pre>
      </div>

      <AIPromptCard />
    </div>
  );
}

function AIPromptCard() {
  const [ref, inView] = useInView<HTMLDivElement>(0.5);
  const [typed, setTyped] = useState(0);
  const finished = typed >= prompt.length;

  useEffect(() => {
    if (!inView) return;
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setTyped(prompt.length);
      return;
    }
    if (finished) return;
    const timer = setTimeout(() => setTyped((t) => t + 1), 35);
    return () => clearTimeout(timer);
  }, [inView, typed, finished]);

  return (
    <div
      ref={ref}
      className="rounded-2xl border bg-muted/50 p-6 transition-transform duration-300 hover:-translate-y-1"
    >
      <h3 className="text-xl font-semibold">Or let AI build it</h3>
      <p className="text-muted-foreground mt-2">
        Describe the command you want, and Kite&apos;s AI adds the blocks.
      </p>
      <div className="mt-5 rounded-lg border bg-background p-4 text-sm">
        <div className="flex gap-2">
          <SparklesIcon className="mt-0.5 h-4 w-4 flex-none text-primary" />
          <span>
            {prompt.slice(0, typed)}
            {!finished && (
              <span className="ml-px inline-block h-4 w-0.5 translate-y-0.5 bg-foreground motion-safe:animate-pulse" />
            )}
          </span>
        </div>
        <div
          className={
            finished
              ? "mt-3 flex items-center gap-1.5 text-emerald-500 motion-safe:animate-in motion-safe:fade-in motion-safe:slide-in-from-left-2"
              : "invisible mt-3 flex items-center gap-1.5"
          }
        >
          <CheckIcon className="h-4 w-4" />
          Added /close with 4 blocks
        </div>
      </div>
    </div>
  );
}
