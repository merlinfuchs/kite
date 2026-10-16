import { discordEmojiUrl } from "@/tools/common/utils/discordCdn";
import Picker from "@emoji-mart/react";
import { ReactNode, useMemo, useState } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useAppEmojis, useAppStateEmojis } from "@/lib/hooks/api";

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
  const serverEmojis = useAppStateEmojis();

  const customEmojis = useMemo(() => {
    const categories: any[] = [];

    if (appEmojis?.length) {
      categories.push({
        id: "custom",
        name: "Bot Emojis",
        emojis: appEmojis.map((emoji) => ({
          id: emoji!.id,
          name: emoji!.name,
          keywords: ["discord", "custom"],
          skins: [
            {
              src: discordEmojiUrl(emoji!.id, emoji!.animated),
            },
          ],
        })),
      });
    }

    // Emojis of the servers the bot is in. Bots can use them anywhere, as
    // long as they have the Use External Emojis permission where they're sent.
    for (const guild of serverEmojis ?? []) {
      if (!guild || !guild.emojis.length) continue;

      categories.push({
        id: guildCategoryId(guild.guild_id),
        name: guild.guild_name,
        icon: guild.guild_icon_url ? { src: guild.guild_icon_url } : undefined,
        emojis: guild.emojis.map((emoji) => ({
          id: emoji!.id,
          name: emoji!.name,
          keywords: ["discord", "custom", "server", guild.guild_name],
          skins: [
            {
              src: discordEmojiUrl(emoji!.id, emoji!.animated),
            },
          ],
        })),
      });
    }

    return categories;
  }, [appEmojis, serverEmojis]);

  const categories = useMemo(
    () => [
      "frequent",
      "custom",
      ...(serverEmojis ?? [])
        .filter((g) => g && g.emojis.length)
        .map((g) => guildCategoryId(g!.guild_id)),
      "people",
      "nature",
      "foods",
      "activity",
      "places",
      "objects",
      "symbols",
      "flags",
    ],
    [serverEmojis]
  );

  return (
    <Popover open={open} onOpenChange={setOpen} modal>
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
          categories={categories}
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

function guildCategoryId(guildId: string) {
  return `guild_${guildId}`;
}
