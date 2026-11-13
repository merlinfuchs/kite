import { Node, useNodes } from "@xyflow/react";
import { PlusIcon, XIcon } from "lucide-react";
import { useFlowVariables } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { FlowTestVariableInput } from "@/lib/flow/flowTest";
import EntitySelect from "../common/EntitySelect";
import { Button } from "../ui/button";
import { Input } from "../ui/input";

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
  const nodes = useNodes<Node<NodeData>>();

  const update = (i: number, patch: Partial<FlowTestVariableInput>) =>
    onChange(value.map((v, j) => (j === i ? { ...v, ...patch } : v)));
  const add = () =>
    onChange([...value, { variableId: "", scope: "", value: "" }]);

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
        // The scope as the blocks of the flow fill it in, as a hint of what
        // to enter, like {{guild.id}}_{{user.id}}.
        const scopeTemplate = nodes.find(
          (n) => n.data.variable_id === row.variableId && n.data.variable_scope
        )?.data.variable_scope;

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
                {scopeTemplate && (
                  <div className="text-xs text-muted-foreground break-all">
                    The flow uses the scope {scopeTemplate}. Enter it with the
                    IDs filled in.
                  </div>
                )}
              </>
            )}
            <Input
              value={row.value}
              onChange={(e) => update(i, { value: e.target.value })}
              placeholder="Value, empty for none"
            />
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
