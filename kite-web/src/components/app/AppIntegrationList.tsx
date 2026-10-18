import { ReactNode, useState } from "react";
import { PlugIcon } from "lucide-react";
import { toast } from "sonner";
import { integrations, needsCredential } from "@/lib/integrations";
import { CredentialIntegration, Integration } from "@/lib/integrations/types";
import { useAppIntegrations } from "@/lib/hooks/api";
import {
  useAppIntegrationConnectMutation,
  useAppIntegrationRemoveMutation,
  useAppIntegrationUpdateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { formatDateTime } from "@/lib/utils";
import { AppIntegration } from "@/lib/types/wire.gen";
import { Button } from "../ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Skeleton } from "../ui/skeleton";
import { Switch } from "../ui/switch";
import ConfirmDialog from "../common/ConfirmDialog";
import LoadingButton from "../common/LoadingButton";

export default function AppIntegrationList() {
  const states = useAppIntegrations();

  if (!states) {
    return (
      <div className="flex flex-col space-y-5">
        <Skeleton className="h-28" />
        <Skeleton className="h-28" />
      </div>
    );
  }

  return (
    <div className="flex flex-col space-y-5">
      {integrations.map((integration) => (
        <AppIntegrationEntry
          key={integration.id}
          integration={integration}
          state={states.find((s) => s?.integration_id === integration.id)}
        />
      ))}
    </div>
  );
}

function AppIntegrationEntry({
  integration,
  state,
}: {
  integration: Integration;
  state?: AppIntegration;
}) {
  const appId = useAppId();
  const updateMutation = useAppIntegrationUpdateMutation(appId, integration.id);
  const removeMutation = useAppIntegrationRemoveMutation(appId, integration.id);

  function setEnabled(enabled: boolean) {
    updateMutation.mutate(
      { enabled },
      {
        onSuccess(res) {
          if (!res.success) {
            toast.error(
              `Failed to update ${integration.name}: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  function remove() {
    removeMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success(`${integration.name} removed`);
        } else {
          toast.error(
            `Failed to remove ${integration.name}: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  // Integrations that need a key are enabled by entering it, and keep it
  // while they're disabled.
  const connectedAt = state?.credential_updated_at;
  const connectable = needsCredential(integration);
  const toggleable =
    integration.availability !== "always" && (!connectable || !!connectedAt);
  const status =
    integration.availability === "always"
      ? "Always enabled"
      : `${state?.enabled ? "Enabled" : "Disabled"}${
          connectedAt
            ? `, key updated ${formatDateTime(new Date(connectedAt))}`
            : ""
        }`;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center justify-between space-x-2">
          <div className="flex items-center space-x-2">
            <PlugIcon className="h-5 w-5 text-muted-foreground" />
            <div>{integration.name}</div>
          </div>
          {toggleable && (
            <Switch
              checked={
                updateMutation.isPending
                  ? !!updateMutation.variables?.enabled
                  : !!state?.enabled
              }
              onCheckedChange={setEnabled}
              disabled={updateMutation.isPending}
            />
          )}
        </CardTitle>
        <CardDescription className="text-sm">
          {integration.description} {status}.
        </CardDescription>
      </CardHeader>
      {needsCredential(integration) && (
        <CardFooter className="flex space-x-3">
          <AppIntegrationConnectDialog
            integration={integration}
            replace={!!connectedAt}
          >
            <Button size="sm" variant={connectedAt ? "outline" : "default"}>
              {connectedAt ? "Replace key" : "Enable"}
            </Button>
          </AppIntegrationConnectDialog>
          {connectedAt && (
            <ConfirmDialog
              title={`Remove ${integration.name}?`}
              description="Its key is deleted, and its blocks fail until you enable it again."
              onConfirm={remove}
            >
              <Button size="sm" variant="ghost">
                Remove
              </Button>
            </ConfirmDialog>
          )}
        </CardFooter>
      )}
    </Card>
  );
}

// Enables an integration by entering its key, or replaces the key.
function AppIntegrationConnectDialog({
  integration,
  replace,
  children,
}: {
  integration: CredentialIntegration;
  replace: boolean;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [credential, setCredential] = useState("");
  const connectMutation = useAppIntegrationConnectMutation(
    useAppId(),
    integration.id
  );
  const { auth } = integration;

  function connect() {
    if (!credential || connectMutation.isPending) return;
    connectMutation.mutate(
      { credential },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success(
              replace
                ? `${integration.name} key replaced`
                : `${integration.name} enabled`
            );
            setOpen(false);
            setCredential("");
          } else {
            toast.error(res.error.message);
          }
        },
      }
    );
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        if (!open) setCredential("");
      }}
    >
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {replace
              ? `Replace ${integration.name} ${auth.label}`
              : `Enable ${integration.name}`}
          </DialogTitle>
          <DialogDescription>
            Enter your {integration.name} {auth.label}. Its blocks send it with
            their requests, and it can&apos;t be read back.
            {auth.help_url && (
              <>
                {" "}
                <a
                  href={auth.help_url}
                  target="_blank"
                  className="text-primary hover:underline"
                >
                  Where to find it
                </a>
              </>
            )}
          </DialogDescription>
        </DialogHeader>
        {!replace && integration.data_shared && (
          <div className="text-sm text-muted-foreground bg-muted rounded p-3">
            {integration.data_shared}
            {integration.privacy_url && (
              <>
                {" "}
                <a
                  href={integration.privacy_url}
                  target="_blank"
                  className="text-primary hover:underline"
                >
                  Privacy policy
                </a>
              </>
            )}
            {integration.terms_url && (
              <>
                {" · "}
                <a
                  href={integration.terms_url}
                  target="_blank"
                  className="text-primary hover:underline"
                >
                  Terms
                </a>
              </>
            )}
          </div>
        )}
        <Input
          type="password"
          autoComplete="off"
          placeholder={auth.label}
          value={credential}
          onChange={(e) => setCredential(e.target.value)}
        />
        <DialogFooter>
          <LoadingButton onClick={connect} loading={connectMutation.isPending}>
            {replace ? "Replace" : "Enable"}
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
