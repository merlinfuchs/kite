import { useShallow } from "zustand/react/shallow";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useCurrentMessage } from "@/lib/message/state";
import MessageInput from "./MessageInput";

export default function MessageEmbedAuthor({
  embedId,
  embedIndex,
}: {
  embedId: number;
  embedIndex: number;
}) {
  const [name, setName] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.author?.name,
      state.setEmbedAuthorName,
    ])
  );
  const [url, setUrl] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.author?.url,
      state.setEmbedAuthorUrl,
    ])
  );
  const [iconUrl, setIconUrl] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.author?.icon_url,
      state.setEmbedAuthorIconUrl,
    ])
  );

  return (
    <CollapsibleSection
      title="Author"
      size="md"
      valiationPathPrefix={`embeds.${embedIndex}.author`}
      className="space-y-3"
    >
      <MessageInput
        type="text"
        label="Name"
        maxLength={256}
        value={name || ""}
        onChange={(v) => setName(embedIndex, v)}
        validationPath={`embeds.${embedIndex}.author.name`}
      />
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="URL"
          value={url || ""}
          onChange={(v) => setUrl(embedIndex, v || undefined)}
          validationPath={`embeds.${embedIndex}.author.url`}
        />
        <MessageInput
          type="url"
          label="Icon URL"
          value={iconUrl || ""}
          onChange={(v) => setIconUrl(embedIndex, v || undefined)}
          validationPath={`embeds.${embedIndex}.author.icon_url`}
        />
      </div>
    </CollapsibleSection>
  );
}
