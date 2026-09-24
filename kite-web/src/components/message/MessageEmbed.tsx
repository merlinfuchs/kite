import { Card } from "@/components/ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import { useDocument, useNodeActions } from "@/lib/message/state";
import { EmbedNode, NodeId } from "@/lib/message/document";
import { nodeScope } from "@/lib/message/validationStore";
import { useMemo } from "react";
import { colorIntToHex } from "@/tools/common/utils/color";
import MessageNodeActions from "./MessageNodeActions";
import MessageEmbedBody from "./MessageEmbedBody";
import MessageEmbedAuthor from "./MessageEmbedAuthor";
import MessageEmbedFooter from "./MessageEmbedFooter";
import MessageEmbedImages from "./MessageEmbedImages";
import MessageEmbedFields from "./MessageEmbedFields";

export default function MessageEmbed({ embedId }: { embedId: NodeId }) {
  // Only the color, so typing in the embed doesn't re-render every field below.
  const color = useDocument(
    (state) => (state.nodes[embedId] as EmbedNode | undefined)?.color
  );
  const actions = useNodeActions(embedId);

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
      <MessageCollapsibleSection
        title={`Embed ${actions.index + 1}`}
        size="lg"
        validation={nodeScope(embedId)}
        defaultOpen={false}
        actions={<MessageNodeActions actions={actions} size="lg" />}
        className="space-y-5"
      >
        <MessageEmbedAuthor embedId={embedId} />
        <MessageEmbedBody embedId={embedId} />
        <MessageEmbedImages embedId={embedId} />
        <MessageEmbedFooter embedId={embedId} />
        <MessageEmbedFields embedId={embedId} />
      </MessageCollapsibleSection>
    </Card>
  );
}
