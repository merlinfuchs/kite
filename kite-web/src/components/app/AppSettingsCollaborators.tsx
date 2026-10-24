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
  useAppCollaboratorDeleteMutation,
  useAppCollaboratorInviteDeleteMutation,
} from "@/lib/api/mutations";
import {
  useAppCollaboratorInvites,
  useAppCollaborators,
  useAppFeature,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { MinusIcon } from "lucide-react";
import ConfirmDialog from "../common/ConfirmDialog";
import { Button } from "../ui/button";
import AppCollaboratorAddDialog from "./AppCollaboratorAddDialog";
import { toast } from "sonner";
import { APIResponse } from "@/lib/api/response";

export default function AppSettingsCollaborators() {
  const appId = useAppId();
  const collaborators = useAppCollaborators();
  const invites = useAppCollaboratorInvites();

  const maxCollaborators = useAppFeature((f) => f.max_collaborators) || 0;
  const currentCollaborators =
    (collaborators?.length || 0) + (invites?.length || 0);

  const deleteMutation = useAppCollaboratorDeleteMutation(appId);
  const inviteDeleteMutation = useAppCollaboratorInviteDeleteMutation(appId);

  function onRemoved(res: APIResponse<unknown>) {
    if (!res.success) {
      toast.error(
        `Failed to remove collaborator: ${res.error.message} (${res.error.code})`
      );
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex gap-3">
          <CardTitle>Collaborators</CardTitle>
          <div className="text-muted-foreground">
            {currentCollaborators} / {maxCollaborators}
          </div>
        </div>
        <CardDescription>
          Add or remove other users who can manage this app.
        </CardDescription>
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
            {collaborators?.map((collaborator) => (
              <TableRow key={collaborator!.user.id}>
                <TableCell className="font-medium">
                  {collaborator!.user.display_name}
                </TableCell>
                <TableCell>{collaborator!.user.discord_id}</TableCell>
                <TableCell>{collaborator!.role}</TableCell>
                <TableCell className="text-right">
                  {collaborator!.role !== "owner" && (
                    <RemoveCollaboratorButton
                      onConfirm={() =>
                        deleteMutation.mutate(collaborator!.user.id, {
                          onSuccess: onRemoved,
                        })
                      }
                    />
                  )}
                </TableCell>
              </TableRow>
            ))}
            {invites?.map((invite) => (
              <TableRow key={invite!.discord_username}>
                <TableCell className="font-medium">
                  @{invite!.discord_username}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  Hasn&apos;t logged in yet
                </TableCell>
                <TableCell>{invite!.role}</TableCell>
                <TableCell className="text-right">
                  <RemoveCollaboratorButton
                    onConfirm={() =>
                      inviteDeleteMutation.mutate(invite!.discord_username, {
                        onSuccess: onRemoved,
                      })
                    }
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        <AppCollaboratorAddDialog>
          <Button
            variant="outline"
            disabled={currentCollaborators >= maxCollaborators}
          >
            Add Collaborator
          </Button>
        </AppCollaboratorAddDialog>
      </CardContent>
    </Card>
  );
}

function RemoveCollaboratorButton({ onConfirm }: { onConfirm: () => void }) {
  return (
    <ConfirmDialog
      title="Remove Collaborator"
      description="Are you sure you want to remove this collaborator?"
      onConfirm={onConfirm}
    >
      <Button variant="ghost" size="icon">
        <MinusIcon className="h-5 w-5" />
      </Button>
    </ConfirmDialog>
  );
}
