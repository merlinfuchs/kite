import { Card } from "@/components/ui/card";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useCurrentMessage } from "@/lib/message/state";
import { useShallow } from "zustand/react/shallow";
import { useMemo } from "react";
import { colorIntToHex } from "@/tools/common/utils/color";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  TrashIcon,
} from "lucide-react";
import MessageEmbedBody from "./MessageEmbedBody";
import MessageEmbedAuthor from "./MessageEmbedAuthor";
import MessageEmbedFooter from "./MessageEmbedFooter";
import MessageEmbedImages from "./MessageEmbedImages";
import MessageEmbedFields from "./MessageEmbedFields";

export default function MessageEmbed({
  embedId,
  embedIndex,
}: {
  embedId: number;
  embedIndex: number;
}) {
  const embedName = useCurrentMessage((state) => {
    const embed = state.embeds[embedIndex];
    return embed.author?.name || embed.title;
  });
  const embedCount = useCurrentMessage((state) => state.embeds.length);

  const [moveUp, moveDown, duplicate, remove] = useCurrentMessage(
    useShallow((state) => [
      state.moveEmbedUp,
      state.moveEmbedDown,
      state.duplicateEmbed,
      state.deleteEmbed,
    ])
  );

  const color = useCurrentMessage((state) => state.embeds[embedIndex]?.color);

  const colorHex = useMemo(
    () => (color !== undefined ? colorIntToHex(color) : "#1f2225"),
    [color]
  );

  return (
    <Card
      className="px-4 py-3 border-l-4 rounded-l-sm"
      style={{
        borderLeftColor: colorHex,
      }}
    >
      <CollapsibleSection
        title={`Embed ${embedIndex + 1}`}
        size="lg"
        valiationPathPrefix={`embeds.${embedIndex}`}
        actions={
          <>
            {embedIndex > 0 && (
              <ChevronUpIcon
                className="h-6 w-6"
                onClick={() => moveUp(embedIndex)}
                role="button"
              />
            )}
            {embedIndex < embedCount - 1 && (
              <ChevronDownIcon
                className="h-6 w-6"
                onClick={() => moveDown(embedIndex)}
                role="button"
              />
            )}
            {embedCount < 10 && (
              <CopyIcon
                className="h-5 w-5"
                onClick={() => duplicate(embedIndex)}
                role="button"
              />
            )}
            <TrashIcon
              className="h-5 w-5"
              onClick={() => remove(embedIndex)}
              role="button"
            />
          </>
        }
        className="space-y-5"
      >
        <MessageEmbedAuthor embedIndex={embedIndex} embedId={embedId} />
        <MessageEmbedBody embedIndex={embedIndex} embedId={embedId} />
        <MessageEmbedImages embedIndex={embedIndex} embedId={embedId} />
        <MessageEmbedFooter embedIndex={embedIndex} embedId={embedId} />
        <MessageEmbedFields embedIndex={embedIndex} embedId={embedId} />
      </CollapsibleSection>
    </Card>
  );
}
