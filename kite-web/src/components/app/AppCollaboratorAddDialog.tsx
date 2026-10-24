import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ReactNode, useState } from "react";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { useForm } from "react-hook-form";
import {
  useAppCollaboratorCreateMutation,
  useAppCreateMutation,
} from "@/lib/api/mutations";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { CircleHelpIcon } from "lucide-react";
import env from "@/lib/env/client";
import { useAppId } from "@/lib/hooks/params";
import { track } from "@/lib/analytics";

interface FormFields {
  discord_user: string;
  role: string;
}

export default function AppCollaboratorAddDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  const appId = useAppId();
  const createMutation = useAppCollaboratorCreateMutation(appId);

  const form = useForm<FormFields>({
    defaultValues: {
      discord_user: "",
      role: "admin",
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      { ...data, discord_user: data.discord_user.trim() },
      {
        onSuccess(res) {
          if (res.success) {
            track("collaborator_added", {
              role: data.role,
              pending: !!res.data.invite,
            });
            if (res.data.invite) {
              toast.success(
                `@${res.data.invite.discord_username} will be added when they log in to Kite.`
              );
            } else {
              toast.success("Collaborator added!");
            }
            form.reset();
            setOpen(false);
          } else {
            if (res.error.code === "validation_failed") {
              setValidationErrors(form, res.error.data);
            } else {
              toast.error(
                `Failed to add collaborator: ${res.error.message} (${res.error.code})`
              );
            }
          }
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Collaborator</DialogTitle>
          <DialogDescription>
            Add a new collaborator to your app to allow them to manage it.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <FormField
              control={form.control}
              name="discord_user"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Discord username or user ID</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="username" {...field} />
                  </FormControl>
                  <FormDescription>
                    If they haven&apos;t logged in to Kite yet, they&apos;re
                    added when they first log in.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Add collaborator
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
