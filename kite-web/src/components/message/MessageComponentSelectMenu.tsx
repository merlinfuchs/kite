import { memo } from "react";
import { ChevronDownIcon } from "lucide-react";
import {
  NodeId,
  SelectMenuNode,
  SelectOptionNode,
  isEntitySelect,
  slotLimit,
} from "@/lib/message/document";
import { MessageComponentSelectMenuType } from "@/lib/message/schema";
import {
  channelTypeOptions,
  selectMenuType,
  selectMenuTypes,
} from "@/lib/message/selectMenu";
import {
  useChildIds,
  useDocumentStoreApi,
  useNode,
  useNodeActions,
} from "@/lib/message/state";
import { nodeField, nodeScope, slotScope } from "@/lib/message/validationStore";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageComponentCard from "./MessageComponentCard";
import MessageComponentFlow from "./MessageComponentFlow";
import MessageEmojiPicker from "./MessageEmojiPicker";
import MessageInput from "./MessageInput";
import MessageNodeActions from "./MessageNodeActions";

const valueCountOptions = (from: number) =>
  Array.from(
    { length: slotLimit("selectMenu", "options") + 1 - from },
    (_, i) => ({ label: (i + from).toString(), value: (i + from).toString() })
  );

const minValueOptions = valueCountOptions(0);
const maxValueOptions = valueCountOptions(1);

const typeOptions = selectMenuTypes.map((t) => ({
  label: t.label,
  value: t.type.toString(),
}));

export default function MessageComponentSelectMenu({
  id,
  disableFlowEditor,
}: {
  id: NodeId;
  disableFlowEditor?: boolean;
}) {
  const data = useNode<SelectMenuNode>(id);
  const optionIds = useChildIds(id, "options");
  const actions = useNodeActions(id);
  const { update, insert, removeChildren } = useDocumentStoreApi().getState();

  if (!data) return null;

  const type = selectMenuType(data.select_type);
  const entitySelect = isEntitySelect(data);

  const setType = (value: string) => {
    const selectType = parseInt(value, 10) as MessageComponentSelectMenuType;
    update<SelectMenuNode>(id, {
      select_type: selectType === 3 ? undefined : selectType,
      channel_types: selectType === 8 ? data.channel_types : undefined,
    });
    // A string select can't be sent without an option.
    if (selectType === 3 && optionIds.length === 0) {
      insert(id, "options", "end", {
        type: "selectOption",
        label: "",
        value: "",
      });
    }
  };

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={type.label}
        size="md"
        validation={nodeScope(id)}
        className="space-y-3"
        actions={<MessageNodeActions actions={actions} />}
      >
        <MessageInput
          type="select"
          label="Type"
          value={type.type.toString()}
          options={typeOptions}
          placeholder="Select Menu"
          onChange={setType}
        />
        <div className="flex space-x-3">
          <MessageInput
            type="text"
            label="Placeholder"
            maxLength={150}
            value={data.placeholder ?? ""}
            onChange={(v) =>
              update<SelectMenuNode>(id, { placeholder: v || undefined })
            }
            validation={nodeField<SelectMenuNode>(id, "placeholder")}
            placeholders
          />
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Disabled"
              value={data.disabled ?? false}
              onChange={(v) =>
                update<SelectMenuNode>(id, { disabled: v || undefined })
              }
              validation={nodeField<SelectMenuNode>(id, "disabled")}
            />
          </div>
        </div>
        <div className="flex space-x-3">
          <MessageInput
            type="select"
            label="Min Selections"
            value={(data.min_values ?? 1).toString()}
            options={minValueOptions}
            placeholder="1"
            onChange={(v) =>
              update<SelectMenuNode>(id, { min_values: parseInt(v, 10) })
            }
            validation={nodeField<SelectMenuNode>(id, "min_values")}
          />
          <MessageInput
            type="select"
            label="Max Selections"
            value={(data.max_values ?? 1).toString()}
            options={maxValueOptions}
            placeholder="1"
            onChange={(v) =>
              update<SelectMenuNode>(id, { max_values: parseInt(v, 10) })
            }
            validation={nodeField<SelectMenuNode>(id, "max_values")}
          />
        </div>

        {data.select_type === 8 && (
          <ChannelTypesInput
            values={data.channel_types ?? []}
            onChange={(channelTypes) =>
              update<SelectMenuNode>(id, {
                channel_types:
                  channelTypes.length > 0 ? channelTypes : undefined,
              })
            }
          />
        )}

        {entitySelect ? (
          <div className="text-sm text-muted-foreground">
            Discord fills this menu in for you. The flow receives {type.values}{" "}
            as <code>interaction.values</code>.
          </div>
        ) : (
          <MessageCollapsibleSection
            title="Options"
            size="md"
            validation={slotScope(id, "options")}
            className="space-y-3"
          >
            {optionIds.map((optionId) => (
              <MessageComponentSelectOption key={optionId} id={optionId} />
            ))}
            <div className="space-x-3">
              <Button
                size="sm"
                onClick={() =>
                  insert(id, "options", "end", {
                    type: "selectOption",
                    label: "",
                    value: "",
                  })
                }
                disabled={
                  optionIds.length >= slotLimit("selectMenu", "options")
                }
              >
                Add Option
              </Button>
              <Button
                size="sm"
                variant="destructive"
                onClick={() => removeChildren(id, "options")}
              >
                Clear Options
              </Button>
            </div>
          </MessageCollapsibleSection>
        )}

        {!disableFlowEditor && (
          <MessageComponentFlow
            flowSourceId={data.flow_source_id}
            context="component_select_menu"
          />
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}

// Memoized so editing the menu doesn't re-render every option.
const MessageComponentSelectOption = memo(
  function MessageComponentSelectOption({ id }: { id: NodeId }) {
    const data = useNode<SelectOptionNode>(id);
    const { update } = useDocumentStoreApi().getState();

    if (!data) return null;

    return (
      <MessageComponentCard id={id} label="Option">
        <div className="flex space-x-3">
          <MessageEmojiPicker
            emoji={data.emoji}
            onChange={(emoji) => update<SelectOptionNode>(id, { emoji })}
          />
          <MessageInput
            type="text"
            label="Label"
            maxLength={100}
            value={data.label}
            onChange={(label) => update<SelectOptionNode>(id, { label })}
            validation={nodeField<SelectOptionNode>(id, "label")}
            placeholders
          />
        </div>
        <MessageInput
          type="text"
          label="Value"
          placeholder="What the flow receives as interaction.value"
          maxLength={100}
          value={data.value ?? ""}
          onChange={(value) => update<SelectOptionNode>(id, { value })}
          validation={nodeField<SelectOptionNode>(id, "value")}
          placeholders
        />
        <MessageInput
          type="text"
          label="Description"
          maxLength={100}
          value={data.description ?? ""}
          onChange={(v) =>
            update<SelectOptionNode>(id, { description: v || undefined })
          }
          validation={nodeField<SelectOptionNode>(id, "description")}
          placeholders
        />
      </MessageComponentCard>
    );
  }
);

function ChannelTypesInput({
  values,
  onChange,
}: {
  values: number[];
  onChange: (values: number[]) => void;
}) {
  const selected = channelTypeOptions.filter((o) =>
    values.includes(parseInt(o.value, 10))
  );

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">Channel Types</div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="w-full flex items-center">
            <div className="truncate">
              {selected.length > 0
                ? selected.map((o) => o.label).join(", ")
                : "All channel types"}
            </div>
            <ChevronDownIcon className="h-4 w-4 ml-auto flex-none" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56 max-h-[320px] overflow-y-auto">
          {channelTypeOptions.map((o) => {
            const value = parseInt(o.value, 10);
            return (
              <DropdownMenuCheckboxItem
                key={o.value}
                checked={values.includes(value)}
                onSelect={(e) => e.preventDefault()}
                onCheckedChange={(checked) =>
                  onChange(
                    checked
                      ? [...values, value]
                      : values.filter((v) => v !== value)
                  )
                }
              >
                {o.label}
              </DropdownMenuCheckboxItem>
            );
          })}
        </DropdownMenuContent>
      </DropdownMenu>
      <div className="text-sm text-muted-foreground">
        Leave empty to allow all channel types.
      </div>
    </div>
  );
}
