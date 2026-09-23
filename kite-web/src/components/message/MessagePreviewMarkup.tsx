import { memo } from "react";
import { toHTML } from "@/tools/common/utils/discordMarkdown";

// Memoized on the strings, so a preview update only re-parses the text that changed.
export default memo(function MessagePreviewMarkup({
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
});
