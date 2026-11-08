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
import { useCommandCreateMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { useRouter } from "next/router";
import { setValidationErrors } from "@/lib/form";
import {
  NewCommandInstall,
  NewCommandType,
  newCommandFlow,
} from "@/lib/flow/newFlows";
import { track } from "@/lib/analytics";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

interface FormFields {
  name: string;
  description: string;
  command_style: "slash" | "contextual";
  context_target: "user" | "message";
  install: NewCommandInstall;
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
      command_style: "slash",
      context_target: "user",
      install: "guild_install",
    },
  });
  const commandStyle = form.watch("command_style");

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    const commandType: NewCommandType =
      data.command_style === "contextual" ? data.context_target : "chat_input";

    createMutation.mutate(
      {
        flow_source: newCommandFlow(
          data.name,
          commandType === "chat_input" ? data.description : "",
          commandType,
          data.install
        ),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            track("command_created", { source: "blank" });
            toast.success("Command created.");
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
          <DialogTitle>Create command</DialogTitle>
          <DialogDescription>
            Create a slash command or a contextual command that appears when
            someone opens the menu on a user or message.
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
                  {commandStyle === "contextual" && (
                    <FormDescription>
                      The name shown in the contextual menu.
                    </FormDescription>
                  )}
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="command_style"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Command type</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="slash">Slash command</SelectItem>
                      <SelectItem value="contextual">
                        Contextual command
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            {commandStyle === "contextual" ? (
              <FormField
                control={form.control}
                name="context_target"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Context menu target</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="user">User</SelectItem>
                        <SelectItem value="message">Message</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            ) : (
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
            <FormField
              control={form.control}
              name="install"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Available for</FormLabel>
                  <FormDescription>
                    Choose where this command can be installed.
                  </FormDescription>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="guild_install">
                        Server install
                      </SelectItem>
                      <SelectItem value="user_install">User install</SelectItem>
                    </SelectContent>
                  </Select>
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
