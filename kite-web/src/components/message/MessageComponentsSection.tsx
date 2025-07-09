import { useCurrentMessage } from "@/lib/message/state";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useShallow } from "zustand/react/shallow";
import { Button } from "../ui/button";
import { getUniqueId } from "@/lib/utils";
import { useCallback } from "react";
import MessageComponentRow from "./MessageComponentRow";

export default function MessageComponentsSection({
  disableFlowEditor,
}: {
  disableFlowEditor?: boolean;
}) {
  const components = useCurrentMessage(
    useShallow((state) => state.components.map((e) => e.id))
  );
  const addRow = useCurrentMessage((state) => state.addComponentRow);
  const clearComponents = useCurrentMessage(
    (state) => state.clearComponentRows
  );

  const addButtonRow = useCallback(() => {
    if (components.length >= 5) return;
    addRow({
      id: getUniqueId(),
      type: 1,
      components: [],
    });
  }, [components, addRow]);

  /* const addSelectMenuRow = useCallback(() => {
    if (components.length >= 5) return;
    addRow({
      id: getUniqueId(),
      type: 1,
      components: [
        {
          id: getUniqueId(),
          type: 3,
          options: [],
        },
      ],
    });
  }, [components, addRow]); */

  return (
    <CollapsibleSection
      title="Components"
      valiationPathPrefix="components"
      className="space-y-4"
    >
      {components.map((id, i) => (
        <MessageComponentRow
          key={id}
          rowIndex={i}
          rowId={id}
          disableFlowEditor={disableFlowEditor}
        />
      ))}
      <div className="space-x-3">
        <Button onClick={addButtonRow}>Add Button Row</Button>
        <Button onClick={clearComponents} variant="outline">
          Clear Components
        </Button>
      </div>
    </CollapsibleSection>
  );
}
