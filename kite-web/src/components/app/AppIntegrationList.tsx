import { ReactNode, useState } from "react";
import { PlugIcon } from "lucide-react";
import { toast } from "sonner";
import { integrations, needsCredential } from "@/lib/integrations";
import { CredentialIntegration, Integration } from "@/lib/integrations/types";
import { useAppIntegrations } from "@/lib/hooks/api";
import {
  useAppIntegrationConnectMutation,
  useAppIntegrationDisconnectMutation,
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
import ConfirmDialog from "../common/ConfirmDialog";
import LoadingButton from "../common/LoadingButton";

export default function AppIntegrationList() {
  const connected = useAppIntegrations();

  if (!connected) {
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
          connection={connected.find(
            (c) => c?.integration_id === integration.id
          )}
        />
      ))}
    </div>
  );
}

function AppIntegrationEntry({
  integration,
  connection,
}: {
  integration: Integration;
  connection?: AppIntegration;
}) {
  const appId = useAppId();
  const disconnectMutation = useAppIntegrationDisconnectMutation(
    appId,
    integration.id
  );

  function disconnect() {
    disconnectMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success(`${integration.name} disconnected`);
        } else {
          toast.error(
            `Failed to disconnect: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  const status = !needsCredential(integration)
    ? "Always connected"
    : connection
    ? `Connected, key updated ${formatDateTime(
        new Date(connection.updated_at)
      )}`
    : "Not connected";

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <PlugIcon className="h-5 w-5 text-muted-foreground" />
          <div>{integration.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">
          {integration.description} {status}.
        </CardDescription>
      </CardHeader>
      {needsCredential(integration) && (
        <CardFooter className="flex space-x-3">
          <AppIntegrationConnectDialog integration={integration}>
            <Button size="sm" variant={connection ? "outline" : "default"}>
              {connection ? "Replace key" : "Connect"}
            </Button>
          </AppIntegrationConnectDialog>
          {connection && (
            <ConfirmDialog
              title={`Disconnect ${integration.name}?`}
              description="Its blocks will fail until you connect it again."
              onConfirm={disconnect}
            >
              <Button size="sm" variant="ghost">
                Disconnect
              </Button>
            </ConfirmDialog>
          )}
        </CardFooter>
      )}
    </Card>
  );
}

function AppIntegrationConnectDialog({
  integration,
  children,
}: {
  integration: CredentialIntegration;
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
            toast.success(`${integration.name} connected`);
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
          <DialogTitle>Connect {integration.name}</DialogTitle>
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
        <Input
          type="password"
          autoComplete="off"
          placeholder={auth.label}
          value={credential}
          onChange={(e) => setCredential(e.target.value)}
        />
        <DialogFooter>
          <LoadingButton onClick={connect} loading={connectMutation.isPending}>
            Connect
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
