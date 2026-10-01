import { ReactNode, useMemo, useState } from "react";
import { toast } from "sonner";
import { MailPlusIcon, SatelliteDishIcon, SlashSquareIcon } from "lucide-react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import { Checkbox } from "../ui/checkbox";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { RadioGroup, RadioGroupItem } from "../ui/radio-group";
import LoadingButton from "../common/LoadingButton";
import {
  useMarketplaceListingCreateMutation,
  useMarketplaceListingUpdateMutation,
} from "@/lib/api/marketplace";
import { useCommands, useEventListeners, useMessages } from "@/lib/hooks/api";
import { referencedMessageIds } from "@/lib/marketplace";
import { useAppId } from "@/lib/hooks/params";
import {
  MarketplaceListing,
  MarketplaceListingItemRequest,
} from "@/lib/types/wire.gen";

const maxItems = 25;

const emptyMessageFields = {
  name: "",
  description: null,
  source_id: "",
  message_data: undefined,
  message_flow_sources: {},
};

type Props = {
  // Edits this listing instead of creating a new one.
  listing?: MarketplaceListing;
  children?: ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
};

export default function MarketplacePublishDialog({
  listing,
  children,
  open: controlledOpen,
  onOpenChange,
}: Props) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
  const open = controlledOpen ?? uncontrolledOpen;
  const setOpen = onOpenChange ?? setUncontrolledOpen;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {children && <DialogTrigger asChild>{children}</DialogTrigger>}
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        {open && (
          <PublishForm listing={listing} onDone={() => setOpen(false)} />
        )}
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so the form resets every time it's opened.
function PublishForm({
  listing,
  onDone,
}: {
  listing?: MarketplaceListing;
  onDone: () => void;
}) {
  const appId = useAppId();
  const commands = useCommands();
  const eventListeners = useEventListeners();
  const messages = useMessages();

  const createMutation = useMarketplaceListingCreateMutation();
  const updateMutation = useMarketplaceListingUpdateMutation(listing?.id ?? "");

  const [name, setName] = useState(listing?.name ?? "");
  const [description, setDescription] = useState(listing?.description ?? "");
  const [keepContents, setKeepContents] = useState(!!listing);
  const [selectedCommands, setSelectedCommands] = useState<Set<string>>(
    new Set()
  );
  const [selectedListeners, setSelectedListeners] = useState<Set<string>>(
    new Set()
  );
  const [selectedMessages, setSelectedMessages] = useState<Set<string>>(
    new Set()
  );

  const appCommands = useMemo(
    () => commands?.flatMap((c) => (c ? [c] : [])) ?? [],
    [commands]
  );
  const appListeners = useMemo(
    () => eventListeners?.flatMap((l) => (l ? [l] : [])) ?? [],
    [eventListeners]
  );

  const appMessages = useMemo(
    () => messages?.flatMap((m) => (m ? [m] : [])) ?? [],
    [messages]
  );

  // Templates that selected blocks use but that aren't selected, they'd be
  // cleared on import.
  const missingMessages = useMemo(() => {
    const flows = [
      ...appCommands
        .filter((c) => selectedCommands.has(c.id))
        .map((c) => c.flow_source),
      ...appListeners
        .filter((l) => selectedListeners.has(l.id))
        .map((l) => l.flow_source),
      ...appMessages
        .filter((m) => selectedMessages.has(m.id))
        .flatMap((m) => Object.values(m.flow_sources ?? {})),
    ];
    const referenced = referencedMessageIds(flows);
    return appMessages.filter(
      (m) => referenced.has(m.id) && !selectedMessages.has(m.id)
    );
  }, [
    appCommands,
    appListeners,
    appMessages,
    selectedCommands,
    selectedListeners,
    selectedMessages,
  ]);

  const selectedCount =
    selectedCommands.size + selectedListeners.size + selectedMessages.size;
  const pending = createMutation.isPending || updateMutation.isPending;

  function toggle(
    set: Set<string>,
    setter: (s: Set<string>) => void,
    id: string,
    checked: boolean
  ) {
    const next = new Set(set);
    if (checked) next.add(id);
    else next.delete(id);
    setter(next);
  }

  function buildItems(): MarketplaceListingItemRequest[] {
    if (keepContents && listing) {
      return listing.items.flatMap((item): MarketplaceListingItemRequest[] => {
        if (item.type === "message") {
          return item.message_data
            ? [
                {
                  type: "message",
                  source: "",
                  flow_source: { nodes: [], edges: [] },
                  name: item.name,
                  description: item.description || null,
                  source_id: item.source_id ?? "",
                  message_data: item.message_data,
                  message_flow_sources: item.message_flow_sources ?? {},
                },
              ]
            : [];
        }
        return item.flow_source
          ? [
              {
                ...emptyMessageFields,
                type: item.type,
                source: item.source ?? "",
                flow_source: item.flow_source,
              },
            ]
          : [];
      });
    }

    return [
      ...appCommands
        .filter((c) => selectedCommands.has(c.id))
        .map((c) => ({
          ...emptyMessageFields,
          type: "command",
          source: "",
          flow_source: c.flow_source,
        })),
      ...appListeners
        .filter((l) => selectedListeners.has(l.id))
        .map((l) => ({
          ...emptyMessageFields,
          type: "event_listener",
          source: l.source,
          flow_source: l.flow_source,
        })),
      ...appMessages
        .filter((m) => selectedMessages.has(m.id))
        .map((m) => ({
          type: "message",
          source: "",
          flow_source: { nodes: [], edges: [] },
          name: m.name,
          description: m.description,
          source_id: m.id,
          message_data: m.data,
          message_flow_sources: m.flow_sources ?? {},
        })),
    ];
  }

  function onSubmit() {
    const items = buildItems();

    if (name.trim().length < 3) {
      toast.error("The name must be at least 3 characters");
      return;
    }
    if (description.trim().length < 10) {
      toast.error("The description must be at least 10 characters");
      return;
    }
    if (items.length === 0) {
      toast.error("Select at least one item");
      return;
    }
    if (items.length > maxItems) {
      toast.error(`A listing can hold up to ${maxItems} items`);
      return;
    }

    const req = {
      name: name.trim(),
      description: description.trim(),
      app_id: appId,
      items,
    };

    const onSuccess = (
      res: Awaited<ReturnType<typeof createMutation.mutateAsync>>
    ) => {
      if (!res.success) {
        toast.error(
          `Failed to ${listing ? "update" : "publish"} listing: ${
            res.error.message
          } (${res.error.code})`
        );
        return;
      }

      toast.success(
        res.data.status === "approved"
          ? listing
            ? "Listing updated"
            : "Listing published"
          : "Submitted, a moderator will review it before it goes public"
      );
      onDone();
    };

    if (listing) {
      updateMutation.mutate(req, { onSuccess });
    } else {
      createMutation.mutate(req, { onSuccess });
    }
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {listing ? "Edit listing" : "Publish to the marketplace"}
        </DialogTitle>
        <DialogDescription>
          Share commands, event listeners and message templates from this app.
          Pick more than one to publish them together as a module.
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-5 my-2">
        <div className="space-y-2">
          <Label htmlFor="listing-name">Name</Label>
          <Input
            id="listing-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={64}
            placeholder="Moderation Kit"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="listing-description">Description</Label>
          <Textarea
            id="listing-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            maxLength={2000}
            minRows={3}
            maxRows={8}
            placeholder="What does it do and how do people set it up?"
          />
        </div>

        {listing && (
          <RadioGroup
            value={keepContents ? "keep" : "pick"}
            onValueChange={(v) => setKeepContents(v === "keep")}
            className="space-y-1"
          >
            <div className="flex items-center gap-2">
              <RadioGroupItem value="keep" id="contents-keep" />
              <Label htmlFor="contents-keep">Keep the current contents</Label>
            </div>
            <div className="flex items-center gap-2">
              <RadioGroupItem value="pick" id="contents-pick" />
              <Label htmlFor="contents-pick">
                Replace them with items from this app
              </Label>
            </div>
          </RadioGroup>
        )}

        {!keepContents && (
          <div className="space-y-4">
            <ItemPicker
              title="Commands"
              icon={SlashSquareIcon}
              emptyText="This app has no commands."
              items={appCommands.map((c) => ({
                id: c.id,
                name: `/${c.name}`,
                description: c.description,
              }))}
              selected={selectedCommands}
              onToggle={(id, checked) =>
                toggle(selectedCommands, setSelectedCommands, id, checked)
              }
            />
            <ItemPicker
              title="Event Listeners"
              icon={SatelliteDishIcon}
              emptyText="This app has no event listeners."
              items={appListeners.map((l) => ({
                id: l.id,
                name: l.type,
                description: l.description,
              }))}
              selected={selectedListeners}
              onToggle={(id, checked) =>
                toggle(selectedListeners, setSelectedListeners, id, checked)
              }
            />
            <ItemPicker
              title="Message Templates"
              icon={MailPlusIcon}
              emptyText="This app has no message templates."
              items={appMessages.map((m) => ({
                id: m.id,
                name: m.name,
                description: m.description ?? "",
              }))}
              selected={selectedMessages}
              onToggle={(id, checked) =>
                toggle(selectedMessages, setSelectedMessages, id, checked)
              }
            />
            {missingMessages.length > 0 && (
              <Alert>
                <AlertTitle>Some blocks use message templates</AlertTitle>
                <AlertDescription className="text-muted-foreground space-y-2">
                  <p>
                    {missingMessages.map((m) => m.name).join(", ")}{" "}
                    {missingMessages.length === 1 ? "isn't" : "aren't"}{" "}
                    selected. Include them so the blocks keep working after an
                    import.
                  </p>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => {
                      const next = new Set(selectedMessages);
                      missingMessages.forEach((m) => next.add(m.id));
                      setSelectedMessages(next);
                    }}
                  >
                    Include {missingMessages.length === 1 ? "it" : "them"}
                  </Button>
                </AlertDescription>
              </Alert>
            )}
            <div className="text-sm text-muted-foreground">
              {selectedCount} of up to {maxItems} selected
              {selectedCount > 1 && ", this will be published as a module"}
            </div>
          </div>
        )}

        <Alert>
          <AlertTitle>Everything in these blocks becomes public</AlertTitle>
          <AlertDescription className="text-muted-foreground">
            Anyone can read the flows you publish, including URLs, headers and
            API keys in request blocks. Remove secrets before publishing.
            Attachments of message templates aren&apos;t included. Listings are
            reviewed by moderators, and every change is reviewed again before it
            goes public.
          </AlertDescription>
        </Alert>
      </div>

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <LoadingButton onClick={onSubmit} loading={pending}>
          {listing ? "Save changes" : "Publish"}
        </LoadingButton>
      </DialogFooter>
    </>
  );
}

function ItemPicker({
  title,
  icon: Icon,
  emptyText,
  items,
  selected,
  onToggle,
}: {
  title: string;
  icon: typeof SlashSquareIcon;
  emptyText: string;
  items: { id: string; name: string; description: string }[];
  selected: Set<string>;
  onToggle: (id: string, checked: boolean) => void;
}) {
  return (
    <div>
      <div className="font-medium mb-2">{title}</div>
      {items.length === 0 ? (
        <div className="text-sm text-muted-foreground">{emptyText}</div>
      ) : (
        <div className="flex flex-col gap-1 max-h-56 overflow-y-auto rounded-md border p-2">
          {items.map((item) => (
            <label
              key={item.id}
              className="flex items-start gap-3 rounded-sm px-2 py-1.5 hover:bg-muted/50 cursor-pointer"
            >
              <Checkbox
                className="mt-0.5"
                checked={selected.has(item.id)}
                onCheckedChange={(checked) =>
                  onToggle(item.id, checked === true)
                }
              />
              <div className="min-w-0">
                <div className="flex items-center gap-1.5 text-sm font-medium">
                  <Icon className="h-4 w-4 text-muted-foreground flex-none" />
                  <span className="truncate">{item.name}</span>
                </div>
                <div className="text-xs text-muted-foreground line-clamp-1">
                  {item.description}
                </div>
              </div>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}
