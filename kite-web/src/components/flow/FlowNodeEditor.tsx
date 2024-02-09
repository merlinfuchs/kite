import { ExoticComponent, useMemo } from "react";
import { useNodes, useReactFlow, useStoreApi } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import clsx from "clsx";
import { XMarkIcon } from "@heroicons/react/24/solid";
import { useNodeValues } from "@/lib/flow/nodes";

interface Props {
  nodeId: string;
}

const intputs: Record<string, any> = {
  custom_label: CustomLabelInput,
  name: NameInput,
  description: DescriptionInput,
  text_response: TextResponseInput,
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
    <div className="fixed top-0 left-0 bg-dark-3 w-96 h-full p-5">
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
      <div className="space-y-3">
        {values.dataFields.map((field) => {
          const Input = intputs[field];
          if (!Input) return null;

          return (
            <Input
              key={field}
              data={data}
              updateData={updateData}
              errors={errors}
            />
          );
        })}
      </div>
      {/*<pre className="text-gray-300 mt-5">{JSON.stringify(node, null, 2)}</pre>*/}
    </div>
  );
}

function CustomLabelInput({
  data,
  updateData,
  errors,
}: {
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}) {
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

function NameInput({
  data,
  updateData,
  errors,
}: {
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}) {
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

function DescriptionInput({
  data,
  updateData,
  errors,
}: {
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}) {
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

function TextResponseInput({
  data,
  updateData,
  errors,
}: {
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}) {
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

function BaseInput({
  type,
  field,
  title,
  description,
  errors,
  value,
  updateValue,
}: {
  type?: "text" | "textarea";
  field: string;
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
