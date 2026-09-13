import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Switch } from "@/components/ui/switch";
import { useAppStatusUpdateMutation } from "@/lib/api/mutations";
import { useAppQuery } from "@/lib/api/queries";
import { setValidationErrors } from "@/lib/form";
import { useAppId } from "@/lib/hooks/params";
import { ExternalLinkIcon, PlusIcon, RefreshCwIcon, Trash2Icon } from "lucide-react";
import Link from "next/link";
import { useCallback, useMemo } from "react";
import { useFieldArray, useForm } from "react-hook-form";
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

interface StatusFieldValues {
  id: string;
  label: string;
  status: string;
  activity_type: string;
  activity_name: string;
  activity_url: string;
}

interface FormFields {
  statuses: StatusFieldValues[];
  active_id: string;
  rotate_enabled: boolean;
}

// generateLocalId gives a new status a stable identifier as soon as it's
// added, so it can be selected as the active status (or referenced in the
// rotation) before the form is ever saved. The backend accepts client-chosen
// IDs as-is and only generates its own when one is missing.
function generateLocalId(): string {
  return `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`;
}

function emptyStatus(): StatusFieldValues {
  return {
    id: generateLocalId(),
    label: "",
    status: "online",
    activity_type: "0",
    activity_name: "",
    activity_url: "",
  };
}

export default function AppSettingsPresence() {
  const appId = useAppId();
  const appQuery = useAppQuery(appId);
  const app = appQuery.data?.success ? appQuery.data.data : undefined;

  // isLoading (unlike isFetching) is only true while there's no data to show
  // yet -- it covers both the very first load AND a remount after React
  // Query has evicted its cached copy of the app (e.g. after a few minutes
  // away on another page). Without this, the form's empty-by-default state
  // was indistinguishable from "this app genuinely has no statuses", so the
  // page briefly looked like your statuses had disappeared every time you
  // navigated back to it.
  const isAppLoading = appQuery.isLoading;

  // Derive the form's values from the fetched app. Using RHF's `values`
  // option (rather than `defaultValues` + a manual `reset()` in a
  // useEffect) is the pattern React Hook Form recommends for syncing a form
  // to async data: `reset()` reliably updates plain fields but does NOT
  // reliably sync `useFieldArray`'s own internal field list, so the
  // rotate-enabled toggle would end up correct while the status list stayed
  // stuck empty. `values` keeps both in sync together.
  const formValues = useMemo<FormFields>(() => {
    const statuses =
      app?.discord_status?.statuses?.map((s) => ({
        id: s.id,
        label: s.label || "",
        status: s.status || "online",
        activity_type: s.activity_type?.toString() || "0",
        activity_name: s.activity_name || "",
        activity_url: s.activity_url || "",
      })) || [];

    return {
      statuses,
      active_id: app?.discord_status?.active_id || statuses[0]?.id || "",
      rotate_enabled: app?.discord_status?.rotate_enabled || false,
    };
  }, [app]);

  const form = useForm<FormFields>({
    values: formValues,
    // Don't clobber an in-progress edit if a background refetch resolves
    // while the user is mid-change (e.g. the always-refetch-on-mount fetch
    // landing right as they start editing).
    resetOptions: { keepDirtyValues: true },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "statuses",
    keyName: "fieldKey",
  });

  const activeId = form.watch("active_id");
  const rotateEnabled = form.watch("rotate_enabled");

  const updateMutation = useAppStatusUpdateMutation(appId);

  const onSubmit = useCallback(
    (data: FormFields) => {
      updateMutation.mutate(
        {
          discord_status:
            data.statuses.length > 0
              ? {
                  statuses: data.statuses.map((s) => ({
                    id: s.id,
                    label: s.label || undefined,
                    status: s.status,
                    activity_type: parseInt(s.activity_type) || undefined,
                    activity_name: s.activity_name || undefined,
                    activity_state: s.activity_name || undefined,
                    activity_url: s.activity_url || undefined,
                  })),
                  active_id: data.active_id || undefined,
                  rotate_enabled: data.rotate_enabled,
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

  const handleAddStatus = useCallback(() => {
    const status = emptyStatus();
    append(status);
    if (!activeId) {
      form.setValue("active_id", status.id);
    }
  }, [append, activeId, form]);

  const handleRemoveStatus = useCallback(
    (index: number, id: string) => {
      const remaining = form
        .getValues("statuses")
        .filter((_, i) => i !== index);

      remove(index);

      if (activeId === id) {
        form.setValue("active_id", remaining[0]?.id || "");
      }

      if (remaining.length < 2) {
        form.setValue("rotate_enabled", false);
      }
    },
    [remove, activeId, form]
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Custom Status</CardTitle>
        <CardDescription>
          Configure the status and activity of your app in Discord. Add
          multiple statuses, pick which one is active, or have your app
          rotate through all of them automatically.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
          <CardContent className="space-y-5">
            {isAppLoading ? (
              <p className="text-sm text-muted-foreground">
                Loading your status settings...
              </p>
            ) : (
              <>
                <div className="flex items-center justify-between gap-3 rounded-lg border p-3">
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2 text-sm font-medium">
                      <RefreshCwIcon className="h-4 w-4" />
                      Rotate status every minute
                    </div>
                    <p className="text-sm text-muted-foreground">
                      When enabled, your app cycles through every status
                      below, switching to the next one once a minute. When
                      disabled, the status selected below is shown at all
                      times.
                    </p>
                  </div>
                  <FormField
                    control={form.control}
                    name="rotate_enabled"
                    render={({ field }) => (
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                        disabled={fields.length < 2}
                      />
                    )}
                  />
                </div>

                <RadioGroup
                  value={activeId}
                  onValueChange={(value) => form.setValue("active_id", value)}
                  className="space-y-4"
                >
                  {fields.map((field, index) => (
                    <div
                      key={field.fieldKey}
                  className="rounded-lg border p-4 space-y-4"
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2">
                      <RadioGroupItem
                        value={field.id}
                        id={`active-${field.fieldKey}`}
                        disabled={rotateEnabled}
                      />
                      <label
                        htmlFor={`active-${field.fieldKey}`}
                        className="text-sm text-muted-foreground"
                      >
                        {rotateEnabled
                          ? "Included in rotation"
                          : "Active status"}
                      </label>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      type="button"
                      onClick={() => handleRemoveStatus(index, field.id)}
                    >
                      <Trash2Icon className="h-4 w-4 text-muted-foreground" />
                    </Button>
                  </div>

                  <FormField
                    control={form.control}
                    name={`statuses.${index}.label`}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Label</FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="e.g. Working, Watching for commands"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className="flex gap-3">
                    <FormField
                      control={form.control}
                      name={`statuses.${index}.status`}
                      render={({ field }) => (
                        <FormItem className="min-w-48">
                          <FormLabel>Status</FormLabel>
                          <Select
                            onValueChange={field.onChange}
                            value={field.value}
                          >
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Select custom status" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="online">Online</SelectItem>
                              <SelectItem value="dnd">Do Not Disturb</SelectItem>
                              <SelectItem value="idle">AFK</SelectItem>
                              <SelectItem value="invisible">
                                Invisible
                              </SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name={`statuses.${index}.activity_type`}
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
                  </div>

                  <FormField
                    control={form.control}
                    name={`statuses.${index}.activity_name`}
                    render={({ field }) => (
                      <FormItem className="w-full">
                        <FormLabel>Activity Name</FormLabel>
                        <FormControl>
                          <Input type="text" className="w-full" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name={`statuses.${index}.activity_url`}
                    render={({ field }) => (
                      <FormItem className="w-full">
                        <FormLabel>Activity URL</FormLabel>
                        <FormControl>
                          <Input type="url" className="w-full" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              ))}
                </RadioGroup>

                <Button
                  variant="outline"
                  type="button"
                  onClick={handleAddStatus}
                  className="w-full"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Add status
                </Button>

                {fields.length === 0 && (
                  <p className="text-sm text-muted-foreground">
                    No custom statuses configured. Your app will show
                    Kite&apos;s default status until you add one.
                  </p>
                )}
              </>
            )}
          </CardContent>

          <CardFooter className="flex flex-wrap items-center border-t px-6 py-4 gap-3">
            <Button type="submit">Update status</Button>
            <Link
              href="https://discord.com/developers/docs/topics/gateway-events#update-presence"
              target="_blank"
              className="flex items-center gap-1 text-sm text-muted-foreground hover:underline"
            >
              Learn more about Discord presences
              <ExternalLinkIcon className="h-3 w-3" />
            </Link>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
