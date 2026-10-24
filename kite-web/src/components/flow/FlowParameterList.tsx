import { useNodes } from "@xyflow/react";
import { useMemo } from "react";
import { referencedParameters, syncParameters } from "@/lib/flow/parameters";
import {
  useFlowParameters,
  useFlowParametersStore,
} from "@/lib/flow/parametersStore";
import { FlowParameter } from "@/lib/types/flow.gen";
import { useAppStateGuilds } from "@/lib/hooks/api";
import { Trash2Icon } from "lucide-react";
import { ScrollArea } from "../ui/scroll-area";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import GuildSelect from "../common/GuildSelect";
import {
  ParameterInput,
  useTemplateGuild,
} from "../app/TemplateParameterFields";

const parameterTypes = [
  { value: "text", label: "Text" },
  { value: "textarea", label: "Long text" },
  { value: "channel", label: "Channel" },
  { value: "role", label: "Role" },
];

// FlowParameterList is the sidebar tab with the parameters of the flow. Every
// {{params.<key>}} in a block shows up here, so its value is set once instead
// of in each block.
export default function FlowParameterList() {
  const nodes = useNodes();
  const stored = useFlowParameters();
  const store = useFlowParametersStore();
  const guilds = useAppStateGuilds();
  const [guildId, setGuildId] = useTemplateGuild();

  const used = useMemo(() => referencedParameters({ nodes }), [nodes]);
  const parameters = useMemo(
    () => syncParameters({ nodes }, stored),
    [nodes, stored]
  );
  const needsGuild = parameters.some(
    (p) => p.type === "channel" || p.type === "role"
  );

  function update(key: string, change: Partial<FlowParameter>) {
    store
      .getState()
      .setParameters(
        parameters.map((p) => (p.key === key ? { ...p, ...change } : p))
      );
  }

  function remove(key: string) {
    store.getState().setParameters(parameters.filter((p) => p.key !== key));
  }

  return (
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden">
      <div className="p-5 flex-none">
        <div className="text-xl font-bold text-foreground mb-2">Parameters</div>
        <div className="text-muted-foreground text-sm">
          Values your blocks use as{" "}
          <code className="text-foreground">{"{{params.name}}"}</code>. Change
          them here instead of in each block. When the flow is shared, whoever
          adds it fills in their own.
        </div>
      </div>
      <ScrollArea className="flex-1 min-h-0">
        <div className="flex flex-col gap-3 px-4 pb-4">
          {needsGuild && !!guilds && guilds.length > 1 && (
            <div className="space-y-1.5">
              <Label>Server</Label>
              <GuildSelect value={guildId} onChange={setGuildId} />
            </div>
          )}

          {!parameters.length && (
            <div className="text-sm text-muted-foreground bg-muted/40 rounded-md px-4 py-3">
              Type{" "}
              <code className="text-foreground">
                {"{{params.welcome_channel}}"}
              </code>{" "}
              in a block to add a parameter.
            </div>
          )}

          {parameters.map((parameter) => {
            const isUsed = used.includes(parameter.key);
            return (
              <div
                key={parameter.key}
                className="flex flex-col gap-3 bg-muted/40 rounded-md px-4 py-3"
              >
                <div className="flex items-center justify-between gap-2">
                  <code className="text-sm truncate">
                    params.{parameter.key}
                  </code>
                  {!isUsed && (
                    <div className="flex items-center gap-1 flex-none">
                      <span className="text-xs text-muted-foreground">
                        Not used
                      </span>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => remove(parameter.key)}
                      >
                        <Trash2Icon className="h-4 w-4" />
                      </Button>
                    </div>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <div className="space-y-1.5">
                    <Label className="text-xs">Name</Label>
                    <Input
                      value={parameter.label ?? ""}
                      placeholder={parameter.key}
                      onChange={(e) =>
                        update(parameter.key, { label: e.target.value })
                      }
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label className="text-xs">Type</Label>
                    <Select
                      value={parameter.type}
                      onValueChange={(type) =>
                        // A channel isn't a role, so the value doesn't carry
                        // over between types.
                        update(parameter.key, { type, value: "" })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {parameterTypes.map((t) => (
                          <SelectItem key={t.value} value={t.value}>
                            {t.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs">Value</Label>
                  <ParameterInput
                    parameter={parameter}
                    guildId={guildId}
                    value={parameter.value}
                    onChange={(value) => update(parameter.key, { value })}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs">Description</Label>
                  <Input
                    value={parameter.description ?? ""}
                    placeholder="Shown when the flow is shared"
                    onChange={(e) =>
                      update(parameter.key, { description: e.target.value })
                    }
                  />
                </div>
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <Checkbox
                    checked={!!parameter.required}
                    onCheckedChange={(checked) =>
                      update(parameter.key, { required: !!checked })
                    }
                  />
                  Required when the flow is shared
                </label>
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}
