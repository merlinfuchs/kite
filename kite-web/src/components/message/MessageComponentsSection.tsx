import {
  useChildIds,
  useDocumentStoreApi,
  useRootId,
} from "@/lib/message/state";
import { slotScope } from "@/lib/message/validationStore";
import CollapsibleSection from "./MessageCollapsibleSection";
import { Button } from "../ui/button";
import MessageComponentRow from "./MessageComponentRow";

export default function MessageComponentsSection({
  disableFlowEditor,
}: {
  disableFlowEditor?: boolean;
}) {
  const rootId = useRootId();
  const rowIds = useChildIds(rootId, "components");
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  return (
    <CollapsibleSection
      title="Components"
      validation={slotScope(rootId, "components")}
      className="space-y-4"
    >
      {rowIds.map((id, i) => (
        <MessageComponentRow
          key={id}
          rowId={id}
          rowIndex={i}
          disableFlowEditor={disableFlowEditor}
        />
      ))}
      <div className="space-x-3">
        <Button
          onClick={() =>
            insert(rootId, "components", "end", { type: "actionRow" })
          }
          disabled={rowIds.length >= 5}
        >
          Add Button Row
        </Button>
        <Button
          onClick={() => removeChildren(rootId, "components")}
          variant="outline"
        >
          Clear Components
        </Button>
      </div>
    </CollapsibleSection>
  );
}
