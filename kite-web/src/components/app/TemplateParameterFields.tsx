import { useAppStateGuilds } from "@/lib/hooks/api";
import { useEffect, useState } from "react";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import ChannelSelect from "../common/ChannelSelect";
import GuildSelect from "../common/GuildSelect";
import RoleSelect from "../common/RoleSelect";

// ParameterField is a parameter of a template or a flow.
export type ParameterField = {
  key: string;
  label?: string;
  description?: string;
  type: string;
  required?: boolean;
};

// Parameters that are picked from the server.
const pickerTypes = ["channel", "role"];

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
}: {
  parameters: ParameterField[];
  values: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
  guildId: string | null;
  onGuildChange: (guildId: string | null) => void;
}) {
  const guilds = useAppStateGuilds();
  const needsGuild = parameters.some((p) => pickerTypes.includes(p.type));

  return (
    <div className="space-y-4">
      {needsGuild && guilds?.length === 0 && (
        <div className="text-sm text-muted-foreground">
          Add your bot to a server first, so you can pick its channels and roles
          here.
        </div>
      )}
      {needsGuild && !!guilds && guilds.length > 1 && (
        <div className="space-y-1.5">
          <Label>Server</Label>
          <div className="text-xs text-muted-foreground">
            The server the parameters below are picked from.
          </div>
          <GuildSelect
            value={guildId}
            onChange={(id) => {
              onGuildChange(id);
              // Picks from another server don't exist in this one.
              const next = { ...values };
              for (const p of parameters) {
                if (pickerTypes.includes(p.type)) delete next[p.key];
              }
              onChange(next);
            }}
          />
        </div>
      )}
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

  switch (parameter.type) {
    case "channel":
      return (
        <ChannelSelect
          guildId={guildId}
          value={value || null}
          onChange={(id) => onChange(id ?? "")}
        />
      );
    case "role":
      return (
        <RoleSelect
          guildId={guildId}
          value={value || null}
          onChange={(id) => onChange(id ?? "")}
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
