import { format, parseISO } from "date-fns";
import { Message } from "@/lib/message/schema";
import { colorIntToHex } from "@/tools/common/utils/color";

import {
  DiscordEmbed,
  DiscordEmbedDescription,
  DiscordEmbedField,
  DiscordEmbedFields,
  DiscordEmbedFooter,
  DiscordMessage,
  DiscordMessages,
} from "@skyra/discord-components-react";
import MessageMarkdown from "./MessageMarkdown";
import { cn } from "@/lib/utils";

const defaultUsername = "Captain Hook";
const defaultAvatarUrl = "orange";

export default function MessagePreview({
  msg,
  lightTheme,
  reducePadding,
}: {
  msg: Message;
  lightTheme?: boolean;
  reducePadding?: boolean;
}) {
  return (
    <DiscordMessages lightTheme={!!lightTheme} className="min-h-full flex-auto">
      <DiscordMessage
        className={cn("m-0 py-3", reducePadding && "pr-5")}
        lightTheme={!!lightTheme}
        author={msg.username || defaultUsername}
        avatar={msg.avatar_url || defaultAvatarUrl}
        bot
      >
        <MessageMarkdown>{msg.content}</MessageMarkdown>

        {msg.embeds &&
          msg.embeds.map((embed) => {
            const hexColor = embed.color
              ? colorIntToHex(embed.color)
              : "#1f2225";
            let timestamp = "";
            if (embed.timestamp) {
              const date = parseISO(embed.timestamp);
              if (!isNaN(date.getTime())) {
                timestamp = format(date, "dd/MM/yyyy");
              }
            }
            return (
              <DiscordEmbed
                key={embed.id}
                color={hexColor}
                authorName={embed.author?.name}
                authorImage={embed.author?.icon_url}
                authorUrl={embed.author?.url}
                provider={embed.provider?.name}
                embedTitle={embed.title}
                url={embed.url}
                image={embed.image?.url}
                thumbnail={embed.thumbnail?.url}
                slot="embeds"
              >
                {!!embed.description && (
                  <DiscordEmbedDescription slot="description">
                    {embed.description}
                  </DiscordEmbedDescription>
                )}
                {!!embed.fields?.length && (
                  <DiscordEmbedFields slot="fields">
                    {embed.fields.map((field) => (
                      <DiscordEmbedField
                        key={field.id}
                        fieldTitle={field.name}
                        inline={field.inline}
                      >
                        {field.value}
                      </DiscordEmbedField>
                    ))}
                  </DiscordEmbedFields>
                )}
                {(embed.footer?.text || timestamp) && (
                  <DiscordEmbedFooter
                    slot="footer"
                    footerImage={embed.footer?.icon_url}
                    timestamp={timestamp}
                  >
                    {embed.footer?.text}
                  </DiscordEmbedFooter>
                )}
              </DiscordEmbed>
            );
          })}
      </DiscordMessage>
    </DiscordMessages>
  );
}
