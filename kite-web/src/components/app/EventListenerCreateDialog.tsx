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
import { ReactNode, useId, useMemo, useState } from "react";
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
  discordEventType,
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
  BotIcon,
  LayersIcon,
  LucideIcon,
  MessageSquareIcon,
  MicIcon,
  PencilIcon,
  ScrollTextIcon,
  SearchIcon,
  SmileIcon,
  TagIcon,
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

const dialogDescriptions: Record<EventListenerSource, string> = {
  discord: "Pick the Discord event that runs the flow.",
  schedule: "Pick how often the flow runs.",
  webhook: "The flow runs whenever another service sends a request to its URL.",
};

export default function EventListenerCreateDialog({
  source,
  children,
}: {
  source: EventListenerSource;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  // On mobile the picker and the description are two steps, so the keyboard
  // doesn't cover the picker. Webhooks have nothing to pick.
  const [step, setStep] = useState<"pick" | "describe">("pick");
  const isMobile = useIsMobile();
  const formId = useId();

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

  const name = eventSourceName(source);
  const pickField =
    source === "discord" ? "type" : source === "schedule" ? "cron" : undefined;
  const mobileStep = pickField ? step : "describe";

  // Resets on open instead of on close, so the content doesn't change while
  // the dialog animates out.
  function handleOpenChange(nextOpen: boolean) {
    if (nextOpen) {
      setStep("pick");
      form.reset();
    }
    setOpen(nextOpen);
  }

  async function handleNext() {
    if (pickField && (await form.trigger(pickField))) {
      setStep("describe");
    }
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

  const picker =
    source === "discord" ? (
      <FormField
        control={form.control}
        name="type"
        rules={{ required: "Pick an event." }}
        render={({ field }) => (
          <FormItem className="flex-1 min-h-64 flex flex-col">
            <DiscordEventPicker value={field.value} onChange={field.onChange} />
            <FormMessage />
          </FormItem>
        )}
      />
    ) : source === "schedule" ? (
      <FormField
        control={form.control}
        name="cron"
        rules={{ required: "Pick or enter a schedule." }}
        render={({ field }) => (
          <FormItem className="flex-1 min-h-64 flex flex-col">
            <SchedulePicker value={field.value} onChange={field.onChange} />
            <FormMessage />
          </FormItem>
        )}
      />
    ) : (
      <WebhookPicker />
    );

  const descriptionField = (
    <FormField
      control={form.control}
      name="description"
      rules={{ required: "Enter a description." }}
      render={({ field }) => (
        <FormItem className="flex-none">
          <FormLabel>Description</FormLabel>
          <FormDescription>What will you use it for?</FormDescription>
          <FormControl>
            <Input
              type="text"
              maxLength={100}
              autoFocus={isMobile}
              {...field}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );

  if (isMobile) {
    // The picked values can't change while the describe step is shown.
    const [type, cron] = form.getValues(["type", "cron"]);
    const selectedEvent = discordEventType(type);
    const selectedPreset = schedulePresets.find((p) => p.cron === cron.trim());

    return (
      <Drawer open={open} onOpenChange={handleOpenChange}>
        <DrawerTrigger asChild>{children}</DrawerTrigger>
        <DrawerContent className="h-[85dvh] max-h-[85dvh] flex flex-col">
          <DrawerHeader className="text-left px-4 pt-3 pb-3 border-b flex-none">
            <DrawerTitle>Create {name.singular}</DrawerTitle>
            <DrawerDescription>{dialogDescriptions[source]}</DrawerDescription>
          </DrawerHeader>
          <Form {...form}>
            <form
              id={formId}
              onSubmit={
                // Enter in the search or cron input goes to the next step.
                mobileStep === "pick"
                  ? (e) => {
                      e.preventDefault();
                      handleNext();
                    }
                  : form.handleSubmit(onSubmit)
              }
              className="flex-1 min-h-0 flex flex-col p-4 overflow-y-auto"
            >
              {/* Stays mounted on the describe step, so Back keeps the
                  search, category and custom schedule. */}
              {pickField && (
                <div
                  className={cn(
                    "flex-1 min-h-64 flex flex-col animate-in fade-in duration-200",
                    mobileStep !== "pick" && "hidden"
                  )}
                >
                  {picker}
                </div>
              )}
              {mobileStep === "describe" && (
                <div className="space-y-5 p-1 animate-in fade-in slide-in-from-right-4 duration-200">
                  {source === "discord" && selectedEvent && (
                    <div className="space-y-2">
                      <OptionCard
                        icon={selectedEvent.icon}
                        title={selectedEvent.label}
                        description={selectedEvent.description}
                        selected
                      />
                      {selectedEvent.note && (
                        <PickerNote>{selectedEvent.note}</PickerNote>
                      )}
                    </div>
                  )}
                  {source === "schedule" && (
                    <OptionCard
                      icon={selectedPreset?.icon ?? PencilIcon}
                      title={selectedPreset?.label ?? "Custom"}
                      description={selectedPreset?.description ?? cron}
                      selected
                    />
                  )}
                  {source === "webhook" && <WebhookPicker />}
                  {descriptionField}
                </div>
              )}
            </form>
          </Form>
          <DrawerFooter className="flex-row justify-between px-4 py-3 border-t flex-none">
            {mobileStep === "describe" && pickField ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => setStep("pick")}
                disabled={createMutation.isPending}
              >
                <ArrowLeftIcon className="mr-1.5 size-4" />
                Back
              </Button>
            ) : (
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
              >
                Cancel
              </Button>
            )}
            {mobileStep === "pick" ? (
              <Button type="button" onClick={handleNext}>
                Next
                <ArrowRightIcon className="ml-1.5 size-4" />
              </Button>
            ) : (
              <LoadingButton
                type="submit"
                form={formId}
                loading={createMutation.isPending}
              >
                Create {name.singular}
              </LoadingButton>
            )}
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent
        className={cn(
          "sm:max-w-2xl max-h-[90dvh] flex flex-col gap-0 p-0 overflow-hidden",
          pickField && "sm:h-[640px]"
        )}
      >
        <DialogHeader className="p-6 pb-4 border-b flex-none">
          <DialogTitle>Create {name.singular}</DialogTitle>
          <DialogDescription>{dialogDescriptions[source]}</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            id={formId}
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex-1 min-h-0 flex flex-col gap-4 px-6 py-4 overflow-y-auto"
          >
            {picker}
            {descriptionField}
          </form>
        </Form>
        <DialogFooter className="px-6 py-3 border-t flex-none sm:justify-between">
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
          >
            Cancel
          </Button>
          <LoadingButton
            type="submit"
            form={formId}
            loading={createMutation.isPending}
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
    "flex items-start gap-3 rounded-lg border p-2.5 text-left transition-colors w-full min-h-16 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
    selected
      ? "border-primary bg-primary/5 ring-1 ring-inset ring-primary"
      : onSelect && "hover:bg-muted/50"
  );

  const content = (
    <>
      <div
        className={cn(
          "rounded-md size-8 flex items-center justify-center shrink-0 mt-0.5 transition-colors",
          selected ? "bg-primary text-primary-foreground" : "bg-muted"
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
    <div className="text-sm text-muted-foreground rounded-lg bg-muted/50 px-3 py-2">
      {children}
    </div>
  );
}

// The categories of discordEventTypes, in the order of the picker.
const eventCategories = Array.from(
  new Set(discordEventTypes.map((t) => t.category))
);

const eventCategoryIcons: Record<string, LucideIcon> = {
  Messages: MessageSquareIcon,
  Reactions: SmileIcon,
  Members: UsersIcon,
  Moderation: ScrollTextIcon,
  Voice: MicIcon,
  Bot: BotIcon,
};

function DiscordEventPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  // Searching shows all categories again and picking a category clears the
  // search, so a search never misses events of another category.
  const [search, setSearch] = useState("");
  // Undefined shows all categories.
  const [activeCategory, setActiveCategory] = useState<string>();

  function pickCategory(category?: string) {
    setActiveCategory(category);
    setSearch("");
  }

  const categories = useMemo(() => {
    const res: { name: string; types: typeof discordEventTypes }[] = [];
    for (const type of discordEventTypes) {
      if (activeCategory && type.category !== activeCategory) {
        continue;
      }
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
  }, [search, activeCategory]);

  const selected = discordEventType(value);

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-3">
      <div className="flex-none flex sm:flex-wrap gap-1 overflow-x-auto no-scrollbar p-0.5">
        <CategoryPill
          icon={LayersIcon}
          label="All"
          active={!activeCategory}
          onClick={() => pickCategory(undefined)}
        />
        {eventCategories.map((category) => (
          <CategoryPill
            key={category}
            icon={eventCategoryIcons[category] ?? TagIcon}
            label={category}
            active={activeCategory === category}
            onClick={() => pickCategory(category)}
          />
        ))}
      </div>

      <div className="relative flex-none">
        <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search events"
          aria-label="Search events"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setActiveCategory(undefined);
          }}
          className="pl-9"
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
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>

      {selected?.note && (
        <div className="flex-none">
          <PickerNote>{selected.note}</PickerNote>
        </div>
      )}
    </div>
  );
}

// CategoryPill is a smaller version of the category chips of the template
// list, so all categories fit in one row on desktop.
function CategoryPill({
  icon: Icon,
  label,
  active,
  onClick,
}: {
  icon: LucideIcon;
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <Button
      type="button"
      size="sm"
      variant={active ? "secondary" : "ghost"}
      className="h-8 rounded-full flex-none gap-1.5 px-2.5 text-xs"
      onClick={onClick}
      aria-pressed={active}
    >
      <Icon className="size-3.5" />
      {label}
    </Button>
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
    <div className="flex-1 min-h-0 flex flex-col gap-3">
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
