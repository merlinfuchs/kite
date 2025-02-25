import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAppUpdateMutation } from "@/lib/api/mutations";
import { setValidationErrors } from "@/lib/form";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { ExternalLinkIcon } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";

interface FormFields {
  name: string;
  description: string;
  enabled: boolean;
}

export default function AppSettingsAppearance() {
  const app = useApp();

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      description: "",
      enabled: false,
    },
  });

  useEffect(() => {
    if (app) {
      form.reset({
        name: app.name,
        description: app.description || "",
        enabled: app.enabled,
      });
    }
  }, [app, form]);

  const updateMutation = useAppUpdateMutation(useAppId());

  const onSubmit = useCallback(
    (data: FormFields) => {
      updateMutation.mutate(
        {
          name: data.name,
          description: data.description || null,
          enabled: data.enabled,
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
    },
    [form, updateMutation]
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Appearance</CardTitle>
        <CardDescription>
          Configure how your app appears to users in Discord and Kite.
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
                    <Textarea {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>

          <CardFooter className="flex flex-wrap border-t px-6 py-4 gap-3">
            <Button type="submit">Save settings</Button>
            <Button
              variant="outline"
              type="button"
              className="flex gap-2"
              asChild
            >
              <Link
                href={`https://discord.com/developers/applications/${app?.discord_id}`}
                target="_blank"
              >
                <div>Manage on Discord</div>
                <ExternalLinkIcon className="w-4 h-4" />
              </Link>
            </Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
