import { useAppStateGuilds } from "@/lib/hooks/api";
import { useMemo } from "react";
import EntitySelect from "./EntitySelect";

export default function GuildSelect({
  value,
  onChange,
}: {
  value: string | null;
  onChange: (value: string | null) => void;
}) {
  const allGuilds = useAppStateGuilds();
  const guilds = useMemo(
    () => allGuilds?.flatMap((g) => (g ? [g] : [])),
    [allGuilds]
  );

  return (
    <EntitySelect
      items={guilds}
      value={value}
      onChange={onChange}
      placeholder="Select server..."
      searchPlaceholder="Search server..."
      emptyText="No server found."
    />
  );
}
