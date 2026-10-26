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
import { EventListener } from "@/lib/types/wire.gen";
import { capitalize, eventSourceName } from "@/lib/flow/eventTypes";
import { FlowData } from "@/lib/types/flow.gen";
import { track } from "@/lib/analytics";

interface FormFields {
  source: string;
  type: string;
  description: string;
}

export default function EventListenerDuplicateDialog({
  children,
  listener,
}: {
  children: ReactNode;
  listener: EventListener;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const name = eventSourceName(listener.source);
  const createMutation = useEventListenerCreateMutation(appId);
  const form = useForm<FormFields>({
    defaultValues: {
      description: listener.description,
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        source: listener.source,
        flow_source: updateFlowData(listener.flow_source, data.description),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            track("event_listener_created", {
              type: listener.type,
              source: "duplicate",
            });
            toast.success(`${capitalize(name.singular)} duplicated!`);
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
              `Failed to duplicate ${name.singular}: ${res.error.message} (${res.error.code})`
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
          <DialogTitle>Duplicate {name.title}</DialogTitle>
          <DialogDescription>
            Duplicate an existing {name.singular} with a new description.
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
                    What will you use this {name.singular} for?
                  </FormDescription>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Duplicate {name.singular}
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function updateFlowData(flowData: FlowData, description: string) {
  return {
    ...flowData,
    nodes: flowData.nodes.map((node) => {
      if (node.type === "entry_event") {
        return {
          ...node,
          data: { ...node.data, description },
        };
      }
      return { ...node };
    }),
    edges: [...flowData.edges],
  };
}
