import { composeFieldAnswers } from "@/lib/flow/ai";
import { useAppStateGuildChannels, useAppStateGuilds } from "@/lib/hooks/api";
import { FlowAIField } from "@/lib/types/wire.gen";
import { useEffect, useState } from "react";
import ChannelSelect from "../common/ChannelSelect";
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
  onSend,
}: {
  fields: FlowAIField[];
  onSend: (content: string) => void;
}) {
  const [values, setValues] = useState(() => fields.map(getDefault));
  const answers = composeFieldAnswers(fields, values);

  return (
    <div className="rounded-lg border bg-background p-3 space-y-3">
      {fields.map((field, i) => (
        <div key={i} className="space-y-1">
          <div className="text-xs font-medium">{field.label}</div>
          <FieldInput
            field={field}
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

      <Button size="sm" disabled={!answers} onClick={() => onSend(answers)}>
        Send
      </Button>
    </div>
  );
}

// Only defaults the input can show are used, so nothing hidden is sent.
function getDefault(field: FlowAIField) {
  switch (field.type) {
    case "channel":
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
  value,
  onChange,
}: {
  field: FlowAIField;
  value: string;
  onChange: (value: string) => void;
}) {
  switch (field.type) {
    case "channel":
      return <ChannelFieldInput onChange={onChange} />;
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

// The value is the channel's name and ID, so the AI can use the ID and the
// user sees which channel it is.
function ChannelFieldInput({
  onChange,
}: {
  onChange: (value: string) => void;
}) {
  const guilds = useAppStateGuilds();
  const [guildId, setGuildId] = useState<string | null>(null);
  const [channelId, setChannelId] = useState<string | null>(null);
  const channels = useAppStateGuildChannels(guildId);

  // Most apps are in a single server.
  useEffect(() => {
    if (!guildId && guilds?.length === 1) setGuildId(guilds[0]!.id);
  }, [guildId, guilds]);

  return (
    <div className="space-y-2">
      {guilds && guilds.length > 1 && (
        <GuildSelect
          value={guildId}
          onChange={(id) => {
            setGuildId(id);
            setChannelId(null);
            onChange("");
          }}
        />
      )}
      {/* The AI can ask for categories too, like where tickets go. */}
      <ChannelSelect
        sendableOnly={false}
        guildId={guildId}
        value={channelId}
        onChange={(id) => {
          setChannelId(id);
          const channel = channels?.find((c) => c!.id === id);
          onChange(
            channel ? `#${channel.name} (channel ID ${channel.id})` : ""
          );
        }}
      />
    </div>
  );
}
