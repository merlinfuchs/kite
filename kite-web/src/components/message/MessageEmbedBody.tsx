import { useShallow } from "zustand/react/shallow";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useCurrentMessage } from "@/lib/message/state";
import MessageInput from "./MessageInput";

export default function MessageEmbedBody({
  embedId,
  embedIndex,
}: {
  embedId: number;
  embedIndex: number;
}) {
  const [description, setDescription] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.description,
      state.setEmbedDescription,
    ])
  );
  const [title, setTitle] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.title,
      state.setEmbedTitle,
    ])
  );
  const [url, setUrl] = useCurrentMessage(
    useShallow((state) => [state.embeds[embedIndex]?.url, state.setEmbedUrl])
  );

  const [color, setColor] = useCurrentMessage(
    useShallow((state) => [
      state.embeds[embedIndex]?.color,
      state.setEmbedColor,
    ])
  );

  return (
    <CollapsibleSection
      title="Body"
      size="md"
      valiationPathPrefix={[
        `embeds.${embedIndex}.title`,
        `embeds.${embedIndex}.description`,
        `embeds.${embedIndex}.url`,
        `embeds.${embedIndex}.color`,
      ]}
      className="space-y-3"
    >
      <MessageInput
        type="text"
        label="Title"
        maxLength={256}
        value={title || ""}
        onChange={(v) => setTitle(embedIndex, v || undefined)}
        validationPath={`embeds.${embedIndex}.title`}
      />
      <MessageInput
        type="textarea"
        label="Description"
        maxLength={4000}
        value={description || ""}
        onChange={(v) => setDescription(embedIndex, v || undefined)}
        validationPath={`embeds.${embedIndex}.description`}
      />
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="URL"
          value={url || ""}
          onChange={(v) => setUrl(embedIndex, v || undefined)}
          validationPath={`embeds.${embedIndex}.url`}
        />
        <MessageInput
          type="color"
          label="Color"
          value={color}
          onChange={(v) => setColor(embedIndex, v)}
          validationPath={`embeds.${embedIndex}.color`}
        />
      </div>
    </CollapsibleSection>
  );
}
