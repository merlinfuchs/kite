import Picker from "@emoji-mart/react";
import { ReactNode, useMemo, useState } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useAppEmojis } from "@/lib/hooks/api";

export type PickerEmoji =
  | {
      native: true;
      name: string;
    }
  | {
      native: false;
      id: string;
      name: string;
      animated: boolean;
    };

interface Props {
  onEmojiSelect: (emoji: PickerEmoji) => void;
  children: ReactNode;
}

export default function EmojiPicker({ onEmojiSelect, children }: Props) {
  const [open, setOpen] = useState(false);

  const appEmojis = useAppEmojis();

  const customEmojis = useMemo(() => {
    if (!appEmojis) return [];

    return [
      {
        id: "custom",
        name: "Custom Emojis",
        emojis: appEmojis.map((emoji) => ({
          id: emoji!.id,
          name: emoji!.name,
          keywords: ["discord", "custom"],
          skins: [
            {
              src: `https://cdn.discordapp.com/emojis/${emoji!.id}.${
                emoji!.animated ? "gif" : "webp"
              }`,
            },
          ],
        })),
      },
    ];
  }, [appEmojis]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="p-0 border-none">
        <Picker
          data={async () => {
            const response = await fetch(
              "https://cdn.jsdelivr.net/npm/@emoji-mart/data/sets/15/twitter.json"
            );
            return response.json();
          }}
          onEmojiSelect={(emoji: any) => {
            if (emoji.native) {
              onEmojiSelect({ native: true, name: emoji.native });
            } else {
              onEmojiSelect({
                native: false,
                id: emoji.id,
                name: emoji.name,
                animated: emoji.src.endsWith(".gif"),
              });
            }
            setOpen(false);
          }}
          custom={customEmojis}
          categories={[
            "frequent",
            "custom",
            "people",
            "nature",
            "foods",
            "activity",
            "places",
            "objects",
            "symbols",
            "flags",
          ]}
          theme="dark"
          set="twitter"
          getSpritesheetURL={() => {
            return "https://cdn.jsdelivr.net/npm/emoji-datasource-twitter@15.0.0/img/twitter/sheets-256/64.png";
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
