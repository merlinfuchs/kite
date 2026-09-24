import debounce from "just-debounce-it";
import MessagePreview from "./MessagePreview";
import { useEffect, useState } from "react";
import { Message } from "@/lib/message/schema";
import { getMessage, useDocumentStoreApi } from "@/lib/message/state";
import { useHookedTheme } from "@/lib/hooks/theme";
import { cn } from "@/lib/utils";

export default function MessageEditorPreview({
  className,
  reducePadding,
}: {
  className?: string;
  reducePadding?: boolean;
}) {
  const store = useDocumentStoreApi();
  const [msg, setMsg] = useState<Message>();

  // We debounce the message preview to prevent it from updating too often.
  useEffect(() => {
    const update = debounce(() => setMsg(getMessage(store)), 250);

    update();
    return store.subscribe(update);
  }, [store]);

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
