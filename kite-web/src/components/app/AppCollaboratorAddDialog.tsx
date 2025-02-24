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

interface FormFields {
  discord_user_id: string;
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
      discord_user_id: "",
      role: "admin",
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(data, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Collaborator added!");
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
    });
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
              name="discord_user_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Discord User ID</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
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
