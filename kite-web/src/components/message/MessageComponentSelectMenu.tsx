import { useCurrentFlow, useCurrentMessage } from "@/lib/message/state";
import { useShallow } from "zustand/react/shallow";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  TrashIcon,
  PlusIcon,
} from "lucide-react";
import MessageInput from "./MessageInput";
import { useCallback, useMemo } from "react";
import MessageEmojiPicker from "./MessageEmojiPicker";
import FlowPreview from "../flow/FlowPreview";
import FlowDialog from "../flow/FlowDialog";
import { FlowData } from "@/lib/flow/dataSchema";
import { getUniqueId } from "@/lib/utils";
import { Button } from "../ui/button";

const initialFlow = {
  nodes: [
    {
      id: getUniqueId().toString(),
      position: { x: 0, y: 0 },
      data: {},
      type: "entry_component_select",
    },
  ],
  edges: [],
};

export default function MessageComponentSelectMenu({
  rowIndex,
  compIndex,
  disableFlowEditor,
}: {
  rowIndex: number;
  compIndex: number;
  disableFlowEditor?: boolean;
}) {
  const selectMenu = useCurrentMessage((state) =>
    state.getSelectMenu(rowIndex, compIndex)
  );

  const [placeholder, setPlaceholder] = useCurrentMessage(
    useShallow((state) => [
      state.getSelectMenu(rowIndex, compIndex)?.placeholder || "",
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
    useShallow((state) =>
      (state.getSelectMenu(rowIndex, compIndex)?.options || []).map((o) => o.id)
    )
  );

  const [
    addOption,
    clearOptions,
    moveOptionUp,
    moveOptionDown,
    duplicateOption,
    deleteOption,
  ] = useCurrentMessage(
    useShallow((state) => [
      state.addSelectMenuOption,
      state.clearSelectMenuOptions,
      state.moveSelectMenuOptionUp,
      state.moveSelectMenuOptionDown,
      state.duplicateSelectMenuOption,
      state.deleteSelectMenuOption,
    ])
  );

  const [setOptionLabel, setOptionDescription, setOptionEmoji] =
    useCurrentMessage(
      useShallow((state) => [
        state.setSelectMenuOptionLabel,
        state.setSelectMenuOptionDescription,
        state.setSelectMenuOptionEmoji,
      ])
    );

  const flowSourceId = useCurrentMessage(
    (state) => state.getSelectMenu(rowIndex, compIndex)?.flow_source_id
  );

  const [flowData, replaceFlow] = useCurrentFlow(
    useShallow((s) => [s.getFlow(flowSourceId || ""), s.replaceFlow])
  );

  const onFlowDialogClose = useCallback(
    (d: FlowData) => {
      if (flowSourceId) {
        replaceFlow(flowSourceId, d);
      }
    },
    [replaceFlow, flowSourceId]
  );

  const handleAddOption = useCallback(() => {
    addOption(rowIndex, compIndex, {
      id: getUniqueId(),
      label: `Option ${options.length + 1}`,
      flow_source_id: getUniqueId().toString(),
    });
  }, [addOption, rowIndex, compIndex, options.length]);

  if (!selectMenu) {
    return <div></div>;
  }

  return (
    <Card
      className="p-3 border-l-[3px] rounded-l-[5px]"
      style={{
        borderLeftColor: "#5865F2",
      }}
    >
      <MessageCollapsibleSection
        title="Select Menu"
        size="md"
        valiationPathPrefix={`components.${rowIndex}.components.${compIndex}`}
        className="space-y-3"
        animate={false}
        defaultOpen={true}
        actions={<></>}
      >
        <div className="flex space-x-3">
          <div className="w-full">
            <MessageInput
              type="text"
              label="Placeholder"
              maxLength={150}
              value={placeholder}
              onChange={(v) => setPlaceholder(rowIndex, compIndex, v || undefined)}
              validationPath={`components.${rowIndex}.components.${compIndex}.placeholder`}
              placeholders
            />
          </div>
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Disabled"
              value={disabled || false}
              onChange={(v) => setDisabled(rowIndex, compIndex, v || undefined)}
              validationPath={`components.${rowIndex}.components.${compIndex}.disabled`}
            />
          </div>
        </div>

        <MessageCollapsibleSection
          title="Options"
          size="sm"
          valiationPathPrefix={`components.${rowIndex}.components.${compIndex}.options`}
          className="space-y-2"
          defaultOpen={true}
        >
          {options.map((optionId, optionIndex) => (
            <SelectMenuOption
              key={optionId}
              rowIndex={rowIndex}
              compIndex={compIndex}
              optionIndex={optionIndex}
              optionCount={options.length}
              onMoveUp={() => moveOptionUp(rowIndex, compIndex, optionIndex)}
              onMoveDown={() => moveOptionDown(rowIndex, compIndex, optionIndex)}
              onDuplicate={() => duplicateOption(rowIndex, compIndex, optionIndex)}
              onDelete={() => deleteOption(rowIndex, compIndex, optionIndex)}
              setLabel={(label) =>
                setOptionLabel(rowIndex, compIndex, optionIndex, label)
              }
              setDescription={(desc) =>
                setOptionDescription(rowIndex, compIndex, optionIndex, desc)
              }
              setEmoji={(emoji) =>
                setOptionEmoji(rowIndex, compIndex, optionIndex, emoji)
              }
            />
          ))}
          <div className="space-x-3">
            <Button
              onClick={handleAddOption}
              size="sm"
              disabled={options.length >= 25}
            >
              <PlusIcon className="h-4 w-4 mr-1" />
              Add Option
            </Button>
            <Button
              onClick={() => clearOptions(rowIndex, compIndex)}
              variant="destructive"
              size="sm"
              disabled={options.length === 0}
            >
              Clear Options
            </Button>
          </div>
        </MessageCollapsibleSection>

        {!disableFlowEditor && (
          <FlowDialog
            flowData={flowData || initialFlow}
            context="component_select"
            onClose={onFlowDialogClose}
          >
            <FlowPreview className="h-64 p-16 w-full" onClick={() => {}} />
          </FlowDialog>
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}

function SelectMenuOption({
  rowIndex,
  compIndex,
  optionIndex,
  optionCount,
  onMoveUp,
  onMoveDown,
  onDuplicate,
  onDelete,
  setLabel,
  setDescription,
  setEmoji,
}: {
  rowIndex: number;
  compIndex: number;
  optionIndex: number;
  optionCount: number;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
  setLabel: (label: string) => void;
  setDescription: (description: string | undefined) => void;
  setEmoji: (emoji: { id?: string; name: string; animated: boolean } | undefined) => void;
}) {
  const option = useCurrentMessage((state) => {
    const menu = state.getSelectMenu(rowIndex, compIndex);
    return menu?.options?.[optionIndex];
  });

  if (!option) {
    return null;
  }

  return (
    <Card className="p-2 bg-muted/50">
      <MessageCollapsibleSection
        title={`Option ${optionIndex + 1}`}
        size="sm"
        valiationPathPrefix={`components.${rowIndex}.components.${compIndex}.options.${optionIndex}`}
        className="space-y-2"
        animate={false}
        defaultOpen={false}
        actions={
          <>
            {optionIndex > 0 && (
              <ChevronUpIcon
                className="h-4 w-4"
                onClick={onMoveUp}
                role="button"
              />
            )}
            {optionIndex < optionCount - 1 && (
              <ChevronDownIcon
                className="h-4 w-4"
                onClick={onMoveDown}
                role="button"
              />
            )}
            {optionCount < 25 && (
              <CopyIcon
                className="h-3 w-3"
                onClick={onDuplicate}
                role="button"
              />
            )}
            <TrashIcon
              className="h-3 w-3"
              onClick={onDelete}
              role="button"
            />
          </>
        }
      >
        <div className="flex space-x-2">
          <MessageEmojiPicker
            emoji={option.emoji}
            onChange={(v) => setEmoji(v)}
          />
          <MessageInput
            type="text"
            label="Label"
            maxLength={100}
            value={option.label}
            onChange={(v) => setLabel(v)}
            validationPath={`components.${rowIndex}.components.${compIndex}.options.${optionIndex}.label`}
          />
        </div>
        <MessageInput
          type="text"
          label="Description"
          maxLength={100}
          value={option.description || ""}
          onChange={(v) => setDescription(v || undefined)}
          validationPath={`components.${rowIndex}.components.${compIndex}.options.${optionIndex}.description`}
        />
      </MessageCollapsibleSection>
    </Card>
  );
}
