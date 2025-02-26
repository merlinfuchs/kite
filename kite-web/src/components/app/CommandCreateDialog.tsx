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
import { useCommandCreateMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { getUniqueId } from "@/lib/utils";
import { useRouter } from "next/router";
import { setValidationErrors } from "@/lib/form";

interface FormFields {
  name: string;
  description: string;
}

export default function CommandCreateDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useCommandCreateMutation(appId);
  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      description: "",
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        flow_source: getInitialFlowData(data.name, data.description),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Command created!");
            setOpen(false);

            setTimeout(
              () =>
                router.push({
                  pathname: "/apps/[appId]/commands/[cmdId]",
                  query: { appId, cmdId: res.data.id },
                }),
              500
            );
          } else {
            if (res.error.code === "validation_failed") {
              setValidationErrors(form, res.error.data, {
                "flow_source.nodes.0.name": "name",
                "flow_source.nodes.0.description": "description",
              });
            } else {
              toast.error(
                `Failed to create command: ${res.error.message} (${res.error.code})`
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
          <DialogTitle>Create Command</DialogTitle>
          <DialogDescription>
            Create a new command with a name and description.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Create command
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function getInitialFlowData(name: string, description: string) {
  return {
    nodes: [
      {
        id: getUniqueId().toString(),
        position: { x: 0, y: 0 },
        data: { name, description },
        type: "entry_command",
      },
    ],
    edges: [],
  };
}
