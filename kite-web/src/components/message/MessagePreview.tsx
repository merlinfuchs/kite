import { format, parseISO } from "date-fns";
import { Message } from "@/lib/message/schema";
import { colorIntToHex } from "@/tools/common/utils/color";

import {
  DiscordActionRow,
  DiscordAttachments,
  DiscordAudioAttachment,
  DiscordButton,
  DiscordEmbed,
  DiscordEmbedDescription,
  DiscordEmbedField,
  DiscordEmbedFields,
  DiscordEmbedFooter,
  DiscordFileAttachment,
  DiscordImageAttachment,
  DiscordMessage,
  DiscordMessages,
  DiscordVideoAttachment,
} from "@skyra/discord-components-react";
import MessageMarkdown from "./MessageMarkdown";
import { cn } from "@/lib/utils";
import { useAssetQuery } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import { useResponseData } from "@/lib/hooks/api";
import { useMemo } from "react";

const defaultUsername = "Captain Hook";
const defaultAvatarUrl = "orange";

const buttonStyles = {
  1: "primary",
  2: "secondary",
  3: "success",
  4: "destructive",
  5: "secondary",
} as const;

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

        {msg.embeds.map((embed) => {
          const hexColor = embed.color ? colorIntToHex(embed.color) : "#1f2225";
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

        {msg.attachments.length != 0 && (
          <DiscordAttachments slot="attachments">
            {msg.attachments.map((attachment) => (
              <MessagePreviewAttachment
                key={attachment.asset_id}
                assetId={attachment.asset_id}
              />
            ))}
          </DiscordAttachments>
        )}

        {msg.components.length != 0 && (
          <DiscordAttachments slot="components">
            {msg.components.map((row) => (
              <DiscordActionRow key={row.id}>
                {row.components.map((comp) =>
                  comp.type === 2 ? (
                    <DiscordButton
                      key={comp.id}
                      type={buttonStyles[comp.style]}
                      url={comp.style === 5 ? comp.url : undefined}
                      emoji={
                        comp.emoji?.name
                          ? getTwemojiUrl(comp.emoji.name)
                          : undefined
                      }
                      emojiName={comp.emoji?.name}
                      disabled={comp.disabled}
                    >
                      {comp.label}
                    </DiscordButton>
                  ) : null
                )}
              </DiscordActionRow>
            ))}
          </DiscordAttachments>
        )}
      </DiscordMessage>
    </DiscordMessages>
  );
}

function MessagePreviewAttachment({ assetId }: { assetId: string }) {
  const asset = useResponseData(useAssetQuery(useAppId(), assetId));

  const [isImage, isVideo, isAudio] = useMemo(
    () => [
      asset?.content_type.startsWith("image/"),
      asset?.content_type.startsWith("video/"),
      asset?.content_type.startsWith("audio/"),
    ],
    [asset]
  );

  if (!asset) return null;

  if (isImage) {
    return (
      <DiscordImageAttachment
        url={asset.url}
        alt={asset.name}
        height={256}
        width={256}
      />
    );
  } else if (isVideo) {
    return <DiscordVideoAttachment href={asset.url} />;
  } else if (isAudio) {
    return (
      <DiscordAudioAttachment
        name={asset.name}
        href={asset.url}
        bytes={asset.content_size / 1_000_000}
        bytesUnit="MB"
      />
    );
  } else {
    return (
      <DiscordFileAttachment
        name={asset.name}
        bytes={asset.content_size / 1_000_000}
        bytesUnit="MB"
        href={asset.url}
        target="_blank"
        type={asset.content_type}
      />
    );
  }
}

function getTwemojiUrl(emoji: string) {
  const baseUrl =
    "https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/72x72/";

  const codePoints = emoji.codePointAt(0)?.toString(16);
  if (!codePoints) return "";
  return `${baseUrl}${codePoints}.png`;
}
