import { ArrowRightIcon, BotIcon } from "lucide-react";

// A simplified Discord message from a bot, replying to a slash command or
// reacting to an event.
export default function DiscordReplyMockup({
  command,
  event,
  content,
}: {
  command?: string;
  event?: string;
  content: string;
}) {
  return (
    <div className="rounded-lg border bg-white dark:bg-[#313338] p-4 text-[15px] leading-snug not-prose">
      {command && (
        <div className="relative flex items-center gap-1.5 pl-[52px] mb-1 text-sm">
          <div className="absolute left-[19px] top-[9px] w-[29px] h-[12px] border-l-2 border-t-2 rounded-tl-md border-[#c4c9ce] dark:border-[#4e5058]" />
          <div className="w-4 h-4 rounded-full bg-emerald-500 text-white text-[10px] font-semibold flex items-center justify-center">
            A
          </div>
          <span className="font-medium text-[#060607] dark:text-[#f2f3f5]">
            Alex
          </span>
          <span className="text-[#5c5e66] dark:text-[#949ba4]">used</span>
          <span className="text-[#006ce7] dark:text-[#00a8fc]">/{command}</span>
        </div>
      )}
      {event && (
        <div className="flex items-center gap-4 mb-4 text-sm text-[#5c5e66] dark:text-[#949ba4]">
          <div className="w-10 flex justify-center">
            <ArrowRightIcon className="w-4 h-4 text-[#23a55a]" />
          </div>
          {event}
        </div>
      )}
      <div className="flex gap-4">
        <div className="flex-none w-10 h-10 rounded-full bg-[#5865f2] flex items-center justify-center">
          <BotIcon className="w-6 h-6 text-white" />
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-1.5">
            <span className="font-medium text-[#060607] dark:text-[#f2f3f5]">
              My Bot
            </span>
            <span className="rounded bg-[#5865f2] px-1 text-[10px] font-semibold leading-4 text-white">
              APP
            </span>
            <span className="text-xs text-[#5c5e66] dark:text-[#949ba4]">
              Today at 4:20 PM
            </span>
          </div>
          <div className="mt-0.5 text-[#313338] dark:text-[#dbdee1]">
            {content}
          </div>
        </div>
      </div>
    </div>
  );
}
