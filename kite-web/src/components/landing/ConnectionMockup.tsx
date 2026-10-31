import { BanIcon, PlugIcon } from "lucide-react";

// A simplified connection from the consent page, with what it can do.
export default function ConnectionMockup({
  client,
  origin,
}: {
  client: string;
  origin: string;
}) {
  const settings = [
    { label: "Apps access", value: "Full access" },
    { label: "Which apps", value: "Ticket Bot, Mod Bot" },
  ];
  const blocked = [
    "Mass bans",
    "Bulk deletes",
    "Pruning members",
    "Leaving servers",
  ];

  return (
    <div className="rounded-xl border bg-card p-6 shadow-sm not-prose">
      <div className="flex items-center gap-3">
        <div className="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
          <PlugIcon className="size-5" />
        </div>
        <div>
          <div className="font-semibold">{client}</div>
          <div className="text-sm text-muted-foreground">
            Connected from {origin} · Used 3 minutes ago
          </div>
        </div>
      </div>
      <div className="grid sm:grid-cols-2 gap-3 mt-6">
        {settings.map((setting) => (
          <div key={setting.label}>
            <div className="text-sm font-medium">{setting.label}</div>
            <div className="mt-1.5 rounded-md border bg-background px-3 py-2 text-sm truncate">
              {setting.value}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 border-t pt-5">
        <div className="flex items-center gap-2 text-sm font-medium">
          <BanIcon className="size-4 text-rose-500" />
          Always blocked for clients
        </div>
        <div className="flex flex-wrap gap-2 mt-3">
          {blocked.map((item) => (
            <span
              key={item}
              className="rounded-full bg-rose-500/10 px-3 py-1 text-sm text-rose-500"
            >
              {item}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
