import { emojiToText } from "@/lib/insertText";
import { SmileIcon } from "lucide-react";
import EmojiPicker from "./EmojiPicker";
import InputActionButton from "./InputActionButton";

// Emoji button of a text input, next to its placeholder button. Fields
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
      <InputActionButton label="Insert emoji" icon={SmileIcon} />
    </EmojiPicker>
  );
}
