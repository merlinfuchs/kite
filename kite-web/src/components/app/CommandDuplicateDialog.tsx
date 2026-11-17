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
import { useRouter } from "next/router";
import { Command } from "@/lib/types/wire.gen";
import { FlowData } from "@/lib/types/flow.gen";
import { track } from "@/lib/analytics";
import { commandTypeOf } from "@/lib/flow/commandTypes";

interface FormFields {
  name: string;
  description: string;
}

export default function CommandDuplicateDialog({
  children,
  command,
}: {
  children: ReactNode;
  command: Command;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useCommandCreateMutation(appId);
  const isSlash = commandTypeOf(command.flow_source) === "chat_input";
  const form = useForm<FormFields>({
    defaultValues: {
      name: command.name,
      description: command.description,
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        flow_source: updateFlowData(
          command.flow_source,
          data.name,
          isSlash ? data.description : undefined
        ),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            track("command_created", { source: "duplicate" });
            toast.success("Command duplicated.");
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
            toast.error(
              `Failed to duplicate command: ${res.error.message} (${res.error.code})`
            );
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
          <DialogTitle>Duplicate command</DialogTitle>
          <DialogDescription>
            {isSlash
              ? "Duplicate this command with a new name and description."
              : "Duplicate this command with a new name."}
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
            {isSlash && (
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
            )}
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Duplicate command
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function updateFlowData(
  flowData: FlowData,
  name: string,
  description: string | undefined
) {
  return {
    ...flowData,
    nodes: flowData.nodes.map((node) => {
      if (node.type === "entry_command") {
        return {
          ...node,
          data:
            description === undefined
              ? { ...node.data, name }
              : { ...node.data, name, description },
        };
      }
      return { ...node };
    }),
    edges: [...flowData.edges],
  };
}
