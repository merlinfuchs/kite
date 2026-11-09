import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { MailIcon } from "lucide-react";
import { useRouter } from "next/router";
import { toast } from "sonner";
import {
  useUserInviteAcceptMutation,
  useUserInviteDeclineMutation,
} from "@/lib/api/mutations";
import { useUserInvites } from "@/lib/hooks/api";
import {
  CollaboratorRole,
  collaboratorRoleLabel,
  collaboratorRoles,
} from "@/lib/collaborators";
import { abbreviateName, cn, formatDateTime } from "@/lib/utils";
import { Avatar, AvatarFallback } from "../ui/avatar";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import LoadingButton from "../common/LoadingButton";
import { track } from "@/lib/analytics";

// Lists the invites of the user to collaborate on other apps. They only become
// a collaborator once they accept.
export default function UserInvitesDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const router = useRouter();
  const invites = useUserInvites();
  const acceptMutation = useUserInviteAcceptMutation();
  const declineMutation = useUserInviteDeclineMutation();

  function accept(appId: string, appName: string) {
    acceptMutation.mutate(appId, {
      onSuccess(res) {
        if (res.success) {
          track("collaborator_invite_accepted", { role: res.data.role });
          toast.success(`You joined ${appName}.`);
          onOpenChange(false);
          router.push(`/apps/${appId}`);
        } else {
          toast.error(
            `Failed to accept invite: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  function decline(appId: string) {
    declineMutation.mutate(appId, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Invite declined.");
        } else {
          toast.error(
            `Failed to decline invite: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Invites</DialogTitle>
          <DialogDescription>
            Apps you were invited to collaborate on. Invites expire after 7
            days.
          </DialogDescription>
        </DialogHeader>
        {invites?.length ? (
          <div className="space-y-3">
            {invites.map((invite) => {
              if (!invite) return null;
              const pending =
                (acceptMutation.isPending &&
                  acceptMutation.variables === invite.app.id) ||
                (declineMutation.isPending &&
                  declineMutation.variables === invite.app.id);

              return (
                <div
                  key={invite.app.id}
                  className="rounded-lg border p-4 space-y-3"
                >
                  <div className="flex items-start gap-3">
                    <Avatar className="rounded-lg h-10 w-10">
                      <AvatarFallback className="rounded-lg">
                        {abbreviateName(invite.app.name)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold truncate">
                          {invite.app.name}
                        </span>
                        <Badge variant="secondary">
                          {collaboratorRoleLabel(invite.role)}
                        </Badge>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {invite.invited_by
                          ? `Invited by @${invite.invited_by.discord_username}`
                          : "Invited"}{" "}
                        {formatDateTime(new Date(invite.created_at))}
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        {
                          collaboratorRoles[invite.role as CollaboratorRole]
                            ?.description
                        }
                      </div>
                    </div>
                  </div>
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={pending}
                      onClick={() => decline(invite.app.id)}
                    >
                      Decline
                    </Button>
                    <LoadingButton
                      size="sm"
                      loading={
                        acceptMutation.isPending &&
                        acceptMutation.variables === invite.app.id
                      }
                      disabled={pending}
                      onClick={() => accept(invite.app.id, invite.app.name)}
                    >
                      Accept
                    </LoadingButton>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center text-center py-8 gap-2 text-muted-foreground">
            <MailIcon className="size-8" />
            <div className="text-sm">No pending invites.</div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

// InvitesDot marks the account menu while the user has pending invites.
export function InvitesDot({ className }: { className?: string }) {
  const invites = useUserInvites();
  if (!invites?.length) return null;

  return (
    <span
      className={cn(
        "absolute -top-1 -right-1 flex h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-background",
        className
      )}
      aria-label={`${invites.length} pending invites`}
    />
  );
}
