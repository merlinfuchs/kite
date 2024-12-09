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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import Link from "next/link";
import { ExternalLinkIcon } from "lucide-react";

interface FormFields {
  name: string;
  description: string;
  enabled: boolean;
  discord_status: {
    status: string;
    activity_type?: string;
    activity_name?: string;
    activity_url?: string;
  };
}

export default function AppSettingsAppearance() {
  const app = useApp();

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      description: "",
      enabled: false,
      discord_status: {
        status: "",
        activity_type: "0",
        activity_name: "",
        activity_url: "",
      },
    },
  });

  useEffect(() => {
    if (app) {
      form.reset({
        name: app.name,
        description: app.description || "",
        enabled: app.enabled,
        discord_status: {
          status: app.discord_status?.status || "",
          activity_type: app.discord_status?.activity_type?.toString() || "0",
          activity_name: app.discord_status?.activity_name || "",
          activity_url: app.discord_status?.activity_url || "",
        },
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
          discord_status: !!data.discord_status.status
            ? {
                status: data.discord_status.status,
                activity_type:
                  parseInt(data.discord_status.activity_type || "0") ||
                  undefined,
                activity_name: data.discord_status.activity_name || undefined,
                activity_state: data.discord_status.activity_name || undefined,
                activity_url: data.discord_status.activity_url || undefined,
              }
            : undefined,
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

  const discordStatus = form.watch("discord_status.status");

  return (
    <Card x-chunk="dashboard-04-chunk-1">
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
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="flex space-x-3 items-end">
              <FormField
                control={form.control}
                name="discord_status.status"
                render={({ field }) => (
                  <FormItem className="min-w-48">
                    <FormLabel>Custom Status</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select custom status" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="online">Online</SelectItem>
                        <SelectItem value="dnd">Do Not Disturb</SelectItem>
                        <SelectItem value="idle">AFK</SelectItem>
                        <SelectItem value="invisible">Invisible</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button
                variant="outline"
                type="button"
                onClick={() => form.setValue("discord_status.status", "")}
              >
                Clear
              </Button>
            </div>
            {discordStatus && (
              <div className="flex space-x-3">
                <FormField
                  control={form.control}
                  name="discord_status.activity_type"
                  render={({ field }) => (
                    <FormItem className="min-w-48">
                      <FormLabel>Activity Type</FormLabel>
                      <Select
                        onValueChange={field.onChange}
                        value={field.value}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select a custom status inside Discord for your app" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="0">Playing</SelectItem>
                          <SelectItem value="1">Streaming</SelectItem>
                          <SelectItem value="2">Listening</SelectItem>
                          <SelectItem value="3">Watching</SelectItem>
                          <SelectItem value="5">Competing</SelectItem>
                          <SelectItem value="4">Custom</SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="discord_status.activity_name"
                  render={({ field }) => (
                    <FormItem className="w-full">
                      <FormLabel>Activity Name</FormLabel>
                      <FormControl>
                        <Input type="text" className="w-full" {...field} />
                      </FormControl>
                      <FormMessage />
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="discord_status.activity_url"
                  render={({ field }) => (
                    <FormItem className="w-full">
                      <FormLabel>Activity URL</FormLabel>
                      <FormControl>
                        <Input type="url" className="w-full" {...field} />
                      </FormControl>
                      <FormMessage />
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            )}
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
