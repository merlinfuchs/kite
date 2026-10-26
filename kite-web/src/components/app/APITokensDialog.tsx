import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useState } from "react";
import { useForm } from "react-hook-form";
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
import LoadingButton from "../common/LoadingButton";
import ConfirmDialog from "../common/ConfirmDialog";
import {
  useAPITokenCreateMutation,
  useAPITokenDeleteMutation,
} from "@/lib/api/mutations";
import { useAPITokensQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { setValidationErrors } from "@/lib/form";
import { formatDateTime } from "@/lib/utils";
import { track } from "@/lib/analytics";
import { APIToken } from "@/lib/types/wire.gen";

interface FormFields {
  name: string;
  // Days as a string for the select, "never" for no expiry.
  expiresIn: string;
}

const expiryOptions = [
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
  const createMutation = useAPITokenCreateMutation();
  const [createdToken, setCreatedToken] = useState<string | null>(null);

  const form = useForm<FormFields>({
    defaultValues: { name: "", expiresIn: "90" },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        name: data.name,
        expires_in_days:
          data.expiresIn === "never" ? null : Number(data.expiresIn),
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
            token has almost full access to all of your apps. Its flows can do
            anything your bots can and use your app secrets. Never share a token
            with anyone, and delete it if it leaks.
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
                <FormField
                  control={form.control}
                  name="expiresIn"
                  render={({ field }) => (
                    <FormItem className="flex-1 sm:flex-none">
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                      >
                        <FormControl>
                          <SelectTrigger
                            className="w-full sm:w-28"
                            aria-label="Expires"
                          >
                            <SelectValue />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {expiryOptions.map((option) => (
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
                <LoadingButton type="submit" loading={createMutation.isPending}>
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

function APITokenEntry({ token }: { token: APIToken }) {
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
          {token.last_used_at
            ? `Last used ${formatDateTime(new Date(token.last_used_at))}`
            : "Never used"}
          {" · "}
          {!token.expires_at
            ? "Doesn't expire"
            : expired
            ? "Expired"
            : `Expires ${formatDateTime(new Date(token.expires_at))}`}
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
