import { TemplateSetting } from "@/lib/flow/templates";
import { useAppStateGuilds } from "@/lib/hooks/api";
import { useEffect, useState } from "react";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import ChannelSelect from "../common/ChannelSelect";
import GuildSelect from "../common/GuildSelect";
import RoleSelect from "../common/RoleSelect";

// Settings that are picked from the server.
const pickerTypes: TemplateSetting["type"][] = ["channel", "role"];

// useTemplateGuild is the server the channels and roles of the settings are
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

export function TemplateSettingFields({
  settings,
  values,
  onChange,
  guildId,
  onGuildChange,
}: {
  settings: TemplateSetting[];
  values: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
  guildId: string | null;
  onGuildChange: (guildId: string | null) => void;
}) {
  const guilds = useAppStateGuilds();
  const needsGuild = settings.some((s) => pickerTypes.includes(s.type));

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
            The server the settings below are picked from.
          </div>
          <GuildSelect
            value={guildId}
            onChange={(id) => {
              onGuildChange(id);
              // Picks from another server don't exist in this one.
              const next = { ...values };
              for (const s of settings) {
                if (pickerTypes.includes(s.type)) delete next[s.key];
              }
              onChange(next);
            }}
          />
        </div>
      )}
      {settings.map((setting) => (
        <div key={setting.key} className="space-y-1.5">
          <Label htmlFor={`template-setting-${setting.key}`}>
            {setting.label}
            {!setting.required && (
              <span className="text-muted-foreground font-normal ml-1.5">
                optional
              </span>
            )}
          </Label>
          <div className="text-xs text-muted-foreground">
            {setting.description}
          </div>
          <SettingInput
            setting={setting}
            guildId={guildId}
            value={values[setting.key] || ""}
            onChange={(value) => onChange({ ...values, [setting.key]: value })}
          />
        </div>
      ))}
    </div>
  );
}

function SettingInput({
  setting,
  guildId,
  value,
  onChange,
}: {
  setting: TemplateSetting;
  guildId: string | null;
  value: string;
  onChange: (value: string) => void;
}) {
  const id = `template-setting-${setting.key}`;

  switch (setting.type) {
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
