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
        icon: { svg: appIcon },
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

// Solid icons from Font Awesome Free 6.7.2 (CC BY 4.0), to match the icons of
// the other categories
const appIcon =
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512"><path d="M320 0c17.7 0 32 14.3 32 32l0 64 120 0c39.8 0 72 32.2 72 72l0 272c0 39.8-32.2 72-72 72l-304 0c-39.8 0-72-32.2-72-72l0-272c0-39.8 32.2-72 72-72l120 0 0-64c0-17.7 14.3-32 32-32zM208 384c-8.8 0-16 7.2-16 16s7.2 16 16 16l32 0c8.8 0 16-7.2 16-16s-7.2-16-16-16l-32 0zm96 0c-8.8 0-16 7.2-16 16s7.2 16 16 16l32 0c8.8 0 16-7.2 16-16s-7.2-16-16-16l-32 0zm96 0c-8.8 0-16 7.2-16 16s7.2 16 16 16l32 0c8.8 0 16-7.2 16-16s-7.2-16-16-16l-32 0zM264 256a40 40 0 1 0 -80 0 40 40 0 1 0 80 0zm152 40a40 40 0 1 0 0-80 40 40 0 1 0 0 80zM48 224l16 0 0 192-16 0c-26.5 0-48-21.5-48-48l0-96c0-26.5 21.5-48 48-48zm544 0c26.5 0 48 21.5 48 48l0 96c0 26.5-21.5 48-48 48l-16 0 0-192 16 0z"/></svg>';

const serverIcon =
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path d="M64 32C28.7 32 0 60.7 0 96l0 64c0 35.3 28.7 64 64 64l384 0c35.3 0 64-28.7 64-64l0-64c0-35.3-28.7-64-64-64L64 32zm280 72a24 24 0 1 1 0 48 24 24 0 1 1 0-48zm48 24a24 24 0 1 1 48 0 24 24 0 1 1 -48 0zM64 288c-35.3 0-64 28.7-64 64l0 64c0 35.3 28.7 64 64 64l384 0c35.3 0 64-28.7 64-64l0-64c0-35.3-28.7-64-64-64L64 288zm280 72a24 24 0 1 1 0 48 24 24 0 1 1 0-48zm56 24a24 24 0 1 1 48 0 24 24 0 1 1 -48 0z"/></svg>';
