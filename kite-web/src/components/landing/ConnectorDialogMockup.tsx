import { PlugIcon } from "lucide-react";

// A simplified dialog of an AI client for adding Kite as an MCP server, like
// Claude's "Add custom connector".
export default function ConnectorDialogMockup({
  title,
  button,
}: {
  title: string;
  button: string;
}) {
  const fields = [
    { label: "Name", value: "Kite" },
    { label: "Remote MCP server URL", value: "https://kite.onl/api/mcp" },
  ];

  return (
    <div className="rounded-xl border bg-card p-6 shadow-sm not-prose">
      <div className="flex items-center gap-3">
        <div className="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
          <PlugIcon className="size-5" />
        </div>
        <div className="font-semibold">{title}</div>
      </div>
      <div className="mt-6 space-y-4">
        {fields.map((field) => (
          <div key={field.label}>
            <div className="text-sm font-medium">{field.label}</div>
            <div className="mt-1.5 rounded-md border bg-background px-3 py-2 text-sm font-mono truncate">
              {field.value}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 flex justify-end gap-2">
        <div className="rounded-md border px-4 py-2 text-sm font-medium">
          Cancel
        </div>
        <div className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground">
          {button}
        </div>
      </div>
    </div>
  );
}
