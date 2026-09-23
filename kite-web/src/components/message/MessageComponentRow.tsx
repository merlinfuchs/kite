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
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentRow({
  rowId,
  rowIndex,
  disableFlowEditor,
}: {
  rowId: NodeId;
  rowIndex: number;
  disableFlowEditor?: boolean;
}) {
  const childIds = useChildIds(rowId, "components");
  const isButtonRow = useDocument((state) =>
    childIds.every((id) => state.nodes[id]?.type === "button")
  );
  const actions = useNodeActions(rowId);
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  return (
    <Card className="px-4 py-3">
      <MessageCollapsibleSection
        title={`Row ${rowIndex + 1}`}
        size="lg"
        validation={nodeScope(rowId)}
        actions={<MessageNodeActions actions={actions} size="lg" />}
        className="space-y-3"
      >
        {isButtonRow ? (
          <>
            {childIds.map((id, i) => (
              <MessageComponentButton
                key={id}
                buttonId={id}
                buttonIndex={i}
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
          <div className="text-muted-foreground">
            select menus aren&apos;t supported yet
          </div>
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}
