import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAppTokenUpdateMutation } from "@/lib/api/mutations";
import { setValidationErrors } from "@/lib/form";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useCallback, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../ui/alert-dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Input } from "../ui/input";

interface FormFields {
  discord_token: string;
}

export default function AppSettingsCredentials() {
  const app = useApp();

  const form = useForm<FormFields>({
    defaultValues: {
      discord_token: "",
    },
  });

  useEffect(() => {
    if (app) {
      form.reset({
        discord_token: "",
      });
    }
  }, [app, form]);

  const updateMutation = useAppTokenUpdateMutation(useAppId());

  // Set when the token belongs to a different Discord app, so the user can
  // confirm the switch before it happens.
  const [pendingChange, setPendingChange] = useState<{
    token: string;
    message: string;
  } | null>(null);

  const saveToken = useCallback(
    (token: string, changeApp: boolean) => {
      updateMutation.mutate(
        {
          discord_token: token,
          change_app: changeApp,
        },
        {
          onSuccess(res) {
            if (res.success) {
              if (changeApp) {
                toast.success(`Switched to ${res.data.name}!`);
              } else {
                toast.success("Settings saved!");
              }
            } else if (res.error.code === "discord_app_changed") {
              setPendingChange({ token, message: res.error.message });
            } else if (res.error.code === "validation_failed") {
              setValidationErrors(form, res.error.data);
            } else {
              toast.error(
                `Failed to update app: ${res.error.message} (${res.error.code})`
              );
            }
          },
        }
      );
    },
    [form, updateMutation]
  );

  const onSubmit = useCallback(
    (data: FormFields) => {
      if (!data.discord_token) return;
      saveToken(data.discord_token, false);
    },
    [saveToken]
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Credentials</CardTitle>
        <CardDescription>
          Configure your app&apos;s credentials here. This is where you can
          change your app&apos;s Discord token, or move your app to a different
          Discord app by entering its token.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
          <CardContent className="space-y-5">
            <FormField
              control={form.control}
              name="discord_token"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Discord Token</FormLabel>
                  <FormControl>
                    <Input type="password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>

          <CardFooter className="border-t px-6 py-4">
            <Button
              type="submit"
              disabled={
                !form.getValues().discord_token || updateMutation.isPending
              }
            >
              Save token
            </Button>
          </CardFooter>
        </form>
      </Form>

      <AlertDialog
        open={pendingChange !== null}
        onOpenChange={(open) => {
          if (!open) setPendingChange(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Switch to a different Discord app?
            </AlertDialogTitle>
            <AlertDialogDescription asChild>
              <div className="space-y-2">
                <p>{pendingChange?.message}.</p>
                <p>
                  Your commands, event listeners, messages and variables stay,
                  and your commands are deployed to the new bot. The name and
                  description are taken from the new Discord app.
                </p>
                <p>
                  The new bot has to be invited to your servers again, and
                  custom emojis of the old app won&apos;t work anymore. The old
                  bot goes offline and its commands are removed.
                </p>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (pendingChange) saveToken(pendingChange.token, true);
                setPendingChange(null);
              }}
            >
              Switch app
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}
