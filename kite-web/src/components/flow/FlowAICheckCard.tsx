import { composeCheckedPrompt } from "@/lib/flow/ai";
import { useAppStateGuildChannels, useAppStateGuilds } from "@/lib/hooks/api";
import { FlowAICheckField, FlowAICheckResponse } from "@/lib/types/wire.gen";
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
import { Textarea } from "../ui/textarea";

// Suggests a clearer version of the user's first prompt, with fields for
// what's missing. The user can also send their prompt as it is.
export default function FlowAICheckCard({
  prompt,
  check,
  onSend,
}: {
  prompt: string;
  check: FlowAICheckResponse;
  onSend: (content: string) => void;
}) {
  const [suggested, setSuggested] = useState(check.suggested_prompt || prompt);
  const [values, setValues] = useState(() =>
    check.fields.map((f) => f.default)
  );

  return (
    <div className="rounded-lg border bg-background p-3 space-y-3">
      <p>{check.message}</p>

      <div className="space-y-1">
        <div className="text-xs text-muted-foreground">Suggested request</div>
        <Textarea
          value={suggested}
          onChange={(e) => setSuggested(e.target.value)}
          maxLength={4000}
          minRows={2}
          maxRows={6}
          className="resize-none"
        />
      </div>

      {check.fields.map((field, i) => (
        <div key={i} className="space-y-1">
          <div className="text-xs font-medium">{field.label}</div>
          <CheckFieldInput
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

      <div className="flex flex-wrap gap-2">
        <Button
          size="sm"
          disabled={!suggested.trim()}
          onClick={() =>
            onSend(composeCheckedPrompt(suggested.trim(), check.fields, values))
          }
        >
          Send
        </Button>
        <Button size="sm" variant="ghost" onClick={() => onSend(prompt)}>
          Send my original request
        </Button>
      </div>
    </div>
  );
}

function CheckFieldInput({
  field,
  value,
  onChange,
}: {
  field: FlowAICheckField;
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
        <GuildSelect value={guildId} onChange={setGuildId} />
      )}
      <ChannelSelect
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
