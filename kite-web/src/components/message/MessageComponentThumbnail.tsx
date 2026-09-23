import { NodeId } from "@/lib/message/document";
import { useNodeActions } from "@/lib/message/state";
import { nodeScope } from "@/lib/message/validationStore";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageComponentMediaFields from "./MessageComponentMediaFields";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentThumbnail({
  id,
  title = "Thumbnail",
}: {
  id: NodeId;
  title?: string;
}) {
  const actions = useNodeActions(id);

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={title}
        size="md"
        validation={nodeScope(id)}
        className="space-y-3"
        actions={<MessageNodeActions actions={actions} />}
      >
        <MessageComponentMediaFields id={id} />
      </MessageCollapsibleSection>
    </Card>
  );
}
