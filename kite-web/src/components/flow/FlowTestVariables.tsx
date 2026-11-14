import { useStore } from "@xyflow/react";
import { PlusIcon, XIcon } from "lucide-react";
import { useFlowVariables } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { FlowTestVariableInput } from "@/lib/flow/flowTest";
import EntitySelect from "../common/EntitySelect";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { variableValueTypes } from "../app/VariableValueDialog";

// The values of stored variables a test starts with, in place of the stored
// ones, like a balance to test what happens when it's too low.
export default function FlowTestVariables({
  value,
  onChange,
}: {
  value: FlowTestVariableInput[];
  onChange: (value: FlowTestVariableInput[]) => void;
}) {
  const variables = useFlowVariables();

  const update = (i: number, patch: Partial<FlowTestVariableInput>) =>
    onChange(value.map((v, j) => (j === i ? { ...v, ...patch } : v)));
  const add = () =>
    onChange([
      ...value,
      { variableId: "", scope: "", type: "string", value: "" },
    ]);

  if (!value.length) {
    return (
      <Button
        variant="secondary"
        size="sm"
        className="self-start"
        onClick={add}
      >
        <PlusIcon className="size-4 mr-1.5" />
        Start with stored values
      </Button>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="text-sm font-medium">Stored variables</div>
      <div className="text-sm text-muted-foreground">
        Start the test with other values, like a balance or XP. Changes in the
        test don&apos;t change the stored ones.
      </div>

      {value.map((row, i) => {
        const variable = variables?.find((v) => v.id === row.variableId);

        return (
          <div key={i} className="flex flex-col gap-1.5 rounded-md border p-2">
            <div className="flex items-center gap-2">
              <div className="flex-auto min-w-0">
                <EntitySelect
                  items={variables}
                  value={row.variableId || null}
                  onChange={(id) => update(i, { variableId: id ?? "" })}
                  placeholder="Select variable..."
                  searchPlaceholder="Search variable..."
                  emptyText="No variable found."
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="flex-none size-8"
                onClick={() => onChange(value.filter((_, j) => j !== i))}
              >
                <XIcon className="size-4" />
                <span className="sr-only">Remove</span>
              </Button>
            </div>
            {variable?.scoped && (
              <>
                <Input
                  value={row.scope}
                  onChange={(e) => update(i, { scope: e.target.value })}
                  placeholder="Scope"
                />
                <ScopeHint variableId={row.variableId} />
              </>
            )}
            <div className="flex gap-2">
              <Select
                value={row.type}
                onValueChange={(type) => update(i, { type })}
              >
                <SelectTrigger className="w-32 flex-none">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {variableValueTypes.map((t) => (
                    <SelectItem key={t.value} value={t.value}>
                      {t.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Input
                value={row.value}
                onChange={(e) => update(i, { value: e.target.value })}
                placeholder="Value, empty for none"
              />
            </div>
          </div>
        );
      })}

      <Button
        variant="secondary"
        size="sm"
        className="self-start"
        onClick={add}
      >
        <PlusIcon className="size-4 mr-1.5" />
        Add stored value
      </Button>
    </div>
  );
}

// The scope as the blocks of the flow fill it in, like
// {{guild.id}}_{{user.id}}, as a hint of what to enter.
function ScopeHint({ variableId }: { variableId: string }) {
  const template = useStore(
    (s) =>
      s.nodes.find(
        (n) =>
          (n.data as NodeData).variable_id === variableId &&
          (n.data as NodeData).variable_scope
      )?.data.variable_scope as string | undefined
  );
  if (!template) return null;

  return (
    <div className="text-xs text-muted-foreground break-all">
      The flow uses the scope {template}. Enter it with the IDs filled in.
    </div>
  );
}
