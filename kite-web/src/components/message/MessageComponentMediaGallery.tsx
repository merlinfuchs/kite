import { NodeId } from "@/lib/message/document";
import { slotLimit } from "@/lib/message/document";
import {
  useChildIds,
  useDocumentStoreApi,
  useNodeActions,
} from "@/lib/message/state";
import { nodeScope } from "@/lib/message/validationStore";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageComponentCard from "./MessageComponentCard";
import MessageComponentMediaFields from "./MessageComponentMediaFields";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentMediaGallery({ id }: { id: NodeId }) {
  const itemIds = useChildIds(id, "items");
  const actions = useNodeActions(id);
  const { index } = actions;
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  return (
    <Card className="px-4 py-3">
      <MessageCollapsibleSection
        title={`Media Gallery ${index + 1}`}
        size="lg"
        validation={nodeScope(id)}
        className="space-y-3"
        actions={<MessageNodeActions actions={actions} size="lg" />}
      >
        {itemIds.map((itemId) => (
          <MessageComponentMediaGalleryItem key={itemId} id={itemId} />
        ))}
        <div className="space-x-3">
          <Button
            size="sm"
            onClick={() =>
              insert(id, "items", "end", {
                type: "mediaGalleryItem",
                media: { url: "" },
              })
            }
            disabled={itemIds.length >= slotLimit("mediaGallery", "items")}
          >
            Add Item
          </Button>
          <Button
            size="sm"
            variant="destructive"
            onClick={() => removeChildren(id, "items")}
          >
            Clear Items
          </Button>
        </div>
      </MessageCollapsibleSection>
    </Card>
  );
}

function MessageComponentMediaGalleryItem({ id }: { id: NodeId }) {
  return (
    <MessageComponentCard id={id} label="Item">
      <MessageComponentMediaFields id={id} />
    </MessageComponentCard>
  );
}
