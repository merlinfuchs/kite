import { PickerEmoji } from "@/components/common/EmojiPicker";

type TextElement = HTMLInputElement | HTMLTextAreaElement;

// Inserts text at the cursor of an input, replacing the selection if there is
// one, and puts the cursor behind it once the new value has rendered.
export function insertAtCursor(
  element: TextElement | null | undefined,
  text: string,
  onChange: (value: string) => void
) {
  if (!element) return;

  const start = element.selectionStart ?? element.value.length;
  const end = element.selectionEnd ?? start;

  onChange(
    element.value.substring(0, start) + text + element.value.substring(end)
  );

  const cursor = start + text.length;
  requestAnimationFrame(() => {
    try {
      element.setSelectionRange(cursor, cursor);
    } catch {
      // Not every input type has a selection.
    }
  });
}

// The text Discord renders as the emoji: the character itself, or the
// <:name:id> markup for a custom emoji.
export function emojiToText(emoji: PickerEmoji) {
  if (emoji.native) return emoji.name;
  return `<${emoji.animated ? "a" : ""}:${emoji.name}:${emoji.id}>`;
}
