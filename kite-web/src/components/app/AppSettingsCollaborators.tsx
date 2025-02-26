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
import { useAppCollaboratorDeleteMutation } from "@/lib/api/mutations";
import { useAppCollaborators, useAppFeature } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { MinusIcon } from "lucide-react";
import ConfirmDialog from "../common/ConfirmDialog";
import { Button } from "../ui/button";
import AppCollaboratorAddDialog from "./AppCollaboratorAddDialog";
import { toast } from "sonner";

export default function AppSettingsCollaborators() {
  const appId = useAppId();
  const collaborators = useAppCollaborators();

  const maxCollaborators = useAppFeature((f) => f.max_collaborators) || 0;
  const currentCollaborators = collaborators?.length || 0;

  const deleteMutation = useAppCollaboratorDeleteMutation(appId);

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
                    <ConfirmDialog
                      title="Remove Collaborator"
                      description="Are you sure you want to remove this collaborator?"
                      onConfirm={() => {
                        deleteMutation.mutate(collaborator!.user.id, {
                          onSuccess: (res) => {
                            if (!res.success) {
                              toast.error(
                                `Failed to remove collaborator: ${res.error.message} (${res.error.code})`
                              );
                            }
                          },
                        });
                      }}
                    >
                      <Button variant="ghost" size="icon">
                        <MinusIcon className="h-5 w-5" />
                      </Button>
                    </ConfirmDialog>
                  )}
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
