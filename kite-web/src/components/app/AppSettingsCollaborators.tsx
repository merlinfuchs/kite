import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useAppCollaboratorDeleteMutation,
  useAppCollaboratorInviteDeleteMutation,
  useAppCollaboratorUpdateMutation,
} from "@/lib/api/mutations";
import {
  useAppCollaboratorInvites,
  useAppCollaborators,
  useAppFeature,
  useAppRole,
  useUser,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { HistoryIcon, MinusIcon } from "lucide-react";
import Link from "next/link";
import ConfirmDialog from "../common/ConfirmDialog";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import AppCollaboratorAddDialog from "./AppCollaboratorAddDialog";
import { toast } from "sonner";
import { APIResponse } from "@/lib/api/response";
import {
  canManageSettings,
  canRemoveRole,
  collaboratorRoleLabel,
  collaboratorRoles,
  invitableRoles,
} from "@/lib/collaborators";
import { formatDateTime } from "@/lib/utils";

export default function AppSettingsCollaborators() {
  const appId = useAppId();
  const user = useUser();
  const role = useAppRole();
  const collaborators = useAppCollaborators();
  const invites = useAppCollaboratorInvites();

  const maxCollaborators = useAppFeature((f) => f.max_collaborators) || 0;
  const currentCollaborators =
    (collaborators?.length || 0) + (invites?.length || 0);

  const deleteMutation = useAppCollaboratorDeleteMutation(appId);
  const updateMutation = useAppCollaboratorUpdateMutation(appId);
  const inviteDeleteMutation = useAppCollaboratorInviteDeleteMutation(appId);

  const canInvite = invitableRoles(role).length > 0;

  function onRemoved(res: APIResponse<unknown>) {
    if (!res.success) {
      toast.error(
        `Failed to remove collaborator: ${res.error.message} (${res.error.code})`
      );
    }
  }

  function changeRole(userId: string, newRole: string) {
    updateMutation.mutate(
      { userId, req: { role: newRole } },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Role changed.");
          } else {
            toast.error(
              `Failed to change role: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1.5">
            <div className="flex gap-3">
              <CardTitle>Collaborators</CardTitle>
              <div className="text-muted-foreground">
                {currentCollaborators} / {maxCollaborators}
              </div>
            </div>
            <CardDescription>
              Invite other users to help build this app. Admins can change
              everything, developers everything but the app&apos;s settings.
            </CardDescription>
          </div>
          {canManageSettings(role) && (
            <Button variant="outline" size="sm" asChild>
              <Link href={`/apps/${appId}/audit-log`}>
                <HistoryIcon className="size-4 mr-2" />
                View audit log
              </Link>
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <Table className="mb-5">
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Discord ID</TableHead>
              <TableHead>Role</TableHead>
              <TableHead className="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {collaborators?.map((collaborator) => {
              if (!collaborator) return null;
              const isSelf = collaborator.user.id === user?.id;
              const canChangeRole =
                role === "owner" && collaborator.role !== "owner";

              return (
                <TableRow key={collaborator.user.id}>
                  <TableCell className="font-medium">
                    {collaborator.user.display_name}
                    {isSelf && (
                      <span className="text-muted-foreground font-normal">
                        {" "}
                        (you)
                      </span>
                    )}
                  </TableCell>
                  <TableCell>{collaborator.user.discord_id}</TableCell>
                  <TableCell>
                    {canChangeRole ? (
                      <Select
                        value={collaborator.role}
                        onValueChange={(v) =>
                          changeRole(collaborator.user.id, v)
                        }
                        disabled={updateMutation.isPending}
                      >
                        <SelectTrigger className="h-8 w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {invitableRoles("owner").map((r) => (
                            <SelectItem key={r} value={r}>
                              {collaboratorRoles[r].label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      collaboratorRoleLabel(collaborator.role)
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {!isSelf && canRemoveRole(role, collaborator.role) && (
                      <RemoveCollaboratorButton
                        onConfirm={() =>
                          deleteMutation.mutate(collaborator.user.id, {
                            onSuccess: onRemoved,
                          })
                        }
                      />
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
            {invites?.map((invite) => {
              if (!invite) return null;
              return (
                <TableRow key={invite.discord_username}>
                  <TableCell className="font-medium">
                    @{invite.discord_username}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    <Badge variant="outline" className="font-normal">
                      Invite pending
                    </Badge>{" "}
                    <span className="text-xs">
                      expires {formatDateTime(new Date(invite.expires_at))}
                    </span>
                  </TableCell>
                  <TableCell>{collaboratorRoleLabel(invite.role)}</TableCell>
                  <TableCell className="text-right">
                    {canRemoveRole(role, invite.role) && (
                      <RemoveCollaboratorButton
                        title="Revoke invite"
                        description="Are you sure you want to revoke this invite?"
                        onConfirm={() =>
                          inviteDeleteMutation.mutate(invite.discord_username, {
                            onSuccess: onRemoved,
                          })
                        }
                      />
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>

        {canInvite && (
          <AppCollaboratorAddDialog>
            <Button
              variant="outline"
              disabled={currentCollaborators >= maxCollaborators}
            >
              Invite collaborator
            </Button>
          </AppCollaboratorAddDialog>
        )}
      </CardContent>
    </Card>
  );
}

function RemoveCollaboratorButton({
  onConfirm,
  title = "Remove collaborator",
  description = "Are you sure you want to remove this collaborator?",
}: {
  onConfirm: () => void;
  title?: string;
  description?: string;
}) {
  return (
    <ConfirmDialog
      title={title}
      description={description}
      onConfirm={onConfirm}
    >
      <Button variant="ghost" size="icon">
        <MinusIcon className="h-5 w-5" />
      </Button>
    </ConfirmDialog>
  );
}
