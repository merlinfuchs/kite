import { useAppStateGuildChannels } from "@/lib/hooks/api";
import { useMemo } from "react";
import EntitySelect from "./EntitySelect";

// Text, voice, announcement, stage
const sendableChannelTypes = [0, 2, 5, 13];

export default function ChannelSelect({
  guildId,
  value,
  onChange,
  types = sendableChannelTypes,
  placeholder = "Select channel...",
}: {
  guildId: string | null;
  value: string | null;
  onChange: (value: string | null) => void;
  // The channel types to list, the ones messages can be sent to by default.
  types?: number[];
  placeholder?: string;
}) {
  const allChannels = useAppStateGuildChannels(guildId);
  const channels = useMemo(
    () => allChannels?.flatMap((c) => (c && types.includes(c.type) ? [c] : [])),
    [allChannels, types]
  );

  return (
    <EntitySelect
      items={channels}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      searchPlaceholder="Search channel..."
      emptyText="No channel found."
    />
  );
}
