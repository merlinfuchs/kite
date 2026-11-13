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
import { CommandAvailability, newCommandFlow } from "@/lib/flow/newFlows";
import { CommandType, commandTypeOptions } from "@/lib/flow/commandTypes";
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
  command_type: CommandType;
  availability: CommandAvailability;
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
      command_type: "chat_input",
      availability: "everywhere",
    },
  });
  const isSlash = form.watch("command_type") === "chat_input";

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        flow_source: newCommandFlow(
          data.name,
          data.description,
          data.command_type,
          data.availability
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
            Members use slash commands by typing / in the chat. User and Message
            commands show in the Apps menu when someone right-clicks a user or
            message.
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
                      {commandTypeOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
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
                    <Input
                      type="text"
                      placeholder={isSlash ? "ping" : "Report message"}
                      {...field}
                    />
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
            <FormField
              control={form.control}
              name="availability"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Where it works</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="everywhere">Everywhere</SelectItem>
                      <SelectItem value="servers">Only in servers</SelectItem>
                      <SelectItem value="users">
                        Only for users who added the bot
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Users can add the bot to their account to use its commands
                    anywhere. Change this later with the Command Contexts block.
                  </FormDescription>
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
