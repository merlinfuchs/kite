import { toHTML } from "@/tools/common/utils/discordMarkdown";

export default function MessagePreviewMarkup({
  content,
  isTitle,
  className = "discord-message-markup",
}: {
  content: string;
  isTitle?: boolean;
  className?: string;
}) {
  return (
    <div
      className={className}
      dangerouslySetInnerHTML={{ __html: toHTML(content, { isTitle }) }}
    />
  );
}
