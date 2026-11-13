import { useState } from "react";
import { LoaderCircleIcon } from "lucide-react";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

type ModalValue = string | string[] | boolean;

type ModalField = {
  customId: string;
  label: string;
  type: number;
  // Of text inputs, 2 is a paragraph.
  style?: number;
  placeholder?: string;
  options: { label: string; value: string }[];
};

// Component types of Discord's modals.
const textInput = 4;
const stringSelect = 3;
const radioGroup = 21;
const checkboxGroup = 22;
const checkbox = 23;
const entitySelects = [5, 6, 7, 8];

// The modal a flow showed in a test, filled in and submitted like a member
// would, which runs the test again with it. The modal is in Discord's format.
export default function FlowTestModalForm({
  modal,
  pending,
  onSubmit,
}: {
  modal: any;
  pending: boolean;
  onSubmit: (values: Record<string, ModalValue>) => void;
}) {
  const fields = modalFields(modal);
  const [values, setValues] = useState<Record<string, ModalValue>>({});
  const set = (id: string, value: ModalValue) =>
    setValues((v) => ({ ...v, [id]: value }));

  function submit() {
    const res: Record<string, ModalValue> = {};
    for (const field of fields) {
      const value = values[field.customId];
      if (value === undefined || value === "") continue;
      if (entitySelects.includes(field.type) && typeof value === "string") {
        res[field.customId] = value
          .split(",")
          .map((v) => v.trim())
          .filter(Boolean);
      } else {
        res[field.customId] = value;
      }
    }
    onSubmit(res);
  }

  return (
    <div className="flex flex-col gap-3 rounded-md border p-3">
      <div className="text-sm font-medium">{modal?.title}</div>

      {fields.map((field) => {
        const value = values[field.customId];
        let input: React.ReactNode = null;

        if (field.type === textInput) {
          input =
            field.style === 2 ? (
              <Textarea
                value={(value as string) ?? ""}
                placeholder={field.placeholder}
                onChange={(e) => set(field.customId, e.target.value)}
              />
            ) : (
              <Input
                value={(value as string) ?? ""}
                placeholder={field.placeholder}
                onChange={(e) => set(field.customId, e.target.value)}
              />
            );
        } else if (field.type === radioGroup) {
          input = (
            <Select
              value={(value as string) ?? ""}
              onValueChange={(v) => set(field.customId, v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Pick one" />
              </SelectTrigger>
              <SelectContent>
                {field.options.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          );
        } else if (
          field.type === stringSelect ||
          field.type === checkboxGroup
        ) {
          const picked = (value as string[]) ?? [];
          input = (
            <div className="flex flex-col gap-1.5">
              {field.options.map((o) => (
                <label
                  key={o.value}
                  className="flex items-center gap-2 text-sm"
                >
                  <Checkbox
                    checked={picked.includes(o.value)}
                    onCheckedChange={(checked) =>
                      set(
                        field.customId,
                        checked
                          ? [...picked, o.value]
                          : picked.filter((x) => x !== o.value)
                      )
                    }
                  />
                  {o.label}
                </label>
              ))}
            </div>
          );
        } else if (field.type === checkbox) {
          input = (
            <Checkbox
              checked={value === true}
              onCheckedChange={(checked) =>
                set(field.customId, checked === true)
              }
            />
          );
        } else if (entitySelects.includes(field.type)) {
          input = (
            <Input
              value={(value as string) ?? ""}
              placeholder="IDs, separated by commas"
              onChange={(e) => set(field.customId, e.target.value)}
            />
          );
        } else {
          // File uploads can't be tested.
          return null;
        }

        return (
          <div key={field.customId} className="flex flex-col gap-1.5">
            <Label className="text-sm">{field.label}</Label>
            {input}
          </div>
        );
      })}

      <Button onClick={submit} disabled={pending} className="self-start">
        {pending && <LoaderCircleIcon className="h-4 w-4 mr-2 animate-spin" />}
        Submit
      </Button>
    </div>
  );
}

// The inputs of a modal, in labels or in the action rows of older modals.
function modalFields(modal: any): ModalField[] {
  const res: ModalField[] = [];
  const add = (input: any, label: string | undefined) => {
    if (!input?.custom_id) return;
    res.push({
      customId: input.custom_id,
      label: label || input.label || input.custom_id,
      type: input.type,
      style: input.style,
      placeholder: input.placeholder || undefined,
      options: (input.options ?? []).map((o: any) => ({
        label: o.label || o.value,
        value: o.value,
      })),
    });
  };

  for (const c of modal?.components ?? []) {
    if (c.type === 18) {
      add(c.component, c.label);
    } else if (c.type === 1) {
      for (const input of c.components ?? []) add(input, undefined);
    }
  }
  return res;
}
