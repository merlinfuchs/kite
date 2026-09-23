import { NodeId, SectionNode, slotLimit } from "@/lib/message/document";
import {
  useChildIds,
  useDocument,
  useDocumentStoreApi,
  useNode,
  useNodeActions,
  useNodeIndex,
} from "@/lib/message/state";
import { nodeScope, slotScope } from "@/lib/message/validationStore";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageComponentEntry from "./MessageComponentEntry";
import MessageComponentTextDisplay from "./MessageComponentTextDisplay";
import MessageInput from "./MessageInput";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentSection({
  id,
  disableFlowEditor,
}: {
  id: NodeId;
  disableFlowEditor?: boolean;
}) {
  const data = useNode<SectionNode>(id);
  const childIds = useChildIds(id, "components");
  const { index } = useNodeIndex(id);
  const actions = useNodeActions(id);
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  const accessoryType = useDocument(
    (state) => state.nodes[data?.accessoryId ?? ""]?.type
  );

  if (!data) return null;

  const setAccessoryType = (type: string) => {
    if (type === accessoryType) return;

    if (type === "button") {
      insert(id, "accessory", "end", { type: "button", style: 2, label: "" });
    } else {
      insert(id, "accessory", "end", {
        type: "thumbnail",
        media: { url: "" },
      });
    }
  };

  return (
    <Card className="px-4 py-3">
      <MessageCollapsibleSection
        title={`Section ${index + 1}`}
        size="lg"
        validation={nodeScope(id)}
        className="space-y-3"
        actions={<MessageNodeActions actions={actions} size="lg" />}
      >
        {childIds.map((childId) => (
          <MessageComponentTextDisplay key={childId} id={childId} />
        ))}
        <div className="space-x-3">
          <Button
            size="sm"
            onClick={() =>
              insert(id, "components", "end", {
                type: "textDisplay",
                content: "",
              })
            }
            disabled={childIds.length >= slotLimit("section", "components")}
          >
            Add Text
          </Button>
          <Button
            size="sm"
            variant="destructive"
            onClick={() => removeChildren(id, "components")}
          >
            Clear Texts
          </Button>
        </div>

        <MessageCollapsibleSection
          title="Accessory"
          size="md"
          validation={slotScope(id, "accessory")}
          className="space-y-3"
        >
          <MessageInput
            type="select"
            label="Type"
            value={accessoryType ?? ""}
            options={[
              { label: "Thumbnail", value: "thumbnail" },
              { label: "Button", value: "button" },
            ]}
            placeholder="Select an accessory"
            onChange={setAccessoryType}
          />
          {data.accessoryId && (
            <MessageComponentEntry
              id={data.accessoryId}
              title={accessoryType === "button" ? "Button" : "Thumbnail"}
              disableFlowEditor={disableFlowEditor}
            />
          )}
        </MessageCollapsibleSection>
      </MessageCollapsibleSection>
    </Card>
  );
}
