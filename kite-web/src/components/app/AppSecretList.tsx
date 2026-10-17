import { KeyRoundIcon } from "lucide-react";
import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useAppSecrets } from "@/lib/hooks/api";
import AppSecretDialog from "./AppSecretDialog";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import ConfirmDialog from "../common/ConfirmDialog";
import { useAppSecretDeleteMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { AppSecret } from "@/lib/types/wire.gen";
import { formatDateTime } from "@/lib/utils";

export default function AppSecretList() {
  const secrets = useAppSecrets();

  const createButton = (
    <AppSecretDialog>
      <Button>Create secret</Button>
    </AppSecretDialog>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!secrets ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : secrets.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no secrets"
          description="Store API keys here instead of in your blocks."
          action={createButton}
        />
      ) : (
        <>
          {secrets.map((secret) => (
            <AppSecretListEntry secret={secret!} key={secret!.id} />
          ))}
          <div className="flex">{createButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}

function AppSecretListEntry({ secret }: { secret: AppSecret }) {
  const deleteMutation = useAppSecretDeleteMutation(useAppId(), secret.id);

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Secret deleted!");
        } else {
          toast.error(
            `Failed to delete secret: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <KeyRoundIcon className="h-5 w-5 text-muted-foreground" />
          <div>{secret.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">
          <code>{`{{secrets.${secret.name}}}`}</code>, updated{" "}
          {formatDateTime(new Date(secret.updated_at))}
        </CardDescription>
      </CardHeader>
      <CardFooter className="flex space-x-3">
        <AppSecretDialog secret={secret}>
          <Button size="sm" variant="outline">
            Edit
          </Button>
        </AppSecretDialog>
        <ConfirmDialog
          title="Are you sure that you want to delete this secret?"
          description="Blocks that use it will fail. This can't be undone."
          onConfirm={remove}
        >
          <Button size="sm" variant="ghost">
            Delete
          </Button>
        </ConfirmDialog>
      </CardFooter>
    </Card>
  );
}
