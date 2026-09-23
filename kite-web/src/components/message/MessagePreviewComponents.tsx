import { discordEmojiUrl } from "@/tools/common/utils/discordCdn";
import { createContext, ReactNode, useContext, useState } from "react";
import {
  MessageComponent,
  MessageComponentActionRow,
  MessageComponentButton,
  MessageComponentContainer,
  MessageComponentFile,
  MessageComponentMediaGallery,
  MessageComponentSection,
  MessageComponentSeparator,
  MessageComponentThumbnail,
} from "@/lib/message/schema";
import { colorIntToHex } from "@/tools/common/utils/color";
import { cn } from "@/lib/utils";
import Twemoji from "../common/Twemoji";
import MessagePreviewMarkup from "./MessagePreviewMarkup";

/** Resolves `attachment://<name>` media to the uploaded asset, other URLs pass through. */
export const AttachmentUrlContext = createContext<Map<string, string>>(
  new Map()
);

const ATTACHMENT_PREFIX = "attachment://";

function useMediaUrl(url: string): string {
  const attachments = useContext(AttachmentUrlContext);
  if (!url.startsWith(ATTACHMENT_PREFIX)) return url;
  return attachments.get(url.slice(ATTACHMENT_PREFIX.length)) ?? "";
}

// Kept at module level, react-twemoji re-parses its children whenever the options object changes.
const buttonEmojiOptions = { className: "discord-button-emoji" };

const buttonColors = {
  1: "discord-button-primary",
  2: "discord-button-secondary",
  3: "discord-button-success",
  4: "discord-button-destructive",
  5: "discord-button-secondary",
};

function ComponentEmoji({
  emoji,
}: {
  emoji: NonNullable<MessageComponentButton["emoji"]>;
}) {
  if (emoji.id) {
    return (
      <img
        src={discordEmojiUrl(emoji.id, emoji.animated)}
        alt=""
        className="discord-button-emoji"
      />
    );
  }

  return <Twemoji options={buttonEmojiOptions}>{emoji.name}</Twemoji>;
}

function PreviewButton({ button }: { button: MessageComponentButton }) {
  const content = (
    <>
      {button.emoji && <ComponentEmoji emoji={button.emoji} />}
      <span>{button.label}</span>
    </>
  );

  const className = cn(
    "discord-button discord-button-hoverable",
    buttonColors[button.style],
    button.disabled && "discord-button-disabled"
  );

  if (button.style === 5) {
    return (
      <a
        className={className}
        target="_blank"
        href={button.url}
        rel="noreferrer"
      >
        {content}
        <svg
          className="discord-button-launch"
          aria-hidden="true"
          width="16"
          height="16"
          viewBox="0 0 24 24"
        >
          <path
            fill="currentColor"
            d="M10 5V3H5.375C4.06519 3 3 4.06519 3 5.375V18.625C3 19.936 4.06519 21 5.375 21H18.625C19.936 21 21 19.936 21 18.625V14H19V19H5V5H10Z"
          ></path>
          <path
            fill="currentColor"
            d="M21 2.99902H14V4.99902H17.586L9.29297 13.292L10.707 14.706L19 6.41302V9.99902H21V2.99902Z"
          ></path>
        </svg>
      </a>
    );
  }

  return <div className={className}>{content}</div>;
}

function PreviewActionRow({ row }: { row: MessageComponentActionRow }) {
  return (
    <div className="discord-action-row">
      {row.components.map((component) =>
        component.type === 2 ? (
          <PreviewButton key={component.id} button={component} />
        ) : null
      )}
    </div>
  );
}

/** Blurs its content until it is clicked, the way Discord hides a spoiler. */
function Spoiler({
  spoiler,
  children,
}: {
  spoiler?: boolean;
  children: ReactNode;
}) {
  const [revealed, setRevealed] = useState(false);

  if (!spoiler || revealed) return <>{children}</>;

  return (
    <button
      type="button"
      className="discord-component-spoiler"
      aria-label="Reveal spoiler"
      onClick={() => setRevealed(true)}
    >
      {children}
      <span className="discord-component-spoiler-label">Spoiler</span>
    </button>
  );
}

function PreviewThumbnail({
  thumbnail,
}: {
  thumbnail: MessageComponentThumbnail;
}) {
  const url = useMediaUrl(thumbnail.media.url);
  if (!url) return null;

  return (
    <Spoiler spoiler={thumbnail.spoiler}>
      <img
        src={url}
        alt={thumbnail.description ?? ""}
        className="discord-component-thumbnail"
      />
    </Spoiler>
  );
}

function PreviewSection({ section }: { section: MessageComponentSection }) {
  return (
    <div className="discord-component-section">
      <div className="discord-component-section-content">
        {section.components.map((component) => (
          <MessagePreviewMarkup
            key={component.id}
            content={component.content}
          />
        ))}
      </div>
      {/* The editor leaves the accessory out while it is being picked. */}
      {section.accessory && (
        <div className="discord-component-section-accessory">
          {section.accessory.type === 11 ? (
            <PreviewThumbnail thumbnail={section.accessory} />
          ) : (
            <PreviewButton button={section.accessory} />
          )}
        </div>
      )}
    </div>
  );
}

/** How many columns fill evenly, so a row is never left half empty. */
function galleryColumns(count: number): number {
  if (count === 1) return 1;
  if (count === 2 || count === 4) return 2;
  return 3;
}

function PreviewGalleryItem({
  item,
}: {
  item: MessageComponentMediaGallery["items"][number];
}) {
  const url = useMediaUrl(item.media.url);
  if (!url) return null;

  return (
    <Spoiler spoiler={item.spoiler}>
      <img
        src={url}
        alt={item.description ?? ""}
        className="discord-component-gallery-item"
      />
    </Spoiler>
  );
}

/**
 * A simplified gallery: Discord sizes items by their count and aspect ratio,
 * this lays them out in an even grid of up to three columns.
 */
function PreviewMediaGallery({
  gallery,
}: {
  gallery: MessageComponentMediaGallery;
}) {
  const items = gallery.items.filter((item) => item.media.url);
  if (!items.length) return null;

  return (
    <div
      className={cn(
        "discord-component-gallery",
        items.length === 1 && "discord-component-gallery-single"
      )}
      style={{
        gridTemplateColumns: `repeat(${galleryColumns(items.length)}, 1fr)`,
      }}
    >
      {items.map((item) => (
        <PreviewGalleryItem key={item.id} item={item} />
      ))}
    </div>
  );
}

function fileName(url: string): string {
  if (url.startsWith(ATTACHMENT_PREFIX)) {
    return url.slice(ATTACHMENT_PREFIX.length) || "file";
  }

  try {
    const path = new URL(url).pathname;
    return decodeURIComponent(path.slice(path.lastIndexOf("/") + 1)) || "file";
  } catch {
    return url || "file";
  }
}

function PreviewFile({ file }: { file: MessageComponentFile }) {
  return (
    <Spoiler spoiler={file.spoiler}>
      <div className="discord-component-file">
        <svg width="20" height="20" viewBox="0 0 24 24" aria-hidden="true">
          <path
            fill="currentColor"
            d="M13 2H6C4.9 2 4 2.9 4 4V20C4 21.1 4.9 22 6 22H18C19.1 22 20 21.1 20 20V9L13 2ZM13 10V3.5L18.5 10H13Z"
          ></path>
        </svg>
        <span>{fileName(file.file.url)}</span>
      </div>
    </Spoiler>
  );
}

function PreviewSeparator({
  separator,
}: {
  separator: MessageComponentSeparator;
}) {
  return (
    <div
      className={cn(
        "discord-component-separator",
        separator.spacing === 2 && "discord-component-separator-large"
      )}
    >
      {separator.divider && <div className="discord-component-divider" />}
    </div>
  );
}

function PreviewContainer({
  container,
}: {
  container: MessageComponentContainer;
}) {
  // The card stays crisp and its contents are what a spoiler hides.
  return (
    <div
      className="discord-component-container"
      style={{
        borderLeftColor:
          container.accent_color !== undefined
            ? colorIntToHex(container.accent_color)
            : "#4e5058",
      }}
    >
      <Spoiler spoiler={container.spoiler}>
        <MessagePreviewComponents components={container.components} />
      </Spoiler>
    </div>
  );
}

function PreviewComponent({ component }: { component: MessageComponent }) {
  switch (component.type) {
    case 1:
      return <PreviewActionRow row={component} />;
    case 9:
      return <PreviewSection section={component} />;
    case 10:
      return <MessagePreviewMarkup content={component.content} />;
    case 12:
      return <PreviewMediaGallery gallery={component} />;
    case 13:
      return <PreviewFile file={component} />;
    case 14:
      return <PreviewSeparator separator={component} />;
    case 17:
      return <PreviewContainer container={component} />;
    default:
      return null;
  }
}

export default function MessagePreviewComponents({
  components,
}: {
  components: MessageComponent[];
}) {
  return (
    <div className="discord-components">
      {components.map((component) => (
        <PreviewComponent key={component.id} component={component} />
      ))}
    </div>
  );
}
