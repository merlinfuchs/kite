import { Node } from "@xyflow/react";
import { NodeData } from "@/lib/flow/dataSchema";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import ChannelSelect from "../common/ChannelSelect";
import RoleSelect from "../common/RoleSelect";
import { FlowTestField } from "./FlowTestField";

// Text, voice, announcement, stage, forum, media
const allChannelTypes = [0, 2, 5, 13, 15, 16];

// The value of a command argument in a test, picked like in Discord where
// there's a picker for it.
export default function FlowTestArgumentField({
  arg,
  guildId,
  value,
  onChange,
}: {
  arg: Node<NodeData>;
  guildId: string | null;
  value: string;
  onChange: (value: string) => void;
}) {
  const type = arg.data.command_argument_type;
  const choices = (arg.data.command_argument_choices ?? []).flatMap((c) =>
    c.name && c.value ? [{ name: c.name, value: c.value }] : []
  );
  const label = arg.data.command_argument_required
    ? arg.data.name ?? ""
    : `${arg.data.name ?? ""} (optional)`;

  let input: React.ReactNode;
  if (choices.length > 0) {
    input = (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue placeholder="Pick a choice" />
        </SelectTrigger>
        <SelectContent>
          {choices.map((c) => (
            <SelectItem key={c.value} value={c.value}>
              {c.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  } else if (type === "boolean") {
    input = (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue placeholder="Pick one" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="true">True</SelectItem>
          <SelectItem value="false">False</SelectItem>
        </SelectContent>
      </Select>
    );
  } else if (type === "channel") {
    input = (
      <ChannelSelect
        guildId={guildId}
        value={value || null}
        onChange={(v) => onChange(v ?? "")}
        types={allChannelTypes}
      />
    );
  } else if (type === "role") {
    input = (
      <RoleSelect
        guildId={guildId}
        value={value || null}
        onChange={(v) => onChange(v ?? "")}
      />
    );
  } else {
    input = (
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        type={type === "integer" || type === "number" ? "number" : "text"}
        placeholder={argumentPlaceholders[type ?? ""]}
      />
    );
  }

  return <FlowTestField label={label}>{input}</FlowTestField>;
}

const argumentPlaceholders: Record<string, string> = {
  user: "ID of a member",
  mentionable: "ID of a member or role",
  attachment: "Link to a file",
};
