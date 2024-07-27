import debounce from "just-debounce-it";
import MessagePreview from "./MessagePreview";
import { useState } from "react";
import { Message } from "../schema/message";
import { useCurrentMessageStore } from "../state/message";
import { useHookedTheme } from "@/lib/hooks/theme";
import { cn } from "@/lib/utils";

export default function MessageEditorPreview({
  className,
  reducePadding,
}: {
  className?: string;
  reducePadding?: boolean;
}) {
  const [msg, setMsg] = useState<Message>();

  const debouncedSetMessage = debounce(setMsg, 250);

  // We debounce the message preview to prevent it from updating too often.
  useCurrentMessageStore((state) => debouncedSetMessage(state));

  const { theme } = useHookedTheme();

  return (
    <div
      className={cn(
        "overflow-x-hidden no-scrollbar scrollbar-none border",
        className
      )}
    >
      {msg && (
        <MessagePreview
          msg={msg}
          lightTheme={theme === "light"}
          reducePadding={reducePadding}
        />
      )}
    </div>
  );
}
