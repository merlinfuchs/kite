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
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { useMessageCreateMutation } from "@/lib/api/mutations";

interface FormFields {
  name: string;
  description: string;
}

const createDefaultMessage = (name: string) => ({
  content: `You message template: **${name}**`,
  tts: false,
  embeds: [
    {
      id: 174831423,
      description:
        "You can use message templates to create Discord messages that contain embeds and interactive components (soon). They can be used to create standalone messages or as a response to commands or other events.",
      color: 16735232,
      author: {
        name: "About Message Templates",
      },
      thumbnail: {
        url: "https://kite.onl/logo.png",
      },
      fields: [],
    },
  ],
  components: [],
  actions: {},
});

// TODO: forwardref
export default function MessageCreateDialog({
  children,
  onMessageCreated,
}: {
  children: ReactNode;
  onMessageCreated?: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);

  const appId = useAppId();

  const createMutation = useMessageCreateMutation(appId);
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
        name: data.name,
        description: data.description || null,
        data: createDefaultMessage(data.name),
        flow_sources: {},
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Message created!");
            setOpen(false);

            if (onMessageCreated) {
              onMessageCreated(res.data.id);
            }
          } else {
            if (res.error.code === "validation_failed") {
              setValidationErrors(form, res.error.data);
            } else {
              toast.error(
                `Failed to create message: ${res.error.message} (${res.error.code})`
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
          <DialogTitle>Create message</DialogTitle>
          <DialogDescription>
            Create a message that can be sent and used as a response in your
            app.
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
                Create message
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
