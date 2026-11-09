import { Message } from "@/lib/message/schema";
import { useTheme } from "next-themes";
import MessagePreview from "./MessagePreview";

// MessagePreviewBox shows a message like Discord in a box with a border, in
// the theme that's shown, which theme isn't when it follows the system.
export default function MessagePreviewBox({ msg }: { msg: Message }) {
  const { resolvedTheme } = useTheme();

  return (
    // The preview has no padding of its own, and its margins would leave gaps
    // above and below its background.
    <div className="rounded-md border overflow-hidden [&_.discord-messages]:flow-root [&_.discord-messages]:px-4">
      <MessagePreview msg={msg} lightTheme={resolvedTheme === "light"} />
    </div>
  );
}
