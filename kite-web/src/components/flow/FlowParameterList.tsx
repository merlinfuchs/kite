import { useState } from "react";
import {
  needsValue,
  newParameter,
  parameterKeyPattern,
} from "@/lib/flow/parameters";
import {
  useFlowParametersStore,
  useReferencedParameters,
  useSyncedParameters,
} from "@/lib/flow/parametersStore";
import { useAppStateGuilds } from "@/lib/hooks/api";
import { FlowParameter } from "@/lib/types/flow.gen";
import { CopyIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
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
  isPickedParameter,
  ParameterInput,
} from "../app/TemplateParameterFields";

const parameterTypes = [
  { value: "text", label: "Text" },
  { value: "textarea", label: "Long text" },
  { value: "channel", label: "Channel" },
  { value: "role", label: "Role" },
];

// Channels and roles are stored by ID, so a message needs these to mention
// them.
const mentionFormats: Record<string, ((key: string) => string) | undefined> = {
  channel: (key) => `<#{{params.${key}}}>`,
  role: (key) => `<@&{{params.${key}}}>`,
};

// FlowParameterList is the sidebar tab and mobile drawer with the parameters
// of the flow. Every {{params.<key>}} in a block shows up here, so its value is
// set once instead of in each block.
export default function FlowParameterList({
  hideTitle,
}: {
  hideTitle?: boolean;
}) {
  const store = useFlowParametersStore();

  // The server each channel and role is picked from, if the bot is in more
  // than one. A flow can use channels and roles of different servers.
  const guilds = useAppStateGuilds();
  const [guildIds, setGuildIds] = useState<Record<string, string | null>>({});
  const pickGuild = guilds !== undefined && guilds.length > 1;
  function guildOf(key: string) {
    if (guilds?.length === 1) return guilds[0]!.id;
    return guildIds[key] ?? null;
  }

  const used = useReferencedParameters();
  const parameters = useSyncedParameters();

  function changeGuild(parameter: FlowParameter, id: string | null) {
    setGuildIds((ids) => ({ ...ids, [parameter.key]: id }));
    // Picking the first one keeps the value, it can be from it.
    if (guildIds[parameter.key]) update(parameter.key, { value: "" });
  }

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

  const [newKey, setNewKey] = useState("");
  const newKeyError = !newKey
    ? null
    : !parameterKeyPattern.test(newKey)
    ? "Only letters, numbers and underscores, not starting with a number."
    : parameters.some((p) => p.key === newKey)
    ? "The flow already has this parameter."
    : null;

  function add() {
    if (!newKey || newKeyError) return;
    store.getState().setParameters([...parameters, newParameter(newKey)]);
    setNewKey("");
  }

  function copy(key: string) {
    navigator.clipboard
      .writeText(`{{params.${key}}}`)
      .then(() => toast.success("Copied to clipboard!"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  return (
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden">
      <div className="px-5 pt-5 pb-4 flex-none">
        {!hideTitle && (
          <div className="text-xl font-bold text-foreground mb-2">
            Parameters
          </div>
        )}
        <div className="text-muted-foreground text-sm">
          Values your blocks use as{" "}
          <code className="text-foreground">{"{{params.name}}"}</code>. Change
          them here instead of in each block. When the flow is shared, whoever
          adds it fills in their own.
        </div>
      </div>
      {/* pt-1 leaves room for the focus ring, the scroll area would clip it. */}
      <ScrollArea className="flex-1 min-h-0">
        <div className="flex flex-col gap-3 px-4 pt-1 pb-4">
          {guilds?.length === 0 && parameters.some(isPickedParameter) && (
            <div className="text-sm text-muted-foreground">
              Add your bot to a server first, so you can pick its channels and
              roles here.
            </div>
          )}

          <div className="space-y-1.5">
            <div className="flex gap-2">
              <Input
                value={newKey}
                placeholder="welcome_channel"
                onChange={(e) => setNewKey(e.target.value.trim())}
                onKeyDown={(e) => e.key === "Enter" && add()}
              />
              <Button
                variant="outline"
                onClick={add}
                disabled={!newKey || !!newKeyError}
              >
                <PlusIcon className="h-4 w-4 mr-1.5" />
                Add
              </Button>
            </div>
            <div className="text-xs text-muted-foreground">
              {newKeyError ??
                "Add a parameter, or type {{params.name}} in a block."}
            </div>
          </div>

          {parameters.map((parameter) => {
            const isUsed = used.includes(parameter.key);
            const showGuild = pickGuild && isPickedParameter(parameter);
            return (
              <div
                key={parameter.key}
                className="flex flex-col gap-3 bg-muted/40 rounded-md px-4 py-3"
              >
                <div className="flex items-center justify-between gap-2">
                  <button
                    className="flex items-center gap-1.5 min-w-0 text-sm hover:text-primary"
                    title="Copy"
                    onClick={() => copy(parameter.key)}
                  >
                    <code className="truncate">{`{{params.${parameter.key}}}`}</code>
                    <CopyIcon className="h-3.5 w-3.5 flex-none text-muted-foreground" />
                  </button>
                  {isUsed && needsValue(parameter) && (
                    <span className="text-xs text-amber-600 dark:text-amber-500 flex-none">
                      No value
                    </span>
                  )}
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
                  <div
                    className={showGuild ? "grid grid-cols-2 gap-2" : undefined}
                  >
                    {showGuild && (
                      <GuildSelect
                        value={guildOf(parameter.key)}
                        onChange={(id) => changeGuild(parameter, id)}
                      />
                    )}
                    <ParameterInput
                      parameter={parameter}
                      guildId={guildOf(parameter.key)}
                      value={parameter.value}
                      onChange={(value) => update(parameter.key, { value })}
                    />
                  </div>
                  {mentionFormats[parameter.type] && (
                    <div className="text-xs text-muted-foreground">
                      Gives the ID. Use{" "}
                      <code className="text-foreground">
                        {mentionFormats[parameter.type]!(parameter.key)}
                      </code>{" "}
                      to mention it.
                    </div>
                  )}
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
                  Required
                </label>
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}
