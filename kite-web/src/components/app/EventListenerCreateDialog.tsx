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
import { useEventListenerCreateMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { getUniqueId } from "@/lib/utils";
import { useRouter } from "next/router";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

interface FormFields {
  source: string;
  type: string;
  description: string;
}

export default function EventListenerCreateDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useEventListenerCreateMutation(appId);
  const form = useForm<FormFields>({
    defaultValues: {
      source: "discord",
      type: "",
      description: "",
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        source: data.source,
        flow_source: getInitialFlowData(data.type, data.description),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Event listener created!");
            setOpen(false);

            setTimeout(
              () =>
                router.push({
                  pathname: "/apps/[appId]/events/[eventId]",
                  query: { appId, eventId: res.data.id },
                }),
              500
            );
          } else {
            toast.error(
              `Failed to create event listener: ${res.error.message} (${res.error.code})`
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
          <DialogTitle>Create Event Listener</DialogTitle>
          <DialogDescription>
            Create a new event listener with a name and description.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormDescription>
                    What will you use this event listener for?
                  </FormDescription>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="source"
              render={({ field }) => (
                <FormItem className="min-w-48">
                  <FormLabel>Source</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select event source" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="discord">Discord</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem className="min-w-48">
                  <FormLabel>Event</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select event type" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="message_create">
                        Message Create
                      </SelectItem>
                      <SelectItem value="message_delete">
                        Message Delete
                      </SelectItem>
                      <SelectItem value="message_update">
                        Message Update
                      </SelectItem>
                      <SelectItem value="guild_member_add">
                        Server Member Add
                      </SelectItem>
                      <SelectItem value="guild_member_remove">
                        Server Member Remove
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Create event listener
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function getInitialFlowData(type: string, description: string) {
  return {
    nodes: [
      {
        id: getUniqueId().toString(),
        position: { x: 0, y: 0 },
        data: { event_type: type, description },
        type: "entry_event",
      },
    ],
    edges: [],
  };
}
