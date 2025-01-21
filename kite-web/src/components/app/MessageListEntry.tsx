import {
  CheckIcon,
  CopyPlusIcon,
  EllipsisIcon,
  MailIcon,
  Trash2Icon,
} from "lucide-react";
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
import { DropdownMenuItem } from "../ui/dropdown-menu";
import {
  DropdownMenu,
  DropdownMenuGroup,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import MessageDuplicateDialog from "./MessageDuplicateDialog";

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
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="icon" variant="ghost">
              <EllipsisIcon className="h-5 w-5 text-muted-foreground" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuGroup>
              <ConfirmDialog
                title="Are you sure that you want to delete this message?"
                description="This will remove the message from your app and cannot be undone."
                onConfirm={remove}
              >
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Trash2Icon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Delete Message
                </DropdownMenuItem>
              </ConfirmDialog>
              <MessageDuplicateDialog message={message}>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <CopyPlusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Duplicate Message
                </DropdownMenuItem>
              </MessageDuplicateDialog>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardFooter>
    </Card>
  );
}
