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
  // Picked channels and roles are stored by ID.
  const [values, setValues] = useState(() => fields.map(getDefault));

  // The server is picked once for all channels and roles.
  const guilds = useAppStateGuilds();
  const [guildId, setGuildId] = useState<string | null>(null);
  const needsGuild = fields.some((f) => pickerTypes.includes(f.type));
  // Most apps are in a single server.
  useEffect(() => {
    if (!guildId && guilds?.length === 1) setGuildId(guilds[0]!.id);
  }, [guildId, guilds]);

  // The AI gets the names and IDs of picked channels and roles.
  const channels = useAppStateGuildChannels(needsGuild ? guildId : null);
  const roles = useAppStateGuildRoles(needsGuild ? guildId : null);
  const answers = composeFieldAnswers(
    fields,
    values.map((value, i) => {
      const type = fields[i].type;
      if (!value || !pickerTypes.includes(type)) return value;
      const item = (type === "role" ? roles : channels)?.find(
        (x) => x?.id === value
      );
      if (!item) return "";
      if (type === "role") return `@${item.name} (role ID ${item.id})`;
      if (type === "category") return `${item.name} (category ID ${item.id})`;
      return `#${item.name} (channel ID ${item.id})`;
    })
  );
  const setValue = (i: number, value: string) =>
    setValues((v) => v.map((old, j) => (j === i ? value : old)));

  return (
    <div className="rounded-lg border bg-background p-3 space-y-3">
      {needsGuild && guilds && guilds.length > 1 && (
        <div className="space-y-1">
          <div className="text-xs font-medium">Server</div>
          <GuildSelect
            value={guildId}
            onChange={(id) => {
              setGuildId(id);
              // Picks of another server are reset.
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
            field={field}
            guildId={guildId}
            value={values[i]}
            onChange={(value) => setValue(i, value)}
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
const categoryTypes = [4];

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
    case "category":
      return (
        <ChannelSelect
          guildId={guildId}
          types={field.type === "category" ? categoryTypes : undefined}
          placeholder={
            field.type === "category" ? "Select category..." : undefined
          }
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
