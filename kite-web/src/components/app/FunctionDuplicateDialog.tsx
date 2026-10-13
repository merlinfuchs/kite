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
import { toast } from "sonner";
import { useRouter } from "next/router";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { useFunctionCreateMutation } from "@/lib/api/mutations";
import { FlowFunction } from "@/lib/types/wire.gen";
import { FlowData } from "@/lib/types/flow.gen";

interface FormFields {
  name: string;
}

export default function FunctionDuplicateDialog({
  children,
  function: fn,
}: {
  children: ReactNode;
  function: FlowFunction;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useFunctionCreateMutation(appId);
  const form = useForm<FormFields>({
    defaultValues: {
      name: `${fn.name} (copy)`.slice(0, 100),
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(
      {
        flow_source: renameFunction(fn.flow_source, data.name),
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Function duplicated!");
            setOpen(false);

            setTimeout(
              () =>
                router.push({
                  pathname: "/apps/[appId]/functions/[functionId]",
                  query: { appId, functionId: res.data.id },
                }),
              500
            );
          } else {
            toast.error(
              `Failed to duplicate function: ${res.error.message} (${res.error.code})`
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
          <DialogTitle>Duplicate Function</DialogTitle>
          <DialogDescription>
            Duplicate an existing function with a new name. Flows that call the
            original keep calling the original.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <FormField
              control={form.control}
              name="name"
              rules={{
                required: "A name is required",
                maxLength: { value: 100, message: "Up to 100 characters" },
              }}
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
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Duplicate function
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function renameFunction(flowData: FlowData, name: string): FlowData {
  return {
    nodes: flowData.nodes.map((node) =>
      node.type === "entry_function"
        ? { ...node, data: { ...node.data, name } }
        : { ...node }
    ),
    edges: [...flowData.edges],
  };
}
