import { CheckIcon, SatelliteDishIcon } from "lucide-react";
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
import ConfirmDialog from "../common/ConfirmDialog";
import { useEventListenerDeleteMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { formatDateTime } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { EventListener } from "@/lib/types/wire.gen";

export default function EventListenerListEntry({
  listener,
}: {
  listener: EventListener;
}) {
  const router = useRouter();

  const deleteMutation = useEventListenerDeleteMutation(
    useAppId(),
    listener.id
  );

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Event listener deleted!");
        } else {
          toast.error(
            `Failed to delete event listener: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card>
      <div className="float-right pt-3 pr-4">
        <div className="flex items-center space-x-2">
          <Tooltip>
            <TooltipTrigger>
              <CheckIcon className="h-5 w-5 text-green-500" />
            </TooltipTrigger>
            <TooltipContent>
              <div className="text-foreground/90">
                All changes have been deployed!
              </div>
            </TooltipContent>
          </Tooltip>
          <div className="text-sm text-muted-foreground">
            {formatDateTime(new Date(listener.updated_at))}
          </div>
        </div>
      </div>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <SatelliteDishIcon className="h-5 w-5 text-muted-foreground" />
          <div>{listener.type}</div>
        </CardTitle>
        <CardDescription className="text-sm">
          {listener.description}
        </CardDescription>
      </CardHeader>
      <CardFooter className="flex space-x-3">
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]/events/[eventId]",
              query: {
                appId: router.query.appId,
                eventId: listener.id,
              },
            }}
          >
            Manage
          </Link>
        </Button>
        <ConfirmDialog
          title="Are you sure that you want to delete this event listener?"
          description="This will remove the event listener from your app and cannot be undone."
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
