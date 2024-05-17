import { useMemo } from "react";
import { useNodes, useReactFlow, useStoreApi } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import clsx from "clsx";
import {
  DocumentDuplicateIcon,
  TrashIcon,
  XMarkIcon,
} from "@heroicons/react/24/solid";
import { useNodeValues } from "@/lib/flow/nodes";
import { getId } from "@/lib/flow/util";
import { CheckIcon } from "@heroicons/react/20/solid";

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
  name: NameInput,
  description: DescriptionInput,
  event_type: EventTypeInput,
  text_response: TextResponseInput,
  log_level: LogLevelInput,
  log_message: LogMessageInput,
  condition_base_value: ConditionBaseValueInput,
  condition_allow_multiple: ConditionAllowMultipleInput,
  condition_item_mode: ConditionItemModeInput,
  condition_item_value: ConditionItemValueInput,
};

export default function FlowNodeEditor({ nodeId }: Props) {
  const { setNodes } = useReactFlow<NodeData>();
  const store = useStoreApi();

  function close() {
    store.getState().addSelectedNodes([]);
  }

  const nodes = useNodes<NodeData>();

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
      id: getId(),
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
    <div className="absolute top-0 left-0 bg-dark-3 w-96 h-full p-5 flex flex-col overflow-y-hidden">
      <div className="flex-none">
        <div className="flex items-start justify-between mb-5">
          <div className="text-xl font-bold text-gray-100">Block Settings</div>
          <XMarkIcon
            className="h-6 w-6 text-gray-300 hover:text-gray-100 cursor-pointer"
            onClick={close}
          />
        </div>
        <div className="mb-5">
          <div className="text-lg font-bold text-gray-100 mb-1">
            {values.defaultTitle}
          </div>
          <div className="text-gray-300">{values.defaultDescription}</div>
        </div>
      </div>
      <div className="space-y-3 flex-auto overflow-y-auto">
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
          <DocumentDuplicateIcon className="h-5 w-5" />
          <div>Duplicate Block</div>
        </button>
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

function TextResponseInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="text"
      title="Text Response"
      value={data.text || ""}
      updateValue={(v) => updateData({ text: v || undefined })}
      errors={errors}
    />
  );
}

function ConditionBaseValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="condition_base_value"
      title="Base Value"
      value={data.condition_base_value || ""}
      updateValue={(v) => updateData({ condition_base_value: v || undefined })}
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
      value={data.condition_item_value || ""}
      updateValue={(v) => updateData({ condition_item_value: v || undefined })}
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
      <div className="font-medium text-gray-100 mb-2">{title}</div>
      {description ? (
        <div className="text-gray-300 text-sm mb-2">{description}</div>
      ) : null}
      {type === "textarea" ? (
        <textarea
          className={clsx(
            "px-3 py-2 rounded bg-dark-2 w-full focus:outline-none text-gray-100 min-h-32",
            error ? "border border-red-500" : ""
          )}
          value={value}
          onChange={(e) => updateValue(e.target.value)}
        />
      ) : type === "select" ? (
        <select
          className={clsx(
            "px-3 py-2 rounded bg-dark-2 w-full focus:outline-none text-gray-100",
            error ? "border border-red-500" : ""
          )}
          value={value}
          onChange={(e) => updateValue(e.target.value)}
        >
          <option value=""></option>
          {options?.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      ) : (
        <input
          type="text"
          className={clsx(
            "px-3 py-2 rounded bg-dark-2 w-full focus:outline-none text-gray-100",
            error ? "border border-red-500" : ""
          )}
          value={value}
          onChange={(e) => updateValue(e.target.value)}
        />
      )}
      {error && <div className="text-red-500 text-sm mt-1">{error}</div>}
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
      <div className="font-medium text-gray-100 mb-2">{title}</div>
      {description ? (
        <div className="text-gray-300 text-sm mb-2">{description}</div>
      ) : null}
      <div
        className="h-10 w-10 rounded b-dark-2 text-gray-300 bg-dark-2 flex items-center justify-center"
        onClick={() => updateValue(!value)}
        role="button"
      >
        {value && <CheckIcon className="h-8 w-8" />}
      </div>
      {error && <div className="text-red-500 text-sm mt-1">{error}</div>}
    </div>
  );
}
