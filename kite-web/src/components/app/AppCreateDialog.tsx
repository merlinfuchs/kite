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
import { useAppCreateMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { CircleHelpIcon } from "lucide-react";
import env from "@/lib/env/client";

interface FormFields {
  discord_token: string;
}

export default function AppCreateDialog({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);

  const createMutation = useAppCreateMutation();
  const form = useForm<FormFields>({
    defaultValues: {
      discord_token: "",
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    createMutation.mutate(data, {
      onSuccess(res) {
        if (res.success) {
          toast.success("App created!");
          setOpen(false);
        } else {
          if (res.error.code === "validation_failed") {
            setValidationErrors(form, res.error.data);
          } else {
            toast.error(
              `Failed to create app: ${res.error.message} (${res.error.code})`
            );
          }
        }
      },
    });
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            <div className="flex items-center space-x-2">
              <div>Create App</div>
              <a
                href={env.NEXT_PUBLIC_DOCS_LINK + "/guides/getting-started"}
                target="_blank"
              >
                <CircleHelpIcon className="h-5 w-5" />
              </a>
            </div>
          </DialogTitle>
          <DialogDescription>
            Create a new app to get started with Kite and start building your
            app!
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <FormField
              control={form.control}
              name="discord_token"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Bot Token</FormLabel>
                  <FormControl>
                    <Input type="password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Create app
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
