import { useMemo } from "react";
import { Node, useNodes, useReactFlow, useStoreApi } from "@xyflow/react";
import { NodeData, NodeProps } from "../../lib/flow/data";
import { useNodeValues } from "@/lib/flow/nodes";
import { getUniqueId } from "@/lib/utils";
import {
  CheckIcon,
  ChevronDownIcon,
  CircleAlertIcon,
  CopyIcon,
  TrashIcon,
  XIcon,
} from "lucide-react";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Textarea } from "../ui/textarea";
import { Toggle } from "../ui/toggle";
import { Switch } from "../ui/switch";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

interface Props {
  nodeId: string;
}

interface InputProps {
  type: string;
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}

const intputs: Record<string, any> = {
  custom_label: CustomLabelInput,
  result_variable_name: ResultVariableNameInput,
  name: NameInput,
  description: DescriptionInput,
  command_argument_type: CommandArgumentTypeInput,
  command_argument_required: CommandArgumentRequiredInput,
  command_permissions: CommandPermissionsInput,
  event_type: EventTypeInput,
  message_data: MessageDataInput,
  message_ephemeral: MessageEphemeralInput,
  log_level: LogLevelInput,
  log_message: LogMessageInput,
  condition_base_value: ConditionBaseValueInput,
  condition_allow_multiple: ConditionAllowMultipleInput,
  condition_item_mode: ConditionItemModeInput,
  condition_item_value: ConditionItemValueInput,
};

export default function FlowNodeEditor({ nodeId }: Props) {
  const { setNodes } = useReactFlow<Node<NodeProps>>();
  const store = useStoreApi();

  function close() {
    store.getState().addSelectedNodes([]);
  }

  const nodes = useNodes<Node<NodeProps>>();

  const node = nodes.find((n) => n.id === nodeId);

  const data = node?.data;

  function updateData(newData: Partial<NodeData>) {
    setNodes((nodes) =>
      nodes.map((n) => {
        if (n.id === nodeId) {
          return {
            ...n,
            data: {
              ...n.data,
              ...newData,
            },
          };
        }
        return n;
      })
    );
  }

  function deleteNode() {
    setNodes((nodes) => nodes.filter((n) => n.id !== nodeId));
  }

  function duplicateNode() {
    if (!node) return;

    const newNode = {
      ...node,
      id: getUniqueId().toString(),
      selected: false,
      position: {
        x: node?.position.x! + 100,
        y: node?.position.y! + 100,
      },
    };
    setNodes((nodes) => nodes.concat(newNode));
  }

  const values = useNodeValues(node?.type!);

  const errors: Record<string, string> = useMemo(() => {
    if (!values.dataSchema) return {};

    const res = values.dataSchema.safeParse(data);
    if (res.success) {
      return {};
    }

    return Object.fromEntries(
      res.error.issues.map((issue) => [issue.path.join("."), issue.message])
    );
  }, [values.dataSchema, data]);

  if (!node || !data) return null;

  return (
    <div className="absolute top-0 left-0 bg-background w-96 h-full p-5 flex flex-col overflow-y-hidden">
      <div className="flex-none">
        <div className="flex items-start justify-between mb-5">
          <div className="text-xl font-bold text-foreground">
            Block Settings
          </div>
          <XIcon
            className="h-6 w-6 text-muted-foreground hover:text-foreground cursor-pointer"
            onClick={close}
          />
        </div>
        <div className="mb-5">
          <div className="text-lg font-bold text-foreground mb-1">
            {values.defaultTitle}
          </div>
          <div className="text-muted-foreground">
            {values.defaultDescription}
          </div>
        </div>
      </div>
      <div className="space-y-3 flex-auto">
        {values.dataFields.map((field) => {
          const Input = intputs[field];
          if (!Input) return null;

          return (
            <Input
              key={field}
              type={node.type}
              data={data}
              updateData={updateData}
              errors={errors}
            />
          );
        })}
      </div>
      <div className="flex-none space-y-3">
        {!values.fixed && (
          <>
            <button
              className="bg-red-500 hover:bg-red-600 px-3 py-2 w-full rounded text-white font-medium flex space-x-2 justify-center items-center"
              onClick={deleteNode}
            >
              <TrashIcon className="h-5 w-5" />
              <div>Delete Block</div>
            </button>
            <button
              className="bg-dark-5 hover:bg-dark-4 px-3 py-2 w-full rounded text-white font-medium flex space-x-2 justify-center items-center"
              onClick={duplicateNode}
            >
              <CopyIcon className="h-5 w-5" />
              <div>Duplicate Block</div>
            </button>
          </>
        )}
      </div>
    </div>
  );
}

function CustomLabelInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="custom_label"
      title="Custom Label"
      description="Set a custom label for this block so its easier to recognize. This is optional."
      value={data.custom_label || ""}
      updateValue={(v) => updateData({ custom_label: v || undefined })}
      errors={errors}
    />
  );
}

function ResultVariableNameInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="result_variable_name"
      title="Variable Name"
      description="Create a variable to store the result of this action. This is optional."
      value={data.result_variable_name || ""}
      updateValue={(v) => updateData({ result_variable_name: v || undefined })}
      errors={errors}
    />
  );
}

function NameInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="name"
      title="Name"
      value={data.name || ""}
      updateValue={(v) => updateData({ name: v || undefined })}
      errors={errors}
    />
  );
}

function DescriptionInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="description"
      title="Description"
      value={data.description || ""}
      updateValue={(v) => updateData({ description: v || undefined })}
      errors={errors}
    />
  );
}

function CommandArgumentTypeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="command_argument_type"
      title="Argument Type"
      type="select"
      options={[
        { value: "string", label: "Text" },
        { value: "integer", label: "Whole Number" },
        { value: "number", label: "Decimal Number" },
        { value: "boolean", label: "True/False" },
        { value: "user", label: "User" },
        { value: "channel", label: "Channel" },
        { value: "role", label: "Role" },
        { value: "mentionable", label: "Mentionable" },
        { value: "attachment", label: "Attachment" },
      ]}
      value={data.command_argument_type || ""}
      updateValue={(v) => updateData({ command_argument_type: v || undefined })}
      errors={errors}
    />
  );
}

function CommandArgumentRequiredInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseCheckbox
      field="command_argument_required"
      title="Argument Required"
      value={!!data.command_argument_required}
      updateValue={(v) =>
        updateData({ command_argument_required: v || undefined })
      }
      errors={errors}
    />
  );
}

const commandPermissionsOptions = [
  { value: "8", label: "Administrator" },
  { value: "16", label: "Moderator" },
];

function CommandPermissionsInput({ data, updateData, errors }: InputProps) {
  const rawPermissions = data.command_permissions || "0";

  return (
    <BaseMultiSelect
      field="command_permissions"
      title="Permissions"
      value={[]}
      options={commandPermissionsOptions}
      updateValue={(v) => {}}
      errors={errors}
    />
  );
}

function EventTypeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="event_type"
      title="Event"
      options={[
        { value: "DISCORD_MESSAGE_CREATE", label: "Discord Message Create" },
      ]}
      value={data.event_type || ""}
      updateValue={(v) => updateData({ event_type: v || undefined })}
      errors={errors}
    />
  );
}

function LogLevelInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="log_level"
      title="Log Level"
      type="select"
      options={[
        { value: "debug", label: "Debug" },
        { value: "info", label: "Info" },
        { value: "warn", label: "Warn" },
        { value: "error", label: "Error" },
      ]}
      value={data.log_level || ""}
      updateValue={(v) => updateData({ log_level: v || undefined })}
      errors={errors}
    />
  );
}

function LogMessageInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="log_message"
      title="Log Message"
      value={data.log_message || ""}
      updateValue={(v) => updateData({ log_message: v || undefined })}
      errors={errors}
    />
  );
}

function MessageDataInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="message_data"
      title="Text Response"
      value={data.message_data?.content || ""}
      updateValue={(v) =>
        updateData({ message_data: v ? { content: v } : undefined })
      }
      errors={errors}
    />
  );
}

function MessageEphemeralInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="message_ephemeral"
      title="Public Response"
      value={!data.message_ephemeral}
      updateValue={(v) => updateData({ message_ephemeral: !v || undefined })}
      errors={errors}
    />
  );
}

function ConditionBaseValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="condition_base_value"
      title="Base Value"
      value={
        data.condition_base_value ? `${data.condition_base_value.value}` : ""
      }
      updateValue={(v) =>
        updateData({
          condition_base_value: v
            ? {
                type: "string",
                value: v,
              }
            : undefined,
        })
      }
      errors={errors}
    />
  );
}

function ConditionAllowMultipleInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="condition_allow_multiple"
      title="Allow Multiple"
      description="Allow multiple conditions to be met. If disabled, only the first condition that is met will be executed."
      value={data.condition_allow_multiple || false}
      updateValue={(v) =>
        updateData({ condition_allow_multiple: v || undefined })
      }
      errors={errors}
    />
  );
}

function ConditionItemModeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="condition_item_mode"
      title="Comparison Mode"
      options={[
        { value: "equal", label: "Equal" },
        { value: "not_equal", label: "Not Equal" },
        { value: "greater_than", label: "Greater Than" },
        { value: "less_than", label: "Less Than" },
        { value: "greater_than_or_equal", label: "Greater Than or Equal" },
        { value: "less_than_or_equal", label: "Less Than or Equal" },
        { value: "contains", label: "Contains" },
      ]}
      value={data.condition_item_mode || ""}
      updateValue={(v) => updateData({ condition_item_mode: v || undefined })}
      errors={errors}
    />
  );
}

function ConditionItemValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="condition_item_value"
      title="Comparison Value"
      value={
        data.condition_item_value ? `${data.condition_item_value.value}` : ""
      }
      updateValue={(v) =>
        updateData({
          condition_item_value: v
            ? {
                type: "string",
                value: v,
              }
            : undefined,
        })
      }
      errors={errors}
    />
  );
}

function BaseInput({
  type,
  field,
  options,
  title,
  description,
  errors,
  value,
  updateValue,
}: {
  type?: "text" | "textarea" | "select";
  field: string;
  options?: { value: string; label: string }[];
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: string;
  updateValue: (value: string) => void;
}) {
  const error = errors[field];

  return (
    <div>
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      {type === "textarea" ? (
        <Textarea value={value} onChange={(e) => updateValue(e.target.value)} />
      ) : type === "select" ? (
        <Select value={value} onValueChange={(v) => updateValue(v)}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {options?.map((o) => (
              <SelectItem key={o.value} value={o.value}>
                {o.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ) : (
        <Input
          type="text"
          value={value}
          onChange={(e) => updateValue(e.target.value)}
        />
      )}
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

function BaseCheckbox({
  field,
  title,
  description,
  errors,
  value,
  updateValue,
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: boolean;
  updateValue: (value: boolean) => void;
}) {
  const error = errors[field];

  return (
    <div>
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      <Switch checked={value} onCheckedChange={updateValue} />
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

function BaseMultiSelect({
  field,
  title,
  description,
  errors,
  options,
  value,
  updateValue,
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  options: { value: string; label: string }[];
  value: string[];
  updateValue: (value: string[]) => void;
}) {
  const error = errors[field];

  return (
    <div>
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="w-full flex items-center">
            <div>{value.length} selected</div>
            <ChevronDownIcon className="h-4 w-4 ml-auto" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          {options.map((o) => (
            <DropdownMenuCheckboxItem
              key={o.value}
              checked={value.includes(o.value)}
              onCheckedChange={(v) => {
                if (v) {
                  updateValue([...value, o.value]);
                } else {
                  updateValue(value.filter((val) => val !== o.value));
                }
              }}
            >
              {o.label}
            </DropdownMenuCheckboxItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}
