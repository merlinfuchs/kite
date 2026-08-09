import { useCurrentFlow, useCurrentMessage } from "@/lib/message/state";
import { useShallow } from "zustand/react/shallow";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageInput from "./MessageInput";
import { Button } from "../ui/button";
import { getUniqueId } from "@/lib/utils";
import { useCallback } from "react";
import MessageComponentSelectMenuOption from "./MessageComponentSelectMenuOption";

export default function MessageComponentSelectMenu({
  rowIndex,
  compIndex,
  compCount,
  disableFlowEditor,
}: {
  rowIndex: number;
  compIndex: number;
  compCount: number;
  disableFlowEditor?: boolean;
}) {
  const [placeholder, setPlaceholder] = useCurrentMessage(
    useShallow((state) => [
      state.getSelectMenu(rowIndex, compIndex)?.placeholder,
      state.setSelectMenuPlaceholder,
    ])
  );

  const [disabled, setDisabled] = useCurrentMessage(
    useShallow((state) => [
      state.getSelectMenu(rowIndex, compIndex)?.disabled,
      state.setSelectMenuDisabled,
    ])
  );

  const options = useCurrentMessage(
    useShallow(
      (state) =>
        state.getSelectMenu(rowIndex, compIndex)?.options?.map((o) => o.id) ||
        []
    )
  );

  const [addOption, clearOptions] = useCurrentMessage(
    useShallow((state) => [
      state.addSelectMenuOption,
      state.clearSelectMenuOptions,
    ])
  );

  const replaceFlow = useCurrentFlow((s) => s.replaceFlow);


  return (
    <Card className="p-3 border-l-[3px] rounded-l-[5px] border-l-violet-500">
      <MessageCollapsibleSection
        title="Select Menu"
        size="md"
        valiationPathPrefix={`components.${rowIndex}.components.${compIndex}`}
        className="space-y-3"
        animate={false}
        defaultOpen={true}
      >
        <div className="flex space-x-3">
          <div className="w-full">
            <MessageInput
              type="text"
              label="Placeholder"
              maxLength={150}
              value={placeholder || ""}
              onChange={(v) =>
                setPlaceholder(rowIndex, compIndex, v || undefined)
              }
              validationPath={`components.${rowIndex}.components.${compIndex}.placeholder`}
              placeholders
            />
          </div>
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Disabled"
              value={disabled || false}
              onChange={(v) =>
                setDisabled(rowIndex, compIndex, v || undefined)
              }
              validationPath={`components.${rowIndex}.components.${compIndex}.disabled`}
            />
          </div>
        </div>

        <div className="space-y-3">
          <div className="text-sm font-medium text-foreground/80">
            Options ({options.length}/25) — connect each option in the flow editor
          </div>
          {options.map((id, k) => (
            <MessageComponentSelectMenuOption
              key={id}
              rowIndex={rowIndex}
              compIndex={compIndex}
              optionIndex={k}
              optionCount={options.length}
            />
          ))}
          <div className="space-x-3">
            <Button
              onClick={() => {
                const flowSourceId = getUniqueId().toString();
                addOption(rowIndex, compIndex, {
                  id: getUniqueId(),
                  label: "",
                  flow_source_id: flowSourceId,
                });
                replaceFlow(flowSourceId, {
                  nodes: [{
                    id: getUniqueId().toString(),
                    position: { x: 0, y: 0 },
                    data: {},
                    type: "entry_component_button",
                  }],
                  edges: [],
                });
              }}
              size="sm"
              disabled={options.length >= 25}
            >
              Add Option
            </Button>
            {options.length > 0 && (
              <Button
                onClick={() => clearOptions(rowIndex, compIndex)}
                variant="destructive"
                size="sm"
              >
                Clear Options
              </Button>
            )}
          </div>
        </div>
      </MessageCollapsibleSection>
    </Card>
  );
}
