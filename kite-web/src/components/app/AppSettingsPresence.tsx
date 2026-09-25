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
import { activityTypeOptions, statusOptions } from "@/lib/discord/presence";
import { useAppFeature } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { cn, getUniqueId } from "@/lib/utils";
import {
  ChevronDownIcon,
  ExternalLinkIcon,
  PlusIcon,
  RefreshCwIcon,
  Trash2Icon,
} from "lucide-react";
import Link from "next/link";
import { useCallback, useMemo, useState } from "react";
import {
  useFieldArray,
  useForm,
  useFormContext,
  useWatch,
} from "react-hook-form";
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

const maxStatuses = 10;

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

function emptyStatus(): StatusFieldValues {
  return {
    // Needed right away so the status can be picked as the active one
    id: getUniqueId().toString(),
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

  const isAppLoading = appQuery.isLoading;
  const rotatingStatusAvailable = !!useAppFeature((f) => f.rotating_status);

  // Passed as `values` instead of calling reset(), which doesn't reliably
  // update the useFieldArray list.
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
    resetOptions: { keepDirtyValues: true },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "statuses",
    keyName: "fieldKey",
  });

  const activeId = form.watch("active_id");
  const rotateEnabled = rotatingStatusAvailable && form.watch("rotate_enabled");

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

  const [expandedId, setExpandedId] = useState<string>();

  const handleAddStatus = useCallback(() => {
    const status = emptyStatus();
    append(status);
    if (!activeId) {
      form.setValue("active_id", status.id);
    }
    setExpandedId(status.id);
  }, [append, activeId, form]);

  const handleRemoveStatus = useCallback(
    (index: number) => {
      if (fields[index].id === activeId) {
        form.setValue("active_id", fields[index === 0 ? 1 : 0]?.id || "");
      }
      remove(index);
    },
    [fields, remove, activeId, form]
  );

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
            {isAppLoading ? (
              <p className="text-sm text-muted-foreground">
                Loading your status settings...
              </p>
            ) : fields.length === 0 ? (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  Your app shows Kite&apos;s default status.
                </p>
                <Button
                  variant="outline"
                  type="button"
                  onClick={handleAddStatus}
                >
                  Set custom status
                </Button>
              </div>
            ) : fields.length === 1 ? (
              <>
                <StatusFields index={0} />
                <div className="flex gap-3">
                  <Button
                    variant="outline"
                    type="button"
                    onClick={handleAddStatus}
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    Add another status
                  </Button>
                  <Button
                    variant="ghost"
                    type="button"
                    onClick={() => handleRemoveStatus(0)}
                  >
                    Clear
                  </Button>
                </div>
              </>
            ) : (
              <>
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <RefreshCwIcon className="h-4 w-4" />
                    Rotate through all statuses every minute
                    {!rotatingStatusAvailable && (
                      <Link
                        href={`/apps/${appId}/premium`}
                        className="text-xs text-primary hover:underline"
                      >
                        Premium
                      </Link>
                    )}
                  </div>
                  <FormField
                    control={form.control}
                    name="rotate_enabled"
                    render={({ field }) => (
                      <Switch
                        checked={rotateEnabled}
                        onCheckedChange={field.onChange}
                        disabled={!rotatingStatusAvailable}
                      />
                    )}
                  />
                </div>

                <RadioGroup
                  value={activeId}
                  onValueChange={(value) => form.setValue("active_id", value)}
                  className="gap-0 rounded-lg border divide-y"
                >
                  {fields.map((field, index) => {
                    const expanded = expandedId === field.id;

                    return (
                      <div key={field.fieldKey}>
                        <div className="flex items-center gap-3 px-4 py-3">
                          {rotateEnabled ? (
                            <RefreshCwIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                          ) : (
                            <RadioGroupItem
                              value={field.id}
                              aria-label="Active status"
                            />
                          )}
                          <button
                            type="button"
                            className="flex flex-1 min-w-0 items-center gap-3 text-left"
                            aria-expanded={expanded}
                            onClick={() =>
                              setExpandedId(expanded ? undefined : field.id)
                            }
                          >
                            <StatusRowSummary index={index} />
                            <ChevronDownIcon
                              className={cn(
                                "h-4 w-4 shrink-0 text-muted-foreground transition-transform",
                                expanded && "rotate-180"
                              )}
                            />
                          </button>
                          <Button
                            variant="ghost"
                            size="icon"
                            type="button"
                            onClick={() => handleRemoveStatus(index)}
                          >
                            <Trash2Icon className="h-4 w-4 text-muted-foreground" />
                          </Button>
                        </div>
                        {expanded && (
                          <div className="space-y-4 px-4 pb-4">
                            <StatusFields index={index} showLabel />
                          </div>
                        )}
                      </div>
                    );
                  })}
                </RadioGroup>

                <Button
                  variant="outline"
                  type="button"
                  onClick={handleAddStatus}
                  disabled={fields.length >= maxStatuses}
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Add status ({fields.length}/{maxStatuses})
                </Button>
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

function StatusRowSummary({ index }: { index: number }) {
  const status = useWatch<FormFields, `statuses.${number}`>({
    name: `statuses.${index}`,
  });
  // Can lag behind the field array for a render after a status is removed
  if (!status) return null;

  const prefix = activityTypeOptions.find(
    (o) => o.value === status.activity_type
  )?.prefix;
  const activity =
    status.activity_name &&
    [prefix, status.activity_name].filter(Boolean).join(" ");
  const summary = [
    activity,
    statusOptions.find((o) => o.value === status.status)?.label,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="flex-1 min-w-0">
      <div className="text-sm font-medium truncate">
        {status.label || `Status ${index + 1}`}
      </div>
      <div className="text-sm text-muted-foreground truncate">{summary}</div>
    </div>
  );
}

function StatusFields({
  index,
  showLabel,
}: {
  index: number;
  showLabel?: boolean;
}) {
  const form = useFormContext<FormFields>();

  return (
    <>
      {showLabel && (
        <FormField
          control={form.control}
          name={`statuses.${index}.label`}
          render={({ field }) => (
            <FormItem>
              <FormLabel>Label</FormLabel>
              <FormControl>
                <Input
                  type="text"
                  placeholder="e.g. Default, Maintenance"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      )}

      <div className="flex gap-3">
        <FormField
          control={form.control}
          name={`statuses.${index}.status`}
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
                  {statusOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
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
              <Select onValueChange={field.onChange} value={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a custom status inside Discord for your app" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {activityTypeOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
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
    </>
  );
}
