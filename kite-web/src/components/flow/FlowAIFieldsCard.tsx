import { composeFieldAnswers } from "@/lib/flow/ai";
import {
  useAppStateGuildChannels,
  useAppStateGuildRoles,
  useAppStateGuilds,
} from "@/lib/hooks/api";
import { FlowAIField } from "@/lib/types/wire.gen";
import { useEffect, useState } from "react";
import ChannelSelect from "../common/ChannelSelect";
import RoleSelect from "../common/RoleSelect";
import GuildSelect from "../common/GuildSelect";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

// Lets the user fill in what the AI asked for, like a channel, instead of
// writing it out.
export default function FlowAIFieldsCard({
  fields,
  disabled,
  onSend,
}: {
  fields: FlowAIField[];
  // No prompts are left.
  disabled?: boolean;
  onSend: (content: string) => void;
}) {
  const [values, setValues] = useState(() => fields.map(getDefault));
  const answers = composeFieldAnswers(fields, values);

  // The server is picked once for all channels and roles.
  const guilds = useAppStateGuilds();
  const [guildId, setGuildId] = useState<string | null>(null);
  const needsGuild = fields.some((f) => pickerTypes.includes(f.type));
  // Most apps are in a single server.
  useEffect(() => {
    if (!guildId && guilds?.length === 1) setGuildId(guilds[0]!.id);
  }, [guildId, guilds]);

  return (
    <div className="rounded-lg border bg-background p-3 space-y-3">
      {needsGuild && guilds && guilds.length > 1 && (
        <div className="space-y-1">
          <div className="text-xs font-medium">Server</div>
          <GuildSelect
            value={guildId}
            onChange={(id) => {
              setGuildId(id);
              setValues((v) =>
                v.map((old, j) =>
                  pickerTypes.includes(fields[j].type) ? "" : old
                )
              );
            }}
          />
        </div>
      )}
      {fields.map((field, i) => (
        <div key={i} className="space-y-1">
          <div className="text-xs font-medium">{field.label}</div>
          <FieldInput
            // Picks of another server are reset.
            key={pickerTypes.includes(field.type) ? guildId : undefined}
            field={field}
            guildId={guildId}
            value={values[i]}
            onChange={(value) =>
              setValues((v) => v.map((old, j) => (j === i ? value : old)))
            }
          />
          {field.description && (
            <div className="text-xs text-muted-foreground">
              {field.description}
            </div>
          )}
        </div>
      ))}

      <Button
        size="sm"
        disabled={disabled || !answers}
        onClick={() => onSend(answers)}
      >
        Send
      </Button>
    </div>
  );
}

// Fields that are picked from the server.
const pickerTypes = ["channel", "category", "role"];

// Only defaults the input can show are used, so nothing hidden is sent.
function getDefault(field: FlowAIField) {
  switch (field.type) {
    case "channel":
    case "category":
    case "role":
      return "";
    case "choice":
      return field.options.includes(field.default) ? field.default : "";
    case "number":
      return isNaN(Number(field.default)) ? "" : field.default;
    default:
      return field.default;
  }
}

function FieldInput({
  field,
  guildId,
  value,
  onChange,
}: {
  field: FlowAIField;
  guildId: string | null;
  value: string;
  onChange: (value: string) => void;
}) {
  switch (field.type) {
    case "channel":
      return <ChannelFieldInput guildId={guildId} onChange={onChange} />;
    case "category":
      return (
        <ChannelFieldInput guildId={guildId} category onChange={onChange} />
      );
    case "role":
      return <RoleFieldInput guildId={guildId} onChange={onChange} />;
    case "choice":
      return (
        <Select value={value || undefined} onValueChange={onChange}>
          <SelectTrigger>
            <SelectValue placeholder="Pick one..." />
          </SelectTrigger>
          <SelectContent>
            {field.options.map((option) => (
              <SelectItem key={option} value={option}>
                {option}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    default:
      return (
        <Input
          type={field.type === "number" ? "number" : "text"}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          maxLength={500}
        />
      );
  }
}

// The values are the name and ID, so the AI can use the ID and the user sees
// what they picked.
function ChannelFieldInput({
  guildId,
  category,
  onChange,
}: {
  guildId: string | null;
  category?: boolean;
  onChange: (value: string) => void;
}) {
  const [channelId, setChannelId] = useState<string | null>(null);
  const channels = useAppStateGuildChannels(guildId);

  return (
    <ChannelSelect
      guildId={guildId}
      types={category ? categoryTypes : undefined}
      placeholder={category ? "Select category..." : undefined}
      value={channelId}
      onChange={(id) => {
        setChannelId(id);
        const channel = channels?.find((c) => c!.id === id);
        onChange(
          !channel
            ? ""
            : category
            ? `${channel.name} (category ID ${channel.id})`
            : `#${channel.name} (channel ID ${channel.id})`
        );
      }}
    />
  );
}

function RoleFieldInput({
  guildId,
  onChange,
}: {
  guildId: string | null;
  onChange: (value: string) => void;
}) {
  const [roleId, setRoleId] = useState<string | null>(null);
  const roles = useAppStateGuildRoles(guildId);

  return (
    <RoleSelect
      guildId={guildId}
      value={roleId}
      onChange={(id) => {
        setRoleId(id);
        const role = roles?.find((r) => r!.id === id);
        onChange(role ? `@${role.name} (role ID ${role.id})` : "");
      }}
    />
  );
}

const categoryTypes = [4];
