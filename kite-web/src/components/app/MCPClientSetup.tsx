import { CopyIcon } from "lucide-react";
import { Button } from "../ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";

interface Client {
  name: string;
  hint: string;
  config: string;
}

// How to connect MCP clients to the server at url, as config to copy. Without
// a token, clients sign in to Kite. Clients that can't, like Cursor, need a
// token.
export default function MCPClientSetup({
  url,
  token,
  copy,
}: {
  url: string;
  token?: string;
  copy: (text: string) => void;
}) {
  const clients = token ? tokenClients(url, token) : signInClients(url);

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

function signInClients(url: string): Client[] {
  return [
    {
      name: "Claude",
      hint: "In claude.ai or the Claude app, open Settings, Connectors, Add custom connector, and paste this URL. Claude opens Kite so you can sign in:",
      config: url,
    },
    {
      name: "ChatGPT",
      hint: "Open chatgpt.com/plugins, click +, choose Add custom MCP server and paste this URL. Then mention @Kite in a chat:",
      config: url,
    },
    {
      name: "Claude Code",
      hint: "Run this in a terminal, then /mcp in Claude Code to sign in:",
      config: `claude mcp add --transport http kite ${url}`,
    },
    {
      name: "Codex",
      hint: "Run this in a terminal, or add the URL under Plugins, MCPs in the ChatGPT app. Run codex mcp login kite if it doesn't open Kite to sign in:",
      config: `codex mcp add kite --url ${url}`,
    },
    {
      name: "VS Code",
      hint: "Run MCP: Open User Configuration and add this. VS Code asks you to sign in when it starts Kite:",
      config: JSON.stringify(
        { servers: { kite: { type: "http", url } } },
        null,
        2
      ),
    },
    {
      name: "Other",
      hint: "Add an MCP server with Streamable HTTP and this URL. Clients that can't sign in, like Cursor, need an API token instead:",
      config: url,
    },
  ];
}

function tokenClients(url: string, token: string): Client[] {
  const header = `Authorization: Bearer ${token}`;
  return [
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
}
