import { format } from "date-fns";
import { useMemo } from "react";
import { hasComponentsV2Flag, Message } from "@/lib/message/schema";
import { cn } from "@/lib/utils";
import { useAssetQueries } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import MessagePreviewComponents, {
  AttachmentUrlContext,
} from "./MessagePreviewComponents";
import MessagePreviewEmbed from "./MessagePreviewEmbed";
import MessagePreviewMarkup from "./MessagePreviewMarkup";

const defaultUsername = "Captain Hook";
const defaultAvatarUrl = "https://cdn.discordapp.com/embed/avatars/3.png";

export default function MessagePreview({
  msg,
  lightTheme,
  reducePadding,
}: {
  msg: Message;
  lightTheme?: boolean;
  reducePadding?: boolean;
}) {
  const currentTime = format(new Date(), "hh:mm aa");
  // A components v2 message carries its content in the components instead.
  const componentsV2 = hasComponentsV2Flag(msg.flags);

  const loadedAssets = useAssetQueries(
    useAppId(),
    msg.attachments.map((a) => a.asset_id)
  );

  // Attachments are uploaded under their asset name, which is what attachment:// refers to.
  const attachmentUrls = useMemo(
    () => new Map(loadedAssets.map((a) => [a.name, a.url])),
    [loadedAssets]
  );

  return (
    <AttachmentUrlContext.Provider value={attachmentUrls}>
      <div
        className={cn(
          "discord-messages min-h-full flex-auto",
          lightTheme ? "discord-light-theme theme-light" : "theme-dark"
        )}
      >
        <div
          className={cn("discord-message m-0 py-3", reducePadding && "pr-5")}
        >
          <div className="discord-message-inner">
            <div className="discord-author-avatar">
              <img src={msg.avatar_url || defaultAvatarUrl} alt="" />
            </div>
            <div className="discord-message-content">
              <span className="discord-author-info">
                <span className="discord-author-username">
                  {msg.username || defaultUsername}
                </span>
                <span className="discord-application-tag">APP</span>
              </span>
              <span className="discord-message-timestamp pl-1">
                Today at {currentTime}
              </span>
              {componentsV2 ? (
                <div className="discord-message-compact-indent">
                  <MessagePreviewComponents components={msg.components} />
                </div>
              ) : (
                <>
                  {!!msg.content && (
                    <div className="discord-message-body">
                      <MessagePreviewMarkup content={msg.content} />
                    </div>
                  )}

                  <div className="discord-message-compact-indent">
                    {msg.embeds.map((embed) => (
                      <MessagePreviewEmbed key={embed.id} embed={embed} />
                    ))}

                    {loadedAssets.length > 0 && (
                      <div className="discord-attachments">
                        {loadedAssets.map((asset) => (
                          <PreviewAttachment key={asset.id} asset={asset} />
                        ))}
                      </div>
                    )}

                    <div className="discord-attachments">
                      <MessagePreviewComponents components={msg.components} />
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </AttachmentUrlContext.Provider>
  );
}

function PreviewAttachment({
  asset,
}: {
  asset: { name: string; url: string; content_type: string };
}) {
  if (asset.content_type.startsWith("image/")) {
    return (
      <img
        src={asset.url}
        alt={asset.name}
        className="max-h-64 max-w-64 rounded"
      />
    );
  }
  if (asset.content_type.startsWith("video/")) {
    return <video src={asset.url} controls className="max-h-64 rounded" />;
  }
  if (asset.content_type.startsWith("audio/")) {
    return <audio src={asset.url} controls />;
  }
  return (
    <a
      href={asset.url}
      target="_blank"
      rel="noreferrer"
      className="discord-component-file"
    >
      {asset.name}
    </a>
  );
}
