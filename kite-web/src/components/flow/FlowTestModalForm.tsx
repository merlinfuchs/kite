import { useMemo, useState } from "react";
import { LoaderCircleIcon } from "lucide-react";
import { importModalData, modalInputIsMultiValue } from "@/lib/flow/modal";
import { splitIds } from "@/lib/flow/flowTest";
import { ModalComponentData } from "@/lib/types/flow.gen";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { FlowTestCheckboxList, FlowTestField } from "./FlowTestField";

type ModalValue = string | string[] | boolean;

const entitySelects = [
  "user_select",
  "role_select",
  "mentionable_select",
  "channel_select",
];

// The modal a flow showed in a test, filled in and submitted like a member
// would, which runs the test again with it. The modal is in Discord's format.
export default function FlowTestModalForm({
  modal,
  pending,
  onSubmit,
}: {
  modal: unknown;
  pending: boolean;
  onSubmit: (values: Record<string, ModalValue>) => void;
}) {
  const data = useMemo(() => {
    try {
      return importModalData(modal).modal;
    } catch {
      return null;
    }
  }, [modal]);
  const [values, setValues] = useState(() => initialValues(data?.components));
  const set = (id: string, value: ModalValue) =>
    setValues((v) => ({ ...v, [id]: value }));

  if (!data) return null;

  function submit() {
    const res: Record<string, ModalValue> = {};
    for (const component of data?.components ?? []) {
      const input = component.components?.[0];
      const value = input?.custom_id ? values[input.custom_id] : undefined;
      if (!input?.custom_id || value === undefined || value === "") continue;
      res[input.custom_id] =
        entitySelects.includes(input.type ?? "") && typeof value === "string"
          ? splitIds(value)
          : value;
    }
    onSubmit(res);
  }

  return (
    <div className="flex flex-col gap-3 rounded-md border p-3">
      <div className="text-sm font-medium">{data.title}</div>

      {data.components?.map((component, i) => {
        if (component.type === "text_display") {
          return (
            <div key={i} className="text-sm text-muted-foreground">
              {component.content}
            </div>
          );
        }
        const input = component.components?.[0];
        if (!input?.custom_id || input.type === "file_upload") return null;

        const id = input.custom_id;
        const label =
          (component.label || id) +
          (input.required && input.type !== "checkbox" ? " *" : "");

        return (
          <FlowTestField key={i} label={label}>
            {component.description && (
              <div className="text-xs text-muted-foreground">
                {component.description}
              </div>
            )}
            <ModalInput
              input={input}
              value={values[id]}
              onChange={(v) => set(id, v)}
            />
          </FlowTestField>
        );
      })}

      <Button onClick={submit} disabled={pending} className="self-start">
        {pending && <LoaderCircleIcon className="h-4 w-4 mr-2 animate-spin" />}
        Submit
      </Button>
    </div>
  );
}

function ModalInput({
  input,
  value,
  onChange,
}: {
  input: ModalComponentData;
  value: ModalValue | undefined;
  onChange: (value: ModalValue) => void;
}) {
  const options = (input.options ?? []).map((o) => ({
    label: o.label || o.value || "",
    value: o.value ?? "",
  }));

  switch (input.type) {
    case "text_input": {
      const props = {
        value: (value as string) ?? "",
        placeholder: input.placeholder,
        onChange: (
          e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        ) => onChange(e.target.value),
      };
      return input.style === 2 ? <Textarea {...props} /> : <Input {...props} />;
    }
    case "checkbox":
      return (
        <Checkbox
          checked={value === true}
          onCheckedChange={(checked) => onChange(checked === true)}
        />
      );
    case "string_select":
    case "radio_group":
    case "checkbox_group":
      // A string select is a list of checkboxes when it takes several
      // options, like a checkbox group.
      if (input.type === "checkbox_group" || modalInputIsMultiValue(input)) {
        return (
          <FlowTestCheckboxList
            options={options}
            value={(value as string[]) ?? []}
            onChange={onChange}
          />
        );
      }
      return (
        <Select
          value={(Array.isArray(value) ? value[0] : (value as string)) ?? ""}
          onValueChange={(v) =>
            onChange(input.type === "radio_group" ? v : [v])
          }
        >
          <SelectTrigger>
            <SelectValue placeholder={input.placeholder || "Pick one"} />
          </SelectTrigger>
          <SelectContent>
            {options.map((o) => (
              <SelectItem key={o.value} value={o.value}>
                {o.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    default:
      return (
        <Input
          value={(value as string) ?? ""}
          placeholder="IDs, separated by commas"
          onChange={(e) => onChange(e.target.value)}
        />
      );
  }
}

// The values the modal starts with, like pre-filled text and options picked
// by default.
function initialValues(components: ModalComponentData[] = []) {
  const res: Record<string, ModalValue> = {};
  for (const component of components) {
    const input = component.components?.[0];
    if (!input?.custom_id) continue;
    const defaults = (input.options ?? [])
      .filter((o) => o.default)
      .map((o) => o.value ?? "");
    if (input.type === "text_input" && input.value) {
      res[input.custom_id] = input.value;
    } else if (input.type === "checkbox" && input.default) {
      res[input.custom_id] = true;
    } else if (input.type === "radio_group" && defaults.length) {
      res[input.custom_id] = defaults[0];
    } else if (defaults.length) {
      res[input.custom_id] = defaults;
    }
  }
  return res;
}
