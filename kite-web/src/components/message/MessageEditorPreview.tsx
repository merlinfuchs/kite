import debounce from "just-debounce-it";
import MessagePreview from "./MessagePreview";
import { useState } from "react";
import { Message } from "@/lib/message/schema";
import { useCurrentMessage } from "@/lib/message/state";
import { useHookedTheme } from "@/lib/hooks/theme";
import { cn } from "@/lib/utils";
import { ScrollArea } from "../ui/scroll-area";

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
  useCurrentMessage((state) => debouncedSetMessage(state));

  const { theme } = useHookedTheme();

  return (
    <div className={cn("overflow-x-hidden border h-full", className)}>
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
