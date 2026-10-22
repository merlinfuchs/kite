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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { getNodeId } from "@/lib/flow/nodes";
import { track } from "@/lib/analytics";
import { commandTypeOptions, isSlashCommand } from "@/lib/flow/commandType";
import {
  CommandType,
  CommandTypeChatInput,
  CommandTypeUser,
} from "@/lib/types/flow.gen";

interface FormFields {
  name: string;
  description: string;
  command_type: CommandType;
}

export default function CommandCreateDialog({
  children,
  contextMenu = false,
}: {
  children: ReactNode;
  contextMenu?: boolean;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useCommandCreateMutation(appId);
  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      description: "",
      command_type: contextMenu ? CommandTypeUser : CommandTypeChatInput,
    },
  });

  const isSlash = isSlashCommand(form.watch("command_type"));

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        flow_source: getInitialFlowData(
          data.name,
          data.description,
          data.command_type
        ),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            track("command_created", { source: "blank" });
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
          <DialogTitle>
            {contextMenu ? "Create Context Menu" : "Create Command"}
          </DialogTitle>
          <DialogDescription>
            {contextMenu
              ? "Create a user or message context menu command."
              : "Create a new command with a name and description."}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            {contextMenu && (
              <FormField
                control={form.control}
                name="command_type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Type</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {commandTypeOptions
                          .filter((o) => !isSlashCommand(o.value))
                          .map((o) => (
                            <SelectItem key={o.value} value={o.value}>
                              {o.label}
                            </SelectItem>
                          ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
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
                Create command
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function getInitialFlowData(
  name: string,
  description: string,
  commandType: CommandType
) {
  return {
    nodes: [
      {
        id: getNodeId("entry_command"),
        position: { x: 0, y: 0 },
        data: {
          name,
          ...(isSlashCommand(commandType)
            ? { description }
            : { command_type: commandType }),
        },
        type: "entry_command",
      },
    ],
    edges: [],
  };
}
