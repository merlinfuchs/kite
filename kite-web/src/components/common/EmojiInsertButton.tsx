import { emojiToText } from "@/lib/insertText";
import { SmileIcon } from "lucide-react";
import EmojiPicker from "./EmojiPicker";

// Emoji button of a text input, shown next to its placeholder button. Fields
// Discord shows as plain text can't render custom emojis, so they pass
// nativeOnly.
export default function EmojiInsertButton({
  onSelect,
  nativeOnly,
}: {
  onSelect: (text: string) => void;
  nativeOnly?: boolean;
}) {
  return (
    <EmojiPicker
      onEmojiSelect={(emoji) => onSelect(emojiToText(emoji))}
      nativeOnly={nativeOnly}
    >
      <SmileIcon
        className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
        role="button"
      />
    </EmojiPicker>
  );
}
