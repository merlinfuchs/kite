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
        name: "App Emojis",
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
    // They share one category, because every category gets a tab and only a
    // few tabs fit.
    const guildEmojis = (serverEmojis ?? []).flatMap((guild) =>
      guild!.emojis.map((emoji) => ({
        id: emoji!.id,
        name: emoji!.name,
        keywords: ["discord", "custom", "server", guild!.guild_name],
        skins: [
          {
            src: discordEmojiUrl(emoji!.id, emoji!.animated),
          },
        ],
      }))
    );

    if (guildEmojis.length) {
      categories.push({
        id: "servers",
        name: "Server Emojis",
        icon: { svg: serverIcon },
        emojis: guildEmojis,
      });
    }

    return categories;
  }, [appEmojis, serverEmojis]);

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
          categories={[
            "frequent",
            "custom",
            "servers",
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

// Lucide's server icon
const serverIcon =
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="2" rx="2" ry="2"/><rect width="20" height="8" x="2" y="14" rx="2" ry="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>';
