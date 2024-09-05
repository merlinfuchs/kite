import { CheckIcon, MailIcon } from "lucide-react";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Button } from "../ui/button";
import { useRouter } from "next/router";
import Link from "next/link";
import { Message } from "@/lib/types/wire.gen";
import ConfirmDialog from "../common/ConfirmDialog";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { useMessageDeleteMutation } from "@/lib/api/mutations";
import { formatDateTime } from "@/lib/utils";

export default function MessageListEntry({ message }: { message: Message }) {
  const router = useRouter();

  const deleteMutation = useMessageDeleteMutation(useAppId(), message.id);

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Message template deleted!");
        } else {
          toast.error(
            `Failed to delete message template: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card>
      <div className="float-right pt-3 pr-4">
        <div className="flex items-center space-x-2">
          <CheckIcon className="h-5 w-5 text-green-500" />
          <div className="text-sm text-muted-foreground">
            {formatDateTime(new Date(message.updated_at))}
          </div>
        </div>
      </div>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <MailIcon className="h-5 w-5 text-muted-foreground" />
          <div>{message.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">
          {message.description}
        </CardDescription>
      </CardHeader>
      <CardFooter className="flex space-x-3">
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]/messages/[messageId]",
              query: {
                appId: router.query.appId,
                messageId: message.id,
              },
            }}
          >
            Manage
          </Link>
        </Button>
        <ConfirmDialog
          title="Are you sure that you want to delete this message?"
          description="This will remove the message from your app and cannot be undone."
          onConfirm={remove}
        >
          <Button
            size="sm"
            variant="ghost"
            className="space-x-2 flex items-center"
          >
            <div>Delete</div>
          </Button>
        </ConfirmDialog>
      </CardFooter>
    </Card>
  );
}
