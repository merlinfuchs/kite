import { useAppStateGuilds } from "@/lib/hooks/api";
import { FlowParameter } from "@/lib/types/flow.gen";
import { useEffect, useState } from "react";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import ChannelSelect from "../common/ChannelSelect";
import GuildSelect from "../common/GuildSelect";
import RoleSelect from "../common/RoleSelect";

// ParameterField is a parameter of a template or a flow.
export type ParameterField = Omit<FlowParameter, "value">;

// Parameters that are picked from the server.
const pickerTypes = ["channel", "role"];

export function isPickedParameter(parameter: ParameterField) {
  return pickerTypes.includes(parameter.type);
}

// ParameterServerSelect is the server the channels and roles of the
// parameters are picked from, if the bot is in more than one.
export function ParameterServerSelect({
  parameters,
  guildId,
  onGuildChange,
}: {
  parameters: ParameterField[];
  guildId: string | null;
  // Picks from another server don't exist in this one, so they should be
  // cleared when it changes from one server to another.
  onGuildChange: (guildId: string | null) => void;
}) {
  const guilds = useAppStateGuilds();
  if (!parameters.some(isPickedParameter)) return null;

  if (guilds?.length === 0) {
    return (
      <div className="text-sm text-muted-foreground">
        Add your bot to a server first, so you can pick its channels and roles
        here.
      </div>
    );
  }
  if (!guilds || guilds.length < 2) return null;

  return (
    <div className="space-y-1.5">
      <Label>Server</Label>
      <div className="text-xs text-muted-foreground">
        The server the parameters below are picked from.
      </div>
      <GuildSelect value={guildId} onChange={onGuildChange} />
    </div>
  );
}

// useTemplateGuild is the server the channels and roles of the parameters are
// picked from.
export function useTemplateGuild() {
  const guilds = useAppStateGuilds();
  const [guildId, setGuildId] = useState<string | null>(null);
  // Most apps are in a single server.
  useEffect(() => {
    if (!guildId && guilds?.length === 1) setGuildId(guilds[0]!.id);
  }, [guildId, guilds]);
  return [guildId, setGuildId] as const;
}

export function TemplateParameterFields({
  parameters,
  values,
  onChange,
  guildId,
  onGuildChange,
  notes,
}: {
  parameters: ParameterField[];
  values: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
  guildId: string | null;
  onGuildChange: (guildId: string | null) => void;
  // Shown below the field of the parameter with the key.
  notes?: Record<string, React.ReactNode>;
}) {
  return (
    <div className="space-y-4">
      <ParameterServerSelect
        parameters={parameters}
        guildId={guildId}
        onGuildChange={(id) => {
          onGuildChange(id);
          // Picking the first one keeps the values, they can be from it.
          if (!guildId) return;
          const next = { ...values };
          for (const p of parameters) {
            if (isPickedParameter(p)) delete next[p.key];
          }
          onChange(next);
        }}
      />
      {parameters.map((parameter) => (
        <div key={parameter.key} className="space-y-1.5">
          <Label htmlFor={`parameter-${parameter.key}`}>
            {parameter.label || parameter.key}
            {!parameter.required && (
              <span className="text-muted-foreground font-normal ml-1.5">
                optional
              </span>
            )}
          </Label>
          {parameter.description && (
            <div className="text-xs text-muted-foreground">
              {parameter.description}
            </div>
          )}
          <ParameterInput
            parameter={parameter}
            guildId={guildId}
            value={values[parameter.key] || ""}
            onChange={(value) =>
              onChange({ ...values, [parameter.key]: value })
            }
          />
          {notes?.[parameter.key]}
        </div>
      ))}
    </div>
  );
}

export function ParameterInput({
  parameter,
  guildId,
  value,
  onChange,
}: {
  parameter: ParameterField;
  guildId: string | null;
  value: string;
  onChange: (value: string) => void;
}) {
  const id = `parameter-${parameter.key}`;
  const emptyText = guildId ? undefined : "Pick a server first.";

  switch (parameter.type) {
    case "channel":
      return (
        <ChannelSelect
          guildId={guildId}
          value={value || null}
          onChange={(id) => onChange(id ?? "")}
          emptyText={emptyText}
        />
      );
    case "role":
      return (
        <RoleSelect
          guildId={guildId}
          value={value || null}
          onChange={(id) => onChange(id ?? "")}
          emptyText={emptyText}
        />
      );
    case "textarea":
      return (
        <Textarea
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      );
    default:
      return (
        <Input
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      );
  }
}
