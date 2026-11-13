import { useAppStateGuildRoles } from "@/lib/hooks/api";
import { useMemo } from "react";
import EntitySelect from "./EntitySelect";

export default function RoleSelect({
  guildId,
  value,
  onChange,
  emptyText = "No role found.",
}: {
  guildId: string | null;
  value: string | null;
  onChange: (value: string | null) => void;
  emptyText?: string;
}) {
  const allRoles = useAppStateGuildRoles(guildId);
  // Everyone and roles of integrations can't be given to members.
  const roles = useMemo(
    () =>
      allRoles
        ?.flatMap((r) => (r && r.id !== guildId && !r.managed ? [r] : []))
        .sort((a, b) => b.position - a.position),
    [allRoles, guildId]
  );

  return (
    <EntitySelect
      items={roles}
      value={value}
      onChange={onChange}
      placeholder="Select role..."
      searchPlaceholder="Search role..."
      emptyText={emptyText}
    />
  );
}
