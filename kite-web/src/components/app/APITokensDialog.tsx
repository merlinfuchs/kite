import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
  CopyIcon,
  KeyRoundIcon,
  PlugIcon,
  TriangleAlertIcon,
} from "lucide-react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "../ui/form";
import { Button } from "../ui/button";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { Skeleton } from "../ui/skeleton";
import LoadingButton from "../common/LoadingButton";
import ConfirmDialog from "../common/ConfirmDialog";
import MCPClientSetup from "./MCPClientSetup";
import APITokenAccessFields, {
  AccessFields,
  Option,
  SelectField,
  accessScopes,
  discordScopeOptions,
  scopeOptions,
} from "./APITokenAccessFields";
import {
  useAPITokenCreateMutation,
  useAPITokenDeleteMutation,
} from "@/lib/api/mutations";
import { useAPITokensQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { setValidationErrors } from "@/lib/form";
import { formatDateTime } from "@/lib/utils";
import { oauthClientDomain } from "@/lib/oauth";
import { isExport } from "@/lib/env/client";
import { track } from "@/lib/analytics";
import { APIToken } from "@/lib/types/wire.gen";

interface FormFields extends AccessFields {
  name: string;
  // Days as a string for the select, "never" for no expiry.
  expiresIn: string;
  understood: boolean;
}

const expiryOptions: Option[] = [
  { value: "30", label: "30 days" },
  { value: "90", label: "90 days" },
  { value: "365", label: "1 year" },
  { value: "never", label: "Never" },
];

// Lists the user's API tokens and creates new ones. A new token is shown once,
// it can't be read back later. The tokens of OAuth clients like Claude are
// listed too, as connections.
export default function APITokensDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const tokens = useResponseData(useAPITokensQuery(open));
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
  const understood = form.watch("understood");

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        name: data.name,
        expires_in_days:
          data.expiresIn === "never" ? null : Number(data.expiresIn),
        scopes: accessScopes(data),
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
            Let AI tools build your bots, through MCP. Claude, ChatGPT, Claude
            Code and Codex connect by signing in and show up here. Create a
            token for clients that can&apos;t, like Cursor.
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
            {!isExport && (
              <div className="min-w-0">
                <div className="mb-1 text-sm font-medium">
                  Connect an MCP client
                </div>
                <MCPClientSetup
                  url={`${window.location.origin}/api/mcp`}
                  token={createdToken}
                  copy={copy}
                />
              </div>
            )}
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
                <APITokenAccessFields />
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
                  No tokens or connected clients yet.
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

function APITokenEntry({ token }: { token: APIToken }) {
  // Write first, as it includes read.
  const access = scopeOptions.find((o) => token.scopes.includes(o.value));
  const discord = discordScopeOptions.find((o) =>
    token.scopes.includes(o.value)
  );
  const deleteMutation = useAPITokenDeleteMutation(token.id);
  const clientDomain =
    token.oauth_client_id && oauthClientDomain(token.oauth_client_id);

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success(clientDomain ? "Disconnected!" : "Token deleted!");
        } else {
          toast.error(
            `Failed to delete token: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  const expired = !!token.expires_at && new Date(token.expires_at) < new Date();
  const Icon = clientDomain ? PlugIcon : KeyRoundIcon;

  return (
    <div className="flex items-center gap-3 rounded-md border px-3 py-2">
      <Icon className="size-4 flex-none text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium">{token.name}</div>
        <div className="text-xs text-muted-foreground">
          {[
            // Clients connected with OAuth stay connected while they're
            // used, so they have no expiry to show.
            clientDomain && `Connected from ${clientDomain}`,
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
            clientDomain
              ? null
              : !token.expires_at
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
        title={
          clientDomain
            ? `Are you sure that you want to disconnect ${token.name}?`
            : "Are you sure that you want to delete this token?"
        }
        description={
          clientDomain
            ? "It can't use your apps anymore until you connect it again."
            : "Tools that use it stop working. This can't be undone."
        }
        onConfirm={remove}
      >
        <Button size="sm" variant="ghost">
          {clientDomain ? "Disconnect" : "Delete"}
        </Button>
      </ConfirmDialog>
    </div>
  );
}
