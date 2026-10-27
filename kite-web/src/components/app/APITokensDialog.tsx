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
import { CopyIcon, KeyRoundIcon } from "lucide-react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "../ui/form";
import { Button } from "../ui/button";
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
  apps: string;
  appIds: string[];
}

interface Option {
  value: string;
  label: string;
}

const scopeOptions: Option[] = [
  { value: "kite:write", label: "Read and change apps" },
  { value: "kite:read", label: "Only read apps" },
];

const appsOptions: Option[] = [
  { value: "all", label: "All apps" },
  { value: "some", label: "Only some apps" },
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
      apps: "all",
      appIds: [],
    },
  });
  const someApps = form.watch("apps") === "some";

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        name: data.name,
        expires_in_days:
          data.expiresIn === "never" ? null : Number(data.expiresIn),
        scopes: [data.scope],
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
            Use Kite from other tools, like an MCP client, without logging in. A
            token that can change apps has almost full access to them. Its flows
            can do anything your bots can and use your app secrets. Never share
            a token with anyone, and delete it if it leaks.
          </DialogDescription>
        </DialogHeader>

        {createdToken ? (
          <div className="grid gap-3">
            <div className="text-sm">
              Copy your token now. It can&apos;t be shown again. Anyone with it
              can control your bots, so keep it private.
            </div>
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
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="flex flex-wrap gap-2 items-start"
              >
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem className="basis-full sm:basis-0 sm:flex-1">
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
                  className="flex-1 sm:flex-none"
                  triggerClassName="sm:w-28"
                />
                <SelectField
                  control={form.control}
                  name="scope"
                  label="Access"
                  options={scopeOptions}
                />
                <SelectField
                  control={form.control}
                  name="apps"
                  label="Apps"
                  options={appsOptions}
                />
                <LoadingButton type="submit" loading={createMutation.isPending}>
                  Create
                </LoadingButton>
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
  className = "flex-1",
  triggerClassName,
}: {
  control: Control<FormFields>;
  name: "expiresIn" | "scope" | "apps";
  label: string;
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
            access?.label,
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
