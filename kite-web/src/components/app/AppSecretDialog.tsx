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
import {
  useAppSecretCreateMutation,
  useAppSecretUpdateMutation,
} from "@/lib/api/mutations";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { AppSecret } from "@/lib/types/wire.gen";

interface FormFields {
  name: string;
  value: string;
}

// Creates a secret, or edits one when given. Values can't be read back, so an
// existing secret's value is only replaced when a new one is entered.
export default function AppSecretDialog({
  children,
  secret,
}: {
  children: ReactNode;
  secret?: AppSecret;
}) {
  const [open, setOpen] = useState(false);
  const appId = useAppId();

  const createMutation = useAppSecretCreateMutation(appId);
  const updateMutation = useAppSecretUpdateMutation(appId, secret?.id ?? "");
  const pending = createMutation.isPending || updateMutation.isPending;

  const form = useForm<FormFields>({
    defaultValues: { name: secret?.name ?? "", value: "" },
  });

  function onSubmit(data: FormFields) {
    if (pending) return;

    const handlers = {
      onSuccess(res: Awaited<ReturnType<typeof createMutation.mutateAsync>>) {
        if (res.success) {
          toast.success(secret ? "Secret updated!" : "Secret created!");
          setOpen(false);
          form.reset({ name: secret ? res.data.name : "", value: "" });
        } else if (res.error.code === "validation_failed") {
          setValidationErrors(form, res.error.data);
        } else {
          toast.error(
            `Failed to save secret: ${res.error.message} (${res.error.code})`
          );
        }
      },
    };

    if (secret) {
      updateMutation.mutate(
        { name: data.name, value: data.value || null },
        handlers
      );
    } else {
      createMutation.mutate({ name: data.name, value: data.value }, handlers);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        // Entered values aren't kept around after closing.
        if (!open) form.reset({ name: secret?.name ?? "", value: "" });
      }}
    >
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{secret ? "Edit Secret" : "Create Secret"}</DialogTitle>
          <DialogDescription>
            Use it in API request blocks as{" "}
            <code>{`{{secrets.${form.watch("name") || "NAME"}}}`}</code>. The
            value can&apos;t be read back once it&apos;s saved.
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
                    <Input
                      type="text"
                      placeholder="API_KEY"
                      {...field}
                      onChange={(e) =>
                        field.onChange(e.target.value.toUpperCase())
                      }
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Value</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="off" {...field} />
                  </FormControl>
                  {secret && (
                    <FormDescription>
                      Leave empty to keep the current value.
                    </FormDescription>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={pending}>
                {secret ? "Save secret" : "Create secret"}
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
