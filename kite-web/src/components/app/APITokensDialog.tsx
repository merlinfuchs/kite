import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useState } from "react";
import { Control, useForm } from "react-hook-form";
import { toast } from "sonner";
import { CopyIcon, KeyRoundIcon, TriangleAlertIcon } from "lucide-react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Button } from "../ui/button";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { Skeleton } from "../ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import LoadingButton from "../common/LoadingButton";
import ConfirmDialog from "../common/ConfirmDialog";
import {
  useAPITokenCreateMutation,
  useAPITokenDeleteMutation,
} from "@/lib/api/mutations";
import { useAPITokensQuery, useAppsQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { setValidationErrors } from "@/lib/form";
import { cn, formatDateTime } from "@/lib/utils";
import { isExport } from "@/lib/env/client";
import { track } from "@/lib/analytics";
import { APIToken } from "@/lib/types/wire.gen";

interface FormFields {
  name: string;
  // Days as a string for the select, "never" for no expiry.
  expiresIn: string;
  scope: string;
  // "none" for no access to the Discord API.
  discordScope: string;
  apps: string;
  appIds: string[];
  understood: boolean;
}

interface Option {
  value: string;
  label: string;
  // How tokens with it are described in the list.
  summary?: string;
}

const scopeOptions: Option[] = [
  {
    value: "kite:write",
    label: "Full access",
    summary: "Read and change apps",
  },
  { value: "kite:read", label: "Read only", summary: "Only read apps" },
];

// Write first, like scopeOptions, so tokens are described by their highest.
const discordScopeOptions: Option[] = [
  {
    value: "discord:write",
    label: "Full access",
    summary: "Read and change Discord",
  },
  { value: "discord:read", label: "Read only", summary: "Read Discord" },
  { value: "none", label: "None" },
];

const appsOptions: Option[] = [
  { value: "all", label: "All apps" },
  { value: "some", label: "Some apps" },
];

const expiryOptions: Option[] = [
  { value: "30", label: "30 days" },
  { value: "90", label: "90 days" },
  { value: "365", label: "1 year" },
  { value: "never", label: "Never" },
];

// Lists the user's API tokens and creates new ones. A new token is shown once,
// it can't be read back later.
export default function APITokensDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const tokens = useResponseData(useAPITokensQuery(open));
  const apps = useResponseData(useAppsQuery(open));
  const createMutation = useAPITokenCreateMutation();
  const [createdToken, setCreatedToken] = useState<string | null>(null);

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      expiresIn: "90",
      scope: "kite:write",
      discordScope: "none",
      apps: "all",
      appIds: [],
      understood: false,
    },
  });
  const someApps = form.watch("apps") === "some";
  const understood = form.watch("understood");

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        name: data.name,
        expires_in_days:
          data.expiresIn === "never" ? null : Number(data.expiresIn),
        scopes: [data.scope, data.discordScope].filter((s) => s !== "none"),
        app_ids: data.apps === "all" ? null : data.appIds,
      },
      {
        onSuccess(res) {
          if (res.success) {
            track("api_token_created");
            setCreatedToken(res.data.token);
            form.reset();
          } else if (res.error.code === "validation_failed") {
            setValidationErrors(form, res.error.data, {
              expires_in_days: "expiresIn",
              scopes: "scope",
              app_ids: "appIds",
            });
          } else {
            toast.error(
              `Failed to create token: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  function copy(text: string) {
    navigator.clipboard
      .writeText(text)
      .then(() => toast.success("Copied to clipboard!"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        onOpenChange(open);
        // The token isn't kept around after closing.
        if (!open) setCreatedToken(null);
      }}
    >
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>API Tokens</DialogTitle>
          <DialogDescription>
            Let AI tools like Claude Code or Cursor build your bots, through
            MCP.
          </DialogDescription>
        </DialogHeader>

        {createdToken ? (
          <div className="grid gap-3">
            <Alert variant="destructive">
              <TriangleAlertIcon className="size-4" />
              <AlertTitle>Copy your token now</AlertTitle>
              <AlertDescription>
                It can&apos;t be shown again. Treat it like a password: never
                paste it in Discord or send it to anyone.
              </AlertDescription>
            </Alert>
            <div className="flex gap-2">
              <Input readOnly value={createdToken} className="font-mono" />
              <Button
                variant="outline"
                size="icon"
                className="flex-none"
                onClick={() => copy(createdToken)}
                aria-label="Copy token"
              >
                <CopyIcon className="size-4" />
              </Button>
            </div>
            {/* The MCP server is an API route, which export builds lack. */}
            {!isExport && <MCPSetup token={createdToken} copy={copy} />}
            <Button
              className="justify-self-end"
              onClick={() => setCreatedToken(null)}
            >
              Done
            </Button>
          </div>
        ) : (
          <div className="flex min-w-0 flex-col gap-5">
            <Alert variant="destructive">
              <TriangleAlertIcon className="size-4" />
              <AlertTitle>A token is a password to your bots</AlertTitle>
              <AlertDescription className="grid gap-2">
                <p>
                  Anyone who has it can change your bots and make them do
                  anything, like banning members or deleting messages. Only
                  paste it into tools you run yourself.
                </p>
                <p>
                  Never send it to anyone, not even Kite staff. Nobody needs
                  your token to help you. If someone asks for it, they want to
                  take over your bots.
                </p>
              </AlertDescription>
            </Alert>
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="grid gap-3"
              >
                <div className="flex gap-2 items-start">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem className="min-w-0 flex-1">
                        <FormControl>
                          <Input
                            placeholder="Token name, like Claude Code"
                            maxLength={100}
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <SelectField
                    control={form.control}
                    name="expiresIn"
                    label="Expires"
                    options={expiryOptions}
                    className="flex-none"
                    triggerClassName="w-28"
                  />
                </div>
                <div className="grid gap-2 sm:grid-cols-3">
                  <SelectField
                    control={form.control}
                    name="scope"
                    label="Apps access"
                    showLabel
                    options={scopeOptions}
                  />
                  <SelectField
                    control={form.control}
                    name="discordScope"
                    label="Discord API"
                    showLabel
                    options={discordScopeOptions}
                  />
                  <SelectField
                    control={form.control}
                    name="apps"
                    label="Which apps"
                    showLabel
                    options={appsOptions}
                  />
                </div>
                {someApps && (
                  <FormField
                    control={form.control}
                    name="appIds"
                    rules={{
                      validate: (v) => v.length > 0 || "Pick at least one app",
                    }}
                    render={({ field }) => (
                      <FormItem className="basis-full">
                        <div className="grid max-h-40 gap-2 overflow-y-auto rounded-md border p-3">
                          {!apps ? (
                            <Skeleton className="h-5" />
                          ) : (
                            apps.map(
                              (app) =>
                                app && (
                                  <label
                                    key={app.id}
                                    className="flex items-center gap-2 text-sm"
                                  >
                                    <Checkbox
                                      checked={field.value.includes(app.id)}
                                      onCheckedChange={(checked) =>
                                        field.onChange(
                                          checked
                                            ? [...field.value, app.id]
                                            : field.value.filter(
                                                (id) => id !== app.id
                                              )
                                        )
                                      }
                                    />
                                    <span className="truncate">{app.name}</span>
                                  </label>
                                )
                            )
                          )}
                        </div>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                <FormField
                  control={form.control}
                  name="understood"
                  render={({ field }) => (
                    <FormItem>
                      <label className="flex items-start gap-2 text-sm">
                        <FormControl>
                          <Checkbox
                            checked={field.value}
                            onCheckedChange={(checked) =>
                              field.onChange(checked === true)
                            }
                            className="mt-0.5"
                          />
                        </FormControl>
                        I understand that anyone with this token can control my
                        bots, and I will never share it.
                      </label>
                    </FormItem>
                  )}
                />
                <LoadingButton
                  type="submit"
                  loading={createMutation.isPending}
                  disabled={!understood}
                  className="justify-self-end"
                >
                  Create
                </LoadingButton>
              </form>
            </Form>

            <div className="flex flex-col gap-2 max-h-80 overflow-y-auto">
              {!tokens ? (
                <Skeleton className="h-14" />
              ) : tokens.length === 0 ? (
                <div className="text-sm text-muted-foreground">
                  No tokens yet.
                </div>
              ) : (
                tokens.map(
                  (token) =>
                    token && <APITokenEntry key={token.id} token={token} />
                )
              )}
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

function SelectField({
  control,
  name,
  label,
  options,
  showLabel,
  className,
  triggerClassName,
}: {
  control: Control<FormFields>;
  name: "expiresIn" | "scope" | "discordScope" | "apps";
  label: string;
  // Above the select, otherwise it's only for screen readers.
  showLabel?: boolean;
  options: Option[];
  className?: string;
  triggerClassName?: string;
}) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={className}>
          {showLabel && (
            <FormLabel className="text-xs text-muted-foreground">
              {label}
            </FormLabel>
          )}
          <Select value={field.value} onValueChange={field.onChange}>
            <FormControl>
              <SelectTrigger
                className={cn("w-full", triggerClassName)}
                aria-label={label}
              >
                <SelectValue />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              {options.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

// How to connect MCP clients with a new token, as config to copy.
function MCPSetup({
  token,
  copy,
}: {
  token: string;
  copy: (text: string) => void;
}) {
  const url = `${window.location.origin}/api/mcp`;
  const header = `Authorization: Bearer ${token}`;
  const clients = [
    {
      name: "Claude Code",
      hint: "Run this in a terminal:",
      config: `claude mcp add --transport http kite ${url} --header "${header}"`,
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
      name: "Other",
      hint: "Add an MCP server with Streamable HTTP and this header:",
      config: `${url}\n${header}`,
    },
  ];

  return (
    <Tabs defaultValue={clients[0].name} className="min-w-0">
      <div className="mb-1 text-sm font-medium">Connect an MCP client</div>
      <TabsList>
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

function APITokenEntry({ token }: { token: APIToken }) {
  // Write first, as it includes read.
  const access = scopeOptions.find((o) => token.scopes.includes(o.value));
  const discord = discordScopeOptions.find((o) =>
    token.scopes.includes(o.value)
  );
  const deleteMutation = useAPITokenDeleteMutation(token.id);

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Token deleted!");
        } else {
          toast.error(
            `Failed to delete token: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  const expired = !!token.expires_at && new Date(token.expires_at) < new Date();

  return (
    <div className="flex items-center gap-3 rounded-md border px-3 py-2">
      <KeyRoundIcon className="size-4 flex-none text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium">{token.name}</div>
        <div className="text-xs text-muted-foreground">
          {[
            access?.summary,
            discord?.summary,
            !token.app_ids
              ? "All apps"
              : token.app_ids.length === 1
              ? "1 app"
              : `${token.app_ids.length} apps`,
            token.last_used_at
              ? `Last used ${formatDateTime(new Date(token.last_used_at))}`
              : "Never used",
            !token.expires_at
              ? "Doesn't expire"
              : expired
              ? "Expired"
              : `Expires ${formatDateTime(new Date(token.expires_at))}`,
          ]
            .filter(Boolean)
            .join(" · ")}
        </div>
      </div>
      <ConfirmDialog
        title="Are you sure that you want to delete this token?"
        description="Tools that use it stop working. This can't be undone."
        onConfirm={remove}
      >
        <Button size="sm" variant="ghost">
          Delete
        </Button>
      </ConfirmDialog>
    </div>
  );
}
