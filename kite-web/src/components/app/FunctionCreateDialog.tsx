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
import { toast } from "sonner";
import { useRouter } from "next/router";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { useFunctionCreateMutation } from "@/lib/api/mutations";
import { getNodeId } from "@/lib/flow/nodes";

interface FormFields {
  name: string;
  description: string;
}

export default function FunctionCreateDialog({
  children,
  onFunctionCreated,
}: {
  children: ReactNode;
  // Called instead of opening the new function, e.g. when it's created from
  // the Call function block.
  onFunctionCreated?: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useFunctionCreateMutation(appId);
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
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Function created!");
            setOpen(false);
            form.reset();

            if (onFunctionCreated) {
              onFunctionCreated(res.data.id);
              return;
            }

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
              `Failed to create function: ${res.error.message} (${res.error.code})`
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
          <DialogTitle>Create Function</DialogTitle>
          <DialogDescription>
            Create a reusable function that your commands, event listeners and
            message templates can call with the Call function block.
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
            <FormField
              control={form.control}
              name="description"
              rules={{
                maxLength: { value: 100, message: "Up to 100 characters" },
              }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormDescription>
                    What does this function do? This is optional.
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
                Create function
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
        id: getNodeId(),
        position: { x: 0, y: 0 },
        data: {
          name,
          description: description || undefined,
          function_parameters: [],
        },
        type: "entry_function",
      },
    ],
    edges: [],
  };
}
