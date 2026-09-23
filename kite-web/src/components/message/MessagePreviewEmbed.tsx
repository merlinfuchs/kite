import { format, parseISO } from "date-fns";
import { useMemo } from "react";
import { MessageEmbed } from "@/lib/message/schema";
import { toHTML } from "@/tools/common/utils/discordMarkdown";
import { colorIntToHex } from "@/tools/common/utils/color";
import MessagePreviewMarkup from "./MessagePreviewMarkup";
import MessagePreviewLink from "./MessagePreviewLink";

export default function MessagePreviewEmbed({
  embed,
}: {
  embed: MessageEmbed;
}) {
  let inlineFieldIndex = 0;
  const hexColor = embed.color ? colorIntToHex(embed.color) : "#1f2225";

  let timestamp = "";
  if (embed.timestamp) {
    const date = parseISO(embed.timestamp);
    if (!Number.isNaN(date.getTime())) {
      timestamp = format(date, "dd/MM/yyyy");
    }
  }

  const title = useMemo(
    () => ({ __html: toHTML(embed.title || "", { isTitle: true }) }),
    [embed.title]
  );

  return (
    <div className="discord-embed overflow-hidden">
      <div
        className="discord-left-border"
        style={{ backgroundColor: hexColor }}
      ></div>
      <div className="discord-embed-root">
        <div className="discord-embed-wrapper">
          <div className="discord-embed-grid">
            {!!embed.provider?.name && (
              <div className="discord-embed-provider overflow-hidden break-all">
                {embed.provider.url ? (
                  <MessagePreviewLink href={embed.provider.url}>
                    {embed.provider.name}
                  </MessagePreviewLink>
                ) : (
                  embed.provider.name
                )}
              </div>
            )}
            {!!embed.author?.name && (
              <div className="discord-embed-author overflow-hidden break-all">
                {!!embed.author.icon_url && (
                  <img
                    src={embed.author.icon_url}
                    alt=""
                    className="discord-author-image"
                  />
                )}
                {embed.author.url ? (
                  <MessagePreviewLink href={embed.author.url}>
                    {embed.author.name}
                  </MessagePreviewLink>
                ) : (
                  embed.author.name
                )}
              </div>
            )}
            {!!embed.title && (
              <div className="discord-embed-title overflow-hidden break-all">
                {embed.url ? (
                  <MessagePreviewLink
                    href={embed.url}
                    dangerouslySetInnerHTML={title}
                  />
                ) : (
                  <span dangerouslySetInnerHTML={title} />
                )}
              </div>
            )}
            {!!embed.description && (
              <MessagePreviewMarkup
                content={embed.description}
                className="discord-embed-description"
              />
            )}
            {!!embed.fields.length && (
              <div className="discord-embed-fields">
                {embed.fields.map((field) => (
                  <div
                    key={field.id}
                    className={`discord-embed-field${
                      field.inline
                        ? ` discord-embed-inline-field discord-embed-inline-field-${
                            (inlineFieldIndex++ % 3) + 1
                          }`
                        : ""
                    }`}
                  >
                    <MessagePreviewMarkup
                      content={field.name}
                      className="discord-field-title overflow-hidden break-all"
                      isTitle
                    />
                    <MessagePreviewMarkup content={field.value} className="" />
                  </div>
                ))}
              </div>
            )}
            {!!embed.image?.url && (
              <div className="discord-embed-media">
                <img
                  src={embed.image.url}
                  alt=""
                  className="discord-embed-image"
                />
              </div>
            )}
            {!!embed.thumbnail?.url && (
              <img
                src={embed.thumbnail.url}
                alt=""
                className="discord-embed-thumbnail"
              />
            )}
            {(embed.footer?.text || timestamp) && (
              <div className="discord-embed-footer overflow-hidden break-all">
                {embed.footer?.icon_url && (
                  <img
                    src={embed.footer.icon_url}
                    alt=""
                    className="discord-footer-image"
                  />
                )}
                {embed.footer?.text}
                {embed.footer?.text && timestamp && (
                  <div className="discord-footer-separator">•</div>
                )}
                <div className="flex-none">{timestamp}</div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
