import { useCurrentMessage } from "@/lib/message/state";
import { useShallow } from "zustand/react/shallow";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  TrashIcon,
} from "lucide-react";
import { Button } from "../ui/button";
import { getUniqueId } from "@/lib/utils";
import MessageComponentButton from "./MessageComponentButton";

export default function MessageComponentRow({
  rowIndex,
  rowId,
}: {
  rowIndex: number;
  rowId: number;
}) {
  const rowCount = useCurrentMessage((state) => state.components.length);
  const components = useCurrentMessage(
    useShallow((state) =>
      state.components[rowIndex].components.map((c) => c.id)
    )
  );
  const isButtonRow = useCurrentMessage((state) =>
    state.components[rowIndex].components.every((c) => c.type === 2)
  );
  const [moveUp, moveDown, duplicate, remove] = useCurrentMessage(
    useShallow((state) => [
      state.moveComponentRowUp,
      state.moveComponentRowDown,
      state.duplicateComponentRow,
      state.deleteComponentRow,
    ])
  );

  const [addButton, clearButtons] = useCurrentMessage(
    useShallow((state) => [state.addButton, state.clearButtons])
  );

  return (
    <Card className="px-4 py-3">
      <MessageCollapsibleSection
        title={`Row ${rowIndex + 1}`}
        size="lg"
        valiationPathPrefix={`components.${rowIndex}`}
        actions={
          <>
            {rowIndex > 0 && (
              <ChevronUpIcon
                className="h-6 w-6"
                onClick={() => moveUp(rowIndex)}
                role="button"
              />
            )}
            {rowIndex < rowCount - 1 && (
              <ChevronDownIcon
                className="h-6 w-6"
                onClick={() => moveDown(rowIndex)}
                role="button"
              />
            )}
            {rowCount < 10 && (
              <CopyIcon
                className="h-5 w-5"
                onClick={() => duplicate(rowIndex)}
                role="button"
              />
            )}
            <TrashIcon
              className="h-5 w-5"
              onClick={() => remove(rowIndex)}
              role="button"
            />
          </>
        }
        className="space-y-3"
      >
        {isButtonRow ? (
          <>
            {components.map((id, i) =>
              isButtonRow ? (
                <MessageComponentButton
                  key={id}
                  rowIndex={rowIndex}
                  rowId={rowId}
                  compIndex={i}
                  compId={id}
                ></MessageComponentButton>
              ) : (
                <div key={id}></div>
              )
            )}
            <div className="space-x-3">
              <Button
                onClick={() =>
                  addButton(rowIndex, {
                    id: getUniqueId(),
                    type: 2,
                    style: 2,
                    label: "",
                    flow_source_id: getUniqueId().toString(), // TODO: refactor this for flow_source_id
                  })
                }
                size="sm"
                disabled={components.length >= 5}
              >
                Add Button
              </Button>
              <Button
                onClick={() => clearButtons(rowIndex)}
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
