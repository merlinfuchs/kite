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
import { useRouter } from "next/router";
import { setValidationErrors } from "@/lib/form";
import { newCommandFlow } from "@/lib/flow/newFlows";
import { track } from "@/lib/analytics";
import { commandTypeOptions, isSlashCommand } from "@/lib/flow/commandType";
import { CommandType, CommandTypeChatInput } from "@/lib/types/flow.gen";

interface FormFields {
  name: string;
  description: string;
  command_type: CommandType;
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
      command_type: CommandTypeChatInput,
    },
  });

  const isSlash = isSlashCommand(form.watch("command_type"));

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        flow_source: newCommandFlow(
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
          <DialogTitle>Create Command</DialogTitle>
          <DialogDescription>
            Create a slash command or a user or message context menu command.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
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
                      {commandTypeOptions.map((o) => (
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
