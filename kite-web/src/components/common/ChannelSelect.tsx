import { useAppStateGuildChannels } from "@/lib/hooks/api";
import {
  HashIcon,
  ImageIcon,
  MegaphoneIcon,
  MessagesSquareIcon,
  RadioIcon,
  Volume2Icon,
} from "lucide-react";
import { useMemo } from "react";
import EntitySelect from "./EntitySelect";

// Text, voice, announcement, stage
const sendableChannelTypes = [0, 2, 5, 13];

const channelTypeIcons: Record<number, typeof HashIcon> = {
  0: HashIcon,
  2: Volume2Icon,
  5: MegaphoneIcon,
  13: RadioIcon,
  15: MessagesSquareIcon,
  16: ImageIcon,
};

function channelIcon(type: number) {
  const Icon = channelTypeIcons[type] ?? HashIcon;
  return <Icon className="h-4 w-4" />;
}

export default function ChannelSelect({
  guildId,
  value,
  onChange,
  types = sendableChannelTypes,
  placeholder = "Select channel...",
  emptyText = "No channel found.",
}: {
  guildId: string | null;
  value: string | null;
  onChange: (value: string | null) => void;
  // The channel types to list, the ones messages can be sent to by default.
  types?: number[];
  placeholder?: string;
  emptyText?: string;
}) {
  const allChannels = useAppStateGuildChannels(guildId);
  const channels = useMemo(
    () =>
      allChannels?.flatMap((c) =>
        c && types.includes(c.type) ? [{ ...c, icon: channelIcon(c.type) }] : []
      ),
    [allChannels, types]
  );

  return (
    <EntitySelect
      items={channels}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      searchPlaceholder="Search channel..."
      emptyText={emptyText}
    />
  );
}
