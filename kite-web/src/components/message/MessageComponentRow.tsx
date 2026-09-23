import {
  useChildIds,
  useDocument,
  useDocumentStoreApi,
  useNodeActions,
} from "@/lib/message/state";
import { NodeId } from "@/lib/message/document";
import { nodeScope } from "@/lib/message/validationStore";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import { Button } from "../ui/button";
import MessageComponentButton from "./MessageComponentButton";
import MessageComponentSelectMenu from "./MessageComponentSelectMenu";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentRow({
  rowId,
  disableFlowEditor,
}: {
  rowId: NodeId;
  disableFlowEditor?: boolean;
}) {
  const childIds = useChildIds(rowId, "components");
  const isButtonRow = useDocument((state) =>
    childIds.every((id) => state.nodes[id]?.type === "button")
  );
  const actions = useNodeActions(rowId);
  const { index } = actions;
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  return (
    <Card className="px-4 py-3">
      <MessageCollapsibleSection
        title={`Row ${index + 1}`}
        size="lg"
        validation={nodeScope(rowId)}
        actions={<MessageNodeActions actions={actions} size="lg" />}
        className="space-y-3"
      >
        {isButtonRow ? (
          <>
            {childIds.map((id) => (
              <MessageComponentButton
                key={id}
                buttonId={id}
                disableFlowEditor={disableFlowEditor}
              />
            ))}
            <div className="space-x-3">
              <Button
                onClick={() =>
                  insert(rowId, "components", "end", {
                    type: "button",
                    style: 2,
                    label: "",
                  })
                }
                size="sm"
                disabled={childIds.length >= 5}
              >
                Add Button
              </Button>
              <Button
                onClick={() => removeChildren(rowId, "components")}
                variant="destructive"
                size="sm"
              >
                Clear Buttons
              </Button>
            </div>
          </>
        ) : (
          // A row with a select menu holds nothing else.
          childIds.map((id) => (
            <MessageComponentSelectMenu
              key={id}
              id={id}
              disableFlowEditor={disableFlowEditor}
            />
          ))
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}
