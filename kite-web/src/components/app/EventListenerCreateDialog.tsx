import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Button } from "@/components/ui/button";
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
import { newEventListenerFlow } from "@/lib/flow/newFlows";
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
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  LayersIcon,
  LucideIcon,
  MessageSquareIcon,
  MicIcon,
  PencilIcon,
  ScrollTextIcon,
  SearchIcon,
  SmileIcon,
  UsersIcon,
} from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { track } from "@/lib/analytics";
import { useIsMobile } from "@/lib/hooks/use-mobile";

interface FormFields {
  type: string;
  cron: string;
  description: string;
}

const dialogTexts: Record<
  EventListenerSource,
  { description: string; picker: string }
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
    picker: "Webhook",
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
  const [step, setStep] = useState<1 | 2>(1);
  const [direction, setDirection] = useState<"forward" | "backward">("forward");
  const isMobile = useIsMobile();

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

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setTimeout(() => {
        setStep(1);
        setDirection("forward");
        form.reset({
          type: "",
          cron: source === "schedule" ? schedulePresets[0].cron : "",
          description: "",
        });
      }, 250);
    }
  }

  function handleNext() {
    if (source === "discord" && !form.getValues("type")) {
      form.setError("type", { message: "Pick an event to continue." });
      return;
    }
    if (source === "schedule" && !form.getValues("cron")) {
      form.setError("cron", { message: "Pick or enter a schedule." });
      return;
    }
    form.clearErrors();
    setDirection("forward");
    setStep(2);
  }

  function handleBack() {
    setDirection("backward");
    setStep(1);
  }

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
        flow_source: newEventListenerFlow(
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
            toast.success(`${capitalize(name.singular)} created.`);
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

  // Preview data for mobile Step 2 confirmation card
  const selectedDiscordEvent = discordEventTypes.find(
    (t) => t.value === form.watch("type")
  );
  const selectedSchedulePreset = schedulePresets.find(
    (p) => p.cron === form.watch("cron")?.trim()
  );

  const pickerFields = (
    <div className="flex-1 min-h-0 flex flex-col">
      {source === "discord" && (
        <FormField
          control={form.control}
          name="type"
          rules={{ required: "Pick an event." }}
          render={({ field }) => (
            <FormItem className="flex-1 min-h-0 flex flex-col">
              <DiscordEventPicker
                value={field.value}
                onChange={(val) => {
                  field.onChange(val);
                  form.clearErrors("type");
                }}
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
            <FormItem className="flex-1 min-h-0 flex flex-col">
              <SchedulePicker
                value={field.value}
                onChange={(val) => {
                  field.onChange(val);
                  form.clearErrors("cron");
                }}
              />
              <FormMessage />
            </FormItem>
          )}
        />
      )}
      {source === "webhook" && <WebhookPicker />}
    </div>
  );

  const descriptionField = (
    <FormField
      control={form.control}
      name="description"
      rules={{ required: "Enter a description." }}
      render={({ field }) => (
        <FormItem className="flex-none pt-1">
          <FormLabel>Description</FormLabel>
          <FormDescription>What will you use it for?</FormDescription>
          <FormControl>
            <Input
              type="text"
              placeholder="e.g. Welcome new members and send rules"
              maxLength={100}
              {...field}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );

  // Mobile 2-step views
  const mobileStep1Content = (
    <div
      key="mobile-step-1"
      className={cn(
        "flex-1 min-h-0 flex flex-col",
        direction === "backward"
          ? "animate-in fade-in slide-in-from-left-4 duration-200"
          : "animate-in fade-in duration-150"
      )}
    >
      {pickerFields}
    </div>
  );

  const mobileStep2Content = (
    <div
      key="mobile-step-2"
      className="flex-1 min-h-0 overflow-y-auto space-y-5 animate-in fade-in slide-in-from-right-4 duration-200 p-1"
    >
      {/* Selected choice summary card */}
      {source === "discord" && selectedDiscordEvent && (
        <div className="space-y-2">
          <div className="flex items-center gap-3 p-3.5 rounded-lg border border-border/60 bg-muted/30">
            <div className="rounded-md w-8 h-8 flex items-center justify-center shrink-0 bg-primary/10 text-primary">
              <selectedDiscordEvent.icon className="size-4" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-sm font-medium leading-5 truncate text-foreground">
                {selectedDiscordEvent.label}
              </div>
              <div className="text-xs text-muted-foreground truncate">
                {selectedDiscordEvent.description}
              </div>
            </div>
          </div>
          {selectedDiscordEvent.note && (
            <PickerNote>{selectedDiscordEvent.note}</PickerNote>
          )}
        </div>
      )}

      {source === "schedule" && (
        <div className="space-y-2">
          <div className="flex items-center gap-3 p-3.5 rounded-lg border border-border/60 bg-muted/30">
            <div className="rounded-md w-8 h-8 flex items-center justify-center shrink-0 bg-primary/10 text-primary">
              {selectedSchedulePreset?.icon ? (
                <selectedSchedulePreset.icon className="size-4" />
              ) : (
                <PencilIcon className="size-4" />
              )}
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-sm font-medium leading-5 truncate text-foreground">
                {selectedSchedulePreset?.label ?? "Custom schedule"}
              </div>
              <div className="text-xs text-muted-foreground truncate">
                {form.watch("cron")}
              </div>
            </div>
          </div>
        </div>
      )}

      {source === "webhook" && (
        <div className="p-3.5 rounded-lg border border-border/60 bg-muted/30">
          <div className="text-sm font-medium text-foreground">
            Webhook endpoint
          </div>
          <div className="text-xs text-muted-foreground mt-0.5">
            A unique incoming URL will be generated upon creation.
          </div>
        </div>
      )}

      <FormField
        control={form.control}
        name="description"
        rules={{ required: "Enter a description." }}
        render={({ field }) => (
          <FormItem className="space-y-2">
            <div>
              <FormLabel>Description</FormLabel>
              <FormDescription className="mt-0.5">
                What will you use it for?
              </FormDescription>
            </div>
            <FormControl>
              <Input
                type="text"
                autoFocus
                placeholder="e.g. Welcome new members and send rules"
                maxLength={100}
                {...field}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );

  const mobileFooterContent =
    step === 1 ? (
      <div className="flex items-center justify-between w-full gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={() => handleOpenChange(false)}
        >
          Cancel
        </Button>
        <Button
          type="button"
          onClick={handleNext}
          disabled={source === "discord" && !form.watch("type")}
        >
          Next
          <ArrowRightIcon className="ml-1.5 size-4" />
        </Button>
      </div>
    ) : (
      <div className="flex items-center justify-between w-full gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={handleBack}
          disabled={createMutation.isPending}
        >
          <ArrowLeftIcon className="mr-1.5 size-4" /> Back
        </Button>
        <LoadingButton
          type="submit"
          form="mobile-event-create-form"
          loading={createMutation.isPending}
        >
          Create {name.singular}
        </LoadingButton>
      </div>
    );

  if (isMobile) {
    return (
      <Drawer open={open} onOpenChange={handleOpenChange}>
        <DrawerTrigger asChild>{children}</DrawerTrigger>
        <DrawerContent className="h-[85dvh] max-h-[85dvh] flex flex-col">
          <DrawerHeader className="text-left px-5 pt-3 pb-2 border-b border-border/40 flex-none">
            <DrawerTitle>Create {name.singular}</DrawerTitle>
            <DrawerDescription>
              {step === 1
                ? texts.description
                : `Add details for this ${name.singular}.`}
            </DrawerDescription>
          </DrawerHeader>
          <div className="p-4 flex-1 min-h-0 overflow-hidden flex flex-col">
            <Form {...form}>
              <form
                id="mobile-event-create-form"
                onSubmit={form.handleSubmit(onSubmit)}
                className="flex-1 min-h-0 flex flex-col"
              >
                {step === 1 ? mobileStep1Content : mobileStep2Content}
              </form>
            </Form>
          </div>
          <DrawerFooter className="px-4 py-3 border-t border-border/40 bg-muted/20 flex-none">
            {mobileFooterContent}
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
    );
  }

  // Desktop single-screen view
  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent
        className={cn(
          "sm:max-w-2xl max-h-[90dvh] flex flex-col p-0 overflow-hidden",
          source === "webhook" ? "sm:h-auto" : "sm:h-[640px]"
        )}
      >
        <DialogHeader className="p-6 pb-3 border-b border-border/40 flex-none">
          <DialogTitle>Create {name.singular}</DialogTitle>
          <DialogDescription>{texts.description}</DialogDescription>
        </DialogHeader>
        <div className="px-6 py-4 flex-1 min-h-0 overflow-hidden flex flex-col">
          <Form {...form}>
            <form
              id="desktop-event-create-form"
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex-1 min-h-0 flex flex-col space-y-3 p-1"
            >
              {pickerFields}
              {descriptionField}
            </form>
          </Form>
        </div>
        <DialogFooter className="px-6 py-3 border-t border-border/40 bg-muted/20 flex-none flex items-center justify-between sm:justify-between w-full">
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
          >
            Cancel
          </Button>
          <LoadingButton
            type="submit"
            form="desktop-event-create-form"
            loading={createMutation.isPending}
            disabled={source === "discord" && !form.watch("type")}
          >
            Create {name.singular}
          </LoadingButton>
        </DialogFooter>
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
}: {
  icon: LucideIcon;
  title: string;
  description: string;
  selected?: boolean;
  onSelect?: () => void;
}) {
  const className = cn(
    "flex items-start gap-3 rounded-lg border p-2.5 text-left transition-colors w-full min-h-[64px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary",
    selected
      ? "border-primary bg-primary/5 ring-1 ring-inset ring-primary"
      : onSelect && "border-border/60 hover:border-border hover:bg-muted/40"
  );

  const content = (
    <>
      <div
        className={cn(
          "rounded-md w-8 h-8 flex items-center justify-center shrink-0 mt-0.5 transition-colors",
          selected
            ? "bg-primary text-primary-foreground"
            : "bg-muted text-foreground"
        )}
      >
        <Icon className="size-4" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium leading-5 truncate">{title}</div>
        <div className="text-xs text-muted-foreground mt-0.5 line-clamp-2 break-words">
          {description}
        </div>
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
    >
      {content}
    </button>
  );
}

function PickerNote({ children }: { children: ReactNode }) {
  return (
    <div className="text-sm text-muted-foreground rounded-lg bg-muted/50 px-3 py-2 leading-relaxed">
      {children}
    </div>
  );
}

const eventCategoryTabs = [
  { id: "all", label: "All", icon: LayersIcon },
  { id: "Messages", label: "Messages", icon: MessageSquareIcon },
  { id: "Members", label: "Members", icon: UsersIcon },
  { id: "Reactions", label: "Reactions", icon: SmileIcon },
  { id: "Moderation", label: "Moderation", icon: ScrollTextIcon },
  { id: "Voice", label: "Voice", icon: MicIcon },
];

function DiscordEventPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const [activeCategory, setActiveCategory] = useState("all");
  const [search, setSearch] = useState("");

  const isSearching = search.trim() !== "";

  const categories = useMemo(() => {
    const res: { name: string; types: typeof discordEventTypes }[] = [];
    for (const type of discordEventTypes) {
      if (
        !isSearching &&
        activeCategory !== "all" &&
        type.category !== activeCategory
      ) {
        continue;
      }
      if (
        isSearching &&
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
  }, [search, isSearching, activeCategory]);

  const selected = discordEventTypes.find((t) => t.value === value);

  return (
    <div className="flex-1 min-h-0 flex flex-col space-y-3">
      {/* Sleek inline filter pills */}
      <div className="flex items-center gap-1.5 overflow-x-auto pb-0.5 no-scrollbar [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none] flex-none">
        {eventCategoryTabs.map((tab) => {
          const { id, label, icon: Icon } = tab;
          const isActive = activeCategory === id;
          return (
            <button
              key={id}
              type="button"
              onClick={() => {
                setActiveCategory(id);
                setSearch("");
              }}
              className={cn(
                "inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors whitespace-nowrap shrink-0",
                isActive
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted/70 text-muted-foreground hover:bg-muted hover:text-foreground"
              )}
            >
              <Icon className="size-3.5" />
              <span>{label}</span>
            </button>
          );
        })}
      </div>

      <div className="relative flex-none">
        <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search events…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9 h-9"
        />
      </div>

      <ScrollArea className="flex-1 min-h-0 pr-2">
        <div className="space-y-4 p-1">
          {categories.length === 0 && (
            <div className="text-sm text-muted-foreground text-center py-8">
              No events match your search.
            </div>
          )}
          {categories.map((category) => (
            <div key={category.name} className="space-y-2">
              <div className="flex items-center justify-between text-xs font-medium text-muted-foreground px-0.5">
                <span>{category.name}</span>
                <span className="text-muted-foreground/70 text-[11px] font-normal">
                  {category.types.length}
                </span>
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
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>

      {selected?.note && (
        <div className="flex-none animate-in fade-in duration-150">
          <PickerNote>{selected.note}</PickerNote>
        </div>
      )}
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
  const [custom, setCustom] = useState(false);
  const isCustom = custom || !preset;
  const [customCron, setCustomCron] = useState("");
  const isValid = useMemo(() => !!getSchedulePreview(value), [value]);

  return (
    <div className="flex-1 min-h-0 flex flex-col space-y-3">
      <ScrollArea className="flex-1 min-h-0 pr-2">
        <div className="space-y-3 p-1">
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
        </div>
      </ScrollArea>
      <div className="flex-none">
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
