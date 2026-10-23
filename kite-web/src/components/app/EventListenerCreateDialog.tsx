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
import { ReactNode, useMemo, useState } from "react";
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
import { useEventListenerCreateMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { useRouter } from "next/router";
import { getNodeId } from "@/lib/flow/nodes";
import { EventTypeScheduleCron, EventTypeWebhook } from "@/lib/types/flow.gen";
import ScheduleCronPreview, {
  ScheduleCronHelp,
} from "../common/ScheduleCronPreview";
import {
  discordEventTypes,
  schedulePresets,
  capitalize,
  eventSourceName,
  EventListenerSource,
} from "@/lib/flow/eventTypes";
import { cn } from "@/lib/utils";
import { matchesSearch } from "@/lib/search";
import { getSchedulePreview } from "@/lib/flow/schedule";
import { LucideIcon, PencilIcon, SearchIcon } from "lucide-react";
import { track } from "@/lib/analytics";

interface FormFields {
  type: string;
  cron: string;
  description: string;
}

const dialogTexts: Record<
  EventListenerSource,
  { description: string; picker?: string }
> = {
  discord: {
    description: "Pick the Discord event that runs the flow.",
    picker: "Event",
  },
  schedule: {
    description: "Pick how often the flow runs.",
    picker: "Schedule",
  },
  webhook: {
    description:
      "The flow runs whenever another service sends a request to its URL.",
  },
};

export default function EventListenerCreateDialog({
  source,
  children,
}: {
  source: EventListenerSource;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  const router = useRouter();
  const appId = useAppId();

  const createMutation = useEventListenerCreateMutation(appId);
  const form = useForm<FormFields>({
    defaultValues: {
      type: "",
      cron: source === "schedule" ? schedulePresets[0].cron : "",
      description: "",
    },
  });

  const texts = dialogTexts[source];
  const name = eventSourceName(source);

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    const type =
      source === "schedule"
        ? EventTypeScheduleCron
        : source === "webhook"
        ? EventTypeWebhook
        : data.type;

    createMutation.mutate(
      {
        source,
        flow_source: getInitialFlowData(
          type,
          data.description,
          source === "schedule" ? data.cron : undefined
        ),
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            track("event_listener_created", { type, source: "blank" });
            toast.success(`${capitalize(name.singular)} created!`);
            setOpen(false);
            form.reset();

            setTimeout(
              () =>
                router.push({
                  pathname: "/apps/[appId]/events/[eventId]",
                  query: { appId, eventId: res.data.id },
                }),
              500
            );
          } else {
            toast.error(
              `Failed to create ${name.singular}: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-2xl max-h-[90dvh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create {name.title}</DialogTitle>
          <DialogDescription>{texts.description}</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-5">
            {source === "discord" && (
              <FormField
                control={form.control}
                name="type"
                rules={{ required: "Pick an event." }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{texts.picker}</FormLabel>
                    <DiscordEventPicker
                      value={field.value}
                      onChange={field.onChange}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            {source === "schedule" && (
              <FormField
                control={form.control}
                name="cron"
                rules={{ required: "Pick or enter a schedule." }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{texts.picker}</FormLabel>
                    <SchedulePicker
                      value={field.value}
                      onChange={field.onChange}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            {source === "webhook" && <WebhookPicker />}
            <FormField
              control={form.control}
              name="description"
              rules={{ required: "Enter a description." }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormDescription>What will you use it for?</FormDescription>
                  <FormControl>
                    <Input type="text" maxLength={100} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Create {name.title}
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

// OptionCard is one choice of a picker. All three pickers are built from it,
// so they look the same.
function OptionCard({
  icon: Icon,
  title,
  description,
  selected,
  onSelect,
  compact,
}: {
  icon: LucideIcon;
  title: string;
  description: string;
  selected?: boolean;
  onSelect?: () => void;
  // Only shows the description once picked, so long lists stay short.
  compact?: boolean;
}) {
  const className = cn(
    "flex items-center gap-3 rounded-lg border p-3 text-left transition-colors w-full",
    selected
      ? "border-primary bg-primary/5 ring-1 ring-primary"
      : onSelect && "hover:bg-muted/50"
  );

  const content = (
    <>
      <div
        className={cn(
          "rounded-md p-2 shrink-0",
          selected ? "bg-primary text-primary-foreground" : "bg-muted"
        )}
      >
        <Icon className="h-4 w-4" />
      </div>
      <div className="min-w-0">
        <div className="text-sm font-medium leading-tight">{title}</div>
        {(!compact || selected) && (
          <div className="text-xs text-muted-foreground mt-0.5 break-words">
            {description}
          </div>
        )}
      </div>
    </>
  );

  if (!onSelect) {
    return <div className={className}>{content}</div>;
  }

  return (
    <button
      type="button"
      className={className}
      onClick={onSelect}
      aria-pressed={selected}
      title={compact ? description : undefined}
    >
      {content}
    </button>
  );
}

function PickerNote({ children }: { children: ReactNode }) {
  return (
    <div className="text-sm text-muted-foreground rounded-lg bg-muted/50 px-3 py-2">
      {children}
    </div>
  );
}

function DiscordEventPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const [search, setSearch] = useState("");

  const categories = useMemo(() => {
    const res: { name: string; types: typeof discordEventTypes }[] = [];
    for (const type of discordEventTypes) {
      if (
        !matchesSearch(search, [type.label, type.description, type.category])
      ) {
        continue;
      }
      let category = res.find((c) => c.name === type.category);
      if (!category) {
        category = { name: type.category, types: [] };
        res.push(category);
      }
      category.types.push(type);
    }
    return res;
  }, [search]);

  const selected = discordEventTypes.find((t) => t.value === value);

  return (
    <div className="space-y-3">
      <div className="relative">
        <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search events"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>
      <div className="space-y-4">
        {categories.length === 0 && (
          <div className="text-sm text-muted-foreground text-center py-8">
            No events match your search.
          </div>
        )}
        {categories.map((category) => (
          <div key={category.name} className="space-y-2">
            <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              {category.name}
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              {category.types.map((type) => (
                <OptionCard
                  key={type.value}
                  icon={type.icon}
                  title={type.label}
                  description={type.description}
                  selected={type.value === value}
                  onSelect={() => onChange(type.value)}
                  compact
                />
              ))}
            </div>
          </div>
        ))}
      </div>
      {selected?.note && <PickerNote>{selected.note}</PickerNote>}
    </div>
  );
}

function SchedulePicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const preset = schedulePresets.find((p) => p.cron === value.trim());
  // Custom stays picked while the expression is edited, even if it happens to
  // match a preset on the way.
  const [custom, setCustom] = useState(false);
  const isCustom = custom || !preset;
  // Custom has its own expression, so picking it doesn't carry over the
  // preset that was picked before, and switching back keeps what was typed.
  const [customCron, setCustomCron] = useState("");
  const isValid = useMemo(() => !!getSchedulePreview(value), [value]);

  return (
    <div className="space-y-3">
      <div className="grid gap-2 sm:grid-cols-2">
        {schedulePresets.map((p) => (
          <OptionCard
            key={p.cron}
            icon={p.icon}
            title={p.label}
            description={p.description}
            selected={!isCustom && preset?.cron === p.cron}
            onSelect={() => {
              setCustom(false);
              onChange(p.cron);
            }}
          />
        ))}
        <OptionCard
          icon={PencilIcon}
          title="Custom"
          description="Write your own cron expression"
          selected={isCustom}
          onSelect={() => {
            setCustom(true);
            onChange(customCron);
          }}
        />
      </div>
      {isCustom && (
        <div className="space-y-1.5">
          <Input
            type="text"
            placeholder="*/5 * * * *"
            value={value}
            onChange={(e) => {
              setCustomCron(e.target.value);
              onChange(e.target.value);
            }}
            autoFocus
          />
          <p className="text-sm text-muted-foreground">
            <ScheduleCronHelp />
          </p>
        </div>
      )}
      <PickerNote>
        {!value.trim() ? (
          "Enter a cron expression to see when it runs."
        ) : isValid ? (
          <ScheduleCronPreview cron={value} />
        ) : (
          "This isn't a valid cron expression."
        )}
      </PickerNote>
    </div>
  );
}

function WebhookPicker() {
  return (
    <PickerNote>
      Once it&apos;s created, copy the URL from the settings of the first block
      and enter it wherever the other service asks for a webhook URL.
    </PickerNote>
  );
}

function getInitialFlowData(type: string, description: string, cron?: string) {
  return {
    nodes: [
      {
        id: getNodeId("entry_event"),
        position: { x: 0, y: 0 },
        data: {
          event_type: type,
          event_schedule_cron: cron || undefined,
          description,
        },
        type: "entry_event",
      },
    ],
    edges: [],
  };
}
