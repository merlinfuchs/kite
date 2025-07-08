import EmojiPicker, { PickerEmoji } from "@/components/common/EmojiPicker";
import { Emoji } from "@/lib/message/schema";
import Twemoji from "../common/Twemoji";
import { Button } from "../ui/button";
import { SmileIcon, XIcon } from "lucide-react";

interface Props {
  emoji: Emoji | undefined;
  onChange: (emoji: Emoji | undefined) => void;
}

export default function MessageEmojiPicker({ emoji, onChange }: Props) {
  function onEmojiSelect(emoji: PickerEmoji) {
    if (emoji.native) {
      onChange({ name: emoji.name, animated: false });
    } else {
      onChange({ name: emoji.name, id: emoji.id, animated: emoji.animated });
    }
  }

  return (
    <div className="flex-none">
      <div className="mb-2">
        <div className="text-base text-slate-800 dark:text-slate-200">
          Emoji
        </div>
      </div>
      <div className="flex">
        <EmojiPicker onEmojiSelect={onEmojiSelect}>
          <Button size="icon" variant="outline">
            {emoji?.id ? (
              <img
                src={`https://cdn.discordapp.com/emojis/${emoji.id}.${
                  emoji.animated ? "gif" : "webp"
                }`}
                alt=""
                className="h-6 w-6"
              />
            ) : emoji ? (
              <Twemoji
                options={{
                  className: "h-6 w-6",
                }}
              >
                {emoji.name}
              </Twemoji>
            ) : (
              <SmileIcon className="h-6 w-6 text-foreground/80" />
            )}
          </Button>
        </EmojiPicker>
        {emoji && (
          <div
            className="flex items-center cursor-pointer pr-1 text-muted-foreground hover:text-foreground"
            onClick={() => onChange(undefined)}
          >
            <XIcon className="h-5 w-5" />
          </div>
        )}
      </div>
    </div>
  );
}
