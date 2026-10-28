import logo from "@/assets/logo/orange@1024.png";
import { discordEventType, eventSourceName } from "@/lib/flow/eventTypes";
import { Template } from "@/lib/flow/templates";
import { cn } from "@/lib/utils";
import { MailIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

// TemplateDiscordPreview looks like what members see of the template in
// Discord: its commands being typed into the message box, or what its event
// listeners run on. rows is how many commands or listeners it shows.
export default function TemplateDiscordPreview({
  template,
  rows = 2,
}: {
  template: Template;
  rows?: number;
}) {
  if (!template.commands.length) {
    return <EventsPreview template={template} rows={rows} />;
  }
  return <CommandsPreview template={template} rows={rows} />;
}

function CommandsPreview({
  template,
  rows,
}: {
  template: Template;
  rows: number;
}) {
  const names = useMemo(
    () => template.commands.map((c) => c.name),
    [template.commands]
  );
  const typed = useTypingLoop(names);
  // The matches first, then the others, so the menu keeps its size.
  const shown = [
    ...template.commands.filter((c) => c.name.startsWith(typed)),
    ...template.commands.filter((c) => !c.name.startsWith(typed)),
  ].slice(0, rows);

  return (
    <div className="rounded-lg bg-[#2b2d31] text-[#dbdee1] shadow-lg overflow-hidden border border-black/20 text-left">
      <div className="px-3 pt-2 pb-1 text-[11px] font-semibold uppercase tracking-wide text-[#b5bac1]">
        Commands matching /{typed}
      </div>
      <div>
        {shown.map((c, i) => (
          <div
            key={c.key}
            className={cn(
              "flex items-center gap-2.5 px-3 h-11",
              i === 0 && "bg-[#404249]"
            )}
          >
            <img src={logo.src} alt="" className="h-5 w-5 flex-none" />
            <div className="min-w-0">
              <div className="text-sm font-medium text-white truncate leading-tight">
                /{c.name}
              </div>
              <div className="text-xs text-[#b5bac1] truncate">
                {c.description}
              </div>
            </div>
          </div>
        ))}
      </div>
      <div className="m-2 mt-1 rounded-md bg-[#383a40] px-3 py-2 text-sm text-white font-mono flex items-center">
        /{typed}
        <span className="ml-px inline-block h-4 w-px bg-white animate-pulse" />
      </div>
    </div>
  );
}

function EventsPreview({
  template,
  rows,
}: {
  template: Template;
  rows: number;
}) {
  return (
    <div className="rounded-lg bg-[#2b2d31] text-[#dbdee1] shadow-lg border border-black/20 p-3 space-y-2.5 text-left">
      {template.eventListeners.slice(0, rows).map((l) => {
        const event =
          l.source === "discord" ? discordEventType(l.type) : undefined;
        const Icon = event?.icon ?? eventSourceName(l.source).icon;
        return (
          <div key={l.key} className="flex items-center gap-2.5">
            <div className="h-7 w-7 rounded-full bg-[#5865f2]/20 text-[#949cf7] flex items-center justify-center flex-none">
              <Icon className="h-3.5 w-3.5" />
            </div>
            <div className="min-w-0">
              <div className="text-[11px] uppercase tracking-wide text-[#b5bac1]">
                Runs when
              </div>
              <div className="text-sm text-white truncate">
                {event?.description ?? eventSourceName(l.source).title}
              </div>
            </div>
          </div>
        );
      })}
      {template.messages
        .slice(0, Math.max(rows - template.eventListeners.length, 0))
        .map((m) => (
          <div key={m.key} className="flex items-center gap-2.5">
            <img src={logo.src} alt="" className="h-7 w-7 flex-none" />
            <div className="text-sm text-white flex items-center gap-1.5 min-w-0">
              <MailIcon className="h-3.5 w-3.5 text-[#b5bac1] flex-none" />
              <span className="truncate">Posts the {m.name} message</span>
            </div>
          </div>
        ))}
    </div>
  );
}

// useTypingLoop is how much of the words is typed, typing one after the
// other, or the one again. It starts with the first one typed, and stays
// there for people who prefer less motion.
function useTypingLoop(words: string[]) {
  const [typed, setTyped] = useState(words[0] ?? "");

  useEffect(() => {
    if (!words.length) return;
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    let word = 0;
    let chars = words[0]!.length;
    // Ticks to wait before going on.
    let pause = 15;
    const id = setInterval(() => {
      if (pause > 0) {
        pause--;
        return;
      }
      if (chars < words[word]!.length) {
        chars++;
        if (chars === words[word]!.length) pause = 15;
      } else {
        word = (word + 1) % words.length;
        chars = 0;
        pause = 3;
      }
      setTyped(words[word]!.slice(0, chars));
    }, 110);
    return () => clearInterval(id);
  }, [words]);

  return typed;
}
