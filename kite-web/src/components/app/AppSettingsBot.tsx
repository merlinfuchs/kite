import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "../ui/input";
import { useForm } from "react-hook-form";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { useCallback, useEffect } from "react";
import { useApp } from "@/lib/hooks/api";
import { useAppUpdateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";

interface FormFields {
  name: string;
  description: string;
  discord_token: string;
}

export default function AppSettingsBot() {
  const app = useApp();

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      description: "",
      discord_token: "",
    },
  });

  useEffect(() => {
    if (app) {
      form.reset({
        name: app.name,
        description: app.description || "",
        discord_token: "",
      });
    }
  }, [app]);

  const updateMutation = useAppUpdateMutation(useAppId());

  const onSubmit = useCallback((data: FormFields) => {
    updateMutation.mutate(
      {
        name: data.name || null,
        description: data.description,
        discord_token: data.discord_token || null,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Settings saved!");
          } else {
            if (res.error.code === "validation_failed") {
              setValidationErrors(form, res.error.data);
            } else {
              toast.error(
                `Failed to update app: ${res.error.message} (${res.error.code})`
              );
            }
          }
        },
      }
    );
  }, []);

  return (
    <Card x-chunk="dashboard-04-chunk-1">
      <CardHeader>
        <CardTitle>Bot Settings</CardTitle>
        <CardDescription>
          Configure the settings for your Discord bot.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
          <CardContent className="space-y-5">
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
            <FormField
              control={form.control}
              name="discord_token"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Bot Token</FormLabel>
                  <FormDescription>
                    Set a new Discord bot token for your app.
                  </FormDescription>
                  <FormControl>
                    <Input type="password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>

          <CardFooter className="border-t px-6 py-4">
            <Button type="submit">Save settings</Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
