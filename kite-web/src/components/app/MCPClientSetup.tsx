import { CopyIcon } from "lucide-react";
import { Button } from "../ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";

// How to connect MCP clients to the server at url with token, as config to
// copy.
export default function MCPClientSetup({
  url,
  token,
  copy,
}: {
  url: string;
  token: string;
  copy: (text: string) => void;
}) {
  const header = `Authorization: Bearer ${token}`;
  const clients = [
    {
      name: "Claude Code",
      hint: "Run this in a terminal:",
      config: `claude mcp add --transport http kite ${url} --header "${header}"`,
    },
    {
      name: "Codex",
      hint: "Run this in a terminal, and add the export line to your shell profile, like ~/.zshrc:",
      config: `export KITE_API_TOKEN=${token}\ncodex mcp add kite --url ${url} --bearer-token-env-var KITE_API_TOKEN`,
    },
    {
      name: "Cursor",
      hint: "Add this to ~/.cursor/mcp.json:",
      config: JSON.stringify(
        {
          mcpServers: {
            kite: { url, headers: { Authorization: `Bearer ${token}` } },
          },
        },
        null,
        2
      ),
    },
    {
      name: "VS Code",
      hint: "Run MCP: Open User Configuration and add this:",
      config: JSON.stringify(
        {
          servers: {
            kite: {
              type: "http",
              url,
              headers: { Authorization: `Bearer ${token}` },
            },
          },
        },
        null,
        2
      ),
    },
    {
      name: "Other",
      hint: "Add an MCP server with Streamable HTTP and this header:",
      config: `${url}\n${header}`,
    },
  ];

  return (
    <Tabs defaultValue={clients[0].name} className="min-w-0">
      <TabsList className="h-auto flex-wrap justify-start">
        {clients.map((client) => (
          <TabsTrigger key={client.name} value={client.name}>
            {client.name}
          </TabsTrigger>
        ))}
      </TabsList>
      {clients.map((client) => (
        <TabsContent
          key={client.name}
          value={client.name}
          className="grid gap-2"
        >
          <div className="text-sm text-muted-foreground">{client.hint}</div>
          <div className="relative">
            <pre className="max-h-48 overflow-auto whitespace-pre-wrap break-all rounded-md border bg-muted p-3 pr-12 font-mono text-xs">
              {client.config}
            </pre>
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-1 top-1"
              onClick={() => copy(client.config)}
              aria-label={`Copy ${client.name} config`}
            >
              <CopyIcon className="size-4" />
            </Button>
          </div>
        </TabsContent>
      ))}
    </Tabs>
  );
}
