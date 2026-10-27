import { entryColor, actionColor } from "@/lib/blocks/colors";
import { cn } from "@/lib/utils";
import {
  EyeIcon,
  ExternalLinkIcon,
  HashIcon,
  IdCardIcon,
  LucideIcon,
  MessageSquareIcon,
  MousePointerClickIcon,
  PickaxeIcon,
  QrCodeIcon,
  ScrollTextIcon,
  ShieldCheckIcon,
  SquareSlashIcon,
  UserPenIcon,
  UserPlusIcon,
} from "lucide-react";
import { ReactNode, useEffect, useRef, useState } from "react";
import DiscordReplyMockup from "./DiscordReplyMockup";
import { PixelHead, QrCode, WelcomeCard } from "./CookieApiVisuals";
import { useInView } from "./Reveal";

interface DemoBlock {
  title: string;
  icon: LucideIcon;
  entry?: boolean;
  cookie?: boolean;
}

interface Demo {
  label: string;
  icon: LucideIcon;
  channel: string;
  command?: string;
  event?: string;
  blocks: DemoBlock[];
  reply: ReactNode;
}

const respond: DemoBlock = {
  title: "Create response message",
  icon: MessageSquareIcon,
};

const demos: Demo[] = [
  {
    label: "Transcripts",
    icon: ScrollTextIcon,
    channel: "ticket-42",
    command: "close",
    blocks: [
      { title: "/close", icon: SquareSlashIcon, entry: true },
      { title: "Create transcript", icon: ScrollTextIcon, cookie: true },
      respond,
    ],
    reply: (
      <>
        Transcript saved, closing the ticket.
        <div className="mt-2 max-w-sm rounded border-l-4 border-[#f9ad15] bg-[#f2f3f5] dark:bg-[#2b2d31] p-3">
          <div className="font-semibold text-[#006ce7] dark:text-[#00a8fc]">
            Transcript of #ticket-42
          </div>
          <div className="text-sm text-[#5c5e66] dark:text-[#949ba4]">
            128 messages, password protected
          </div>
          <div className="mt-2 space-y-1.5">
            {[70, 45, 85].map((width) => (
              <div key={width} className="flex items-center gap-2">
                <div className="w-4 h-4 rounded-full bg-[#5865f2]/60" />
                <div
                  className="h-2 rounded bg-[#c4c9ce] dark:bg-[#4e5058]"
                  style={{ width: `${width}%` }}
                />
              </div>
            ))}
          </div>
        </div>
      </>
    ),
  },
  {
    label: "Welcome cards",
    icon: IdCardIcon,
    channel: "welcome",
    event: "Sam joined the server",
    blocks: [
      { title: "Member joined", icon: UserPlusIcon, entry: true },
      { title: "Generate card", icon: IdCardIcon, cookie: true },
      { title: "Create channel message", icon: MessageSquareIcon },
    ],
    reply: (
      <>
        Welcome to the server,{" "}
        <span className="rounded bg-[#5865f2]/20 px-0.5 text-[#505cdc] dark:text-[#c9cdfb]">
          @Sam
        </span>
        !
        <WelcomeCard
          name="Sam"
          color="#f97316"
          className="mt-2 max-w-sm text-[11px]"
        />
      </>
    ),
  },
  {
    label: "Captcha",
    icon: ShieldCheckIcon,
    channel: "verify",
    event: "Sam clicked Verify",
    blocks: [
      { title: "Verify button", icon: MousePointerClickIcon, entry: true },
      { title: "Create captcha", icon: ShieldCheckIcon, cookie: true },
      respond,
    ],
    reply: (
      <>
        Solve the captcha to get the Verified role.
        <div className="mt-2 inline-flex items-center gap-1.5 rounded bg-[#4e5058] px-4 py-1.5 text-sm font-medium text-white">
          Open captcha
          <ExternalLinkIcon className="h-3.5 w-3.5" />
        </div>
        <div className="mt-1.5 flex items-center gap-1 text-xs text-[#5c5e66] dark:text-[#949ba4]">
          <EyeIcon className="h-3.5 w-3.5" />
          Only you can see this
        </div>
      </>
    ),
  },
  {
    label: "Minecraft",
    icon: PickaxeIcon,
    channel: "link",
    command: "link",
    blocks: [
      { title: "/link", icon: SquareSlashIcon, entry: true },
      { title: "Verify Minecraft player", icon: PickaxeIcon, cookie: true },
      { title: "Edit member nickname", icon: UserPenIcon },
      respond,
    ],
    reply: (
      <div className="flex items-center gap-3">
        <PixelHead className="h-12 w-12 flex-none" />
        <div>
          Linked <span className="font-semibold">Steve</span> (Java Edition).
          Your nickname is now Steve.
        </div>
      </div>
    ),
  },
  {
    label: "QR codes",
    icon: QrCodeIcon,
    channel: "general",
    command: "qr",
    blocks: [
      { title: "/qr", icon: SquareSlashIcon, entry: true },
      { title: "Generate QR code", icon: QrCodeIcon, cookie: true },
      respond,
    ],
    reply: (
      <>
        Scan to visit our website.
        <QrCode className="mt-2 h-32 w-32" />
      </>
    ),
  },
];

const stepMs = 650;
const replyMs = 4500;

// Plays what happens when a Cookie API flow runs: the blocks light up one
// after another, then the bot replies in Discord. It moves on to the next
// example until a tab is clicked.
export default function CookieApiPlayground() {
  const [ref, inView] = useInView<HTMLDivElement>(0.3);
  const [index, setIndex] = useState(0);
  // 0 shows the trigger, then each block runs, then the reply shows.
  const [step, setStep] = useState(0);
  const [autoplay, setAutoplay] = useState(true);
  const tabsRef = useRef<HTMLDivElement>(null);
  const demo = demos[index];
  const done = step > demo.blocks.length;

  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setStep(Infinity);
      setAutoplay(false);
    }
  }, []);

  useEffect(() => {
    if (!inView || (done && !autoplay)) return;
    const timer = setTimeout(
      () => {
        if (!done) {
          setStep((s) => s + 1);
        } else {
          setIndex((i) => (i + 1) % demos.length);
          setStep(0);
        }
      },
      done ? replyMs : stepMs
    );
    return () => clearTimeout(timer);
  }, [inView, done, autoplay, step]);

  // On phones not every tab fits, so the row scrolls to the one playing.
  useEffect(() => {
    const tab = tabsRef.current?.children[index] as HTMLElement | undefined;
    if (tab) {
      tabsRef.current?.scrollTo({
        left: tab.offsetLeft - 8,
        behavior: "smooth",
      });
    }
  }, [index]);

  function select(i: number) {
    setAutoplay(false);
    setIndex(i);
    setStep(0);
  }

  return (
    <div
      ref={ref}
      className="rounded-2xl border bg-muted/50 p-2 shadow-2xl shadow-primary/10"
    >
      <div ref={tabsRef} className="relative flex gap-1 overflow-x-auto p-1">
        {demos.map((d, i) => (
          <button
            key={d.label}
            onClick={() => select(i)}
            className={cn(
              "relative flex flex-none items-center gap-2 overflow-hidden rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              i === index
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground hover:bg-background/50"
            )}
          >
            <d.icon className="h-4 w-4" />
            {d.label}
            {/* How long until the next example plays. */}
            {i === index && autoplay && done && (
              <span className="absolute bottom-0 left-0 h-0.5 w-full origin-left bg-primary animate-[progress_4.5s_linear]" />
            )}
          </button>
        ))}
      </div>

      <div className="mt-2 grid grid-cols-1 gap-2 md:grid-cols-[minmax(0,5fr)_minmax(0,7fr)]">
        <div className="rounded-xl bg-background p-5">
          <div className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-4">
            Your flow in Kite
          </div>
          {demo.blocks.map((block, i) => (
            <FlowStep
              key={`${index}-${i}`}
              block={block}
              state={
                step > i + 1 ? "done" : step === i + 1 ? "running" : "idle"
              }
              last={i === demo.blocks.length - 1}
            />
          ))}
        </div>

        <div className="flex min-h-[340px] flex-col overflow-hidden rounded-xl bg-white dark:bg-[#313338]">
          <div className="flex items-center gap-1.5 border-b border-black/10 px-4 py-3 text-[15px] font-semibold text-[#060607] dark:text-[#f2f3f5]">
            <HashIcon className="h-5 w-5 text-[#80848e]" />
            {demo.channel}
          </div>
          <div className="flex flex-1 flex-col justify-end">
            <DiscordReplyMockup
              key={index}
              command={demo.command}
              event={demo.event}
              className="border-0 bg-transparent dark:bg-transparent motion-safe:animate-in motion-safe:fade-in motion-safe:slide-in-from-bottom-2"
              content={
                done ? (
                  <div className="motion-safe:animate-in motion-safe:fade-in motion-safe:zoom-in-95">
                    {demo.reply}
                  </div>
                ) : (
                  <span className="flex items-center gap-1 text-[#5c5e66] dark:text-[#949ba4] italic">
                    My Bot is thinking
                    <span className="flex gap-0.5">
                      {[0, 150, 300].map((delay) => (
                        <span
                          key={delay}
                          className="h-1 w-1 rounded-full bg-current motion-safe:animate-bounce"
                          style={{ animationDelay: `${delay}ms` }}
                        />
                      ))}
                    </span>
                  </span>
                )
              }
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function FlowStep({
  block,
  state,
  last,
}: {
  block: DemoBlock;
  state: "idle" | "running" | "done";
  last: boolean;
}) {
  const color = block.entry ? entryColor : actionColor;

  return (
    <div className="flex flex-col items-start">
      <div
        className={cn(
          "relative flex w-full items-center gap-3 rounded-lg border bg-card px-3 py-2.5 transition duration-300",
          state === "idle" && "opacity-40",
          state === "running" && "scale-[1.03] border-primary shadow-lg",
          state === "done" && "border-border"
        )}
      >
        <div
          className="flex h-8 w-8 flex-none items-center justify-center rounded-md text-white"
          style={{ background: color }}
        >
          <block.icon className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1 truncate text-sm font-medium">
          {block.title}
        </div>
        {block.cookie && (
          <span className="flex-none rounded bg-[#f9ad15]/20 px-1.5 py-0.5 text-[11px] font-medium text-[#b45309] dark:text-[#f9ad15]">
            Cookie API
          </span>
        )}
        {state === "running" && (
          <span className="absolute -right-1 -top-1 flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full rounded-full bg-primary opacity-75 motion-safe:animate-ping" />
            <span className="relative inline-flex h-3 w-3 rounded-full bg-primary" />
          </span>
        )}
      </div>
      {!last && (
        <div className="relative ml-[27px] h-5 w-0.5 bg-border">
          <div
            className={cn(
              "absolute inset-0 origin-top bg-primary transition-transform duration-500",
              state === "done" ? "scale-y-100" : "scale-y-0"
            )}
          />
        </div>
      )}
    </div>
  );
}
