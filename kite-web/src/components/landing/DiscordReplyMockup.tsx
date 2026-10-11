import MessagePreview from "@/components/message/MessagePreview";
import { useHookedTheme } from "@/lib/hooks/theme";
import { cn } from "@/lib/utils";

// Shows what a bot's reply looks like in Discord, using the message preview of the editor.
export default function DiscordReplyMockup({
  trigger,
  content,
}: {
  trigger: string;
  content: string;
}) {
  const { theme } = useHookedTheme();
  const lightTheme = theme === "light";

  return (
    <div
      className={cn(
        "discord-messages rounded-lg border overflow-hidden",
        lightTheme && "discord-light-theme"
      )}
    >
      <div className="text-sm opacity-60 pt-3 pl-[72px]">{trigger}</div>
      <MessagePreview
        msg={{
          content,
          username: "My Bot",
          attachments: [],
          embeds: [],
          components: [],
          tts: false,
        }}
        lightTheme={lightTheme}
      />
    </div>
  );
}
