import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAppStatusUpdateMutation } from "@/lib/api/mutations";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

interface FormFields {
  discord_status: {
    status: string;
    activity_type?: string;
    activity_name?: string;
    activity_url?: string;
  };
}

export default function AppSettingsPresence() {
  const app = useApp();

  const form = useForm<FormFields>({
    defaultValues: {
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
        discord_status: {
          status: app.discord_status?.status || "",
          activity_type: app.discord_status?.activity_type?.toString() || "0",
          activity_name: app.discord_status?.activity_name || "",
          activity_url: app.discord_status?.activity_url || "",
        },
      });
    }
  }, [app, form]);

  const updateMutation = useAppStatusUpdateMutation(useAppId());

  const onSubmit = useCallback(
    (data: FormFields) => {
      updateMutation.mutate(
        {
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
              toast.success(
                "Status updated! It may take a few minutes to take effect."
              );
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
    <Card>
      <CardHeader>
        <CardTitle>Custom Status</CardTitle>
        <CardDescription>
          Configure the status and activity of your app in Discord.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
          <CardContent className="space-y-5">
            <div className="flex space-x-3 items-end">
              <FormField
                control={form.control}
                name="discord_status.status"
                render={({ field }) => (
                  <FormItem className="min-w-48">
                    <FormLabel>Status</FormLabel>
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
              <>
                <div className="flex gap-3">
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
                </div>
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
              </>
            )}
          </CardContent>

          <CardFooter className="flex flex-wrap border-t px-6 py-4 gap-3">
            <Button type="submit">Update status</Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
