import { ReactNode, useState } from "react";
import { toast } from "sonner";
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
import { Switch } from "../ui/switch";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import {
  useBundleCreateMutation,
  useBundleVersionCreateMutation,
} from "@/lib/api/mutations";
import { useCommands, useEventListeners, useMessages } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import { referencedMessages } from "@/lib/flow/messageTemplates";
import { referencedParameters } from "@/lib/flow/parameters";
import { parameterDefaults } from "@/lib/flow/templates";
import { FlowData } from "@/lib/types/flow.gen";
import { BundleExportRequest, UserBundle } from "@/lib/types/wire.gen";
import {
  BundleItemPicker,
  selectedItems,
  useBundleItemSections,
} from "./BundleItemPicker";
import { isPickedParameter } from "./TemplateParameterFields";

// BundleDialog makes a bundle from items of the app, or publishes them as the
// new version of the bundle.
export default function BundleDialog({
  bundle,
  children,
}: {
  bundle?: UserBundle;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        {/* Only mounted while open, so its state resets every time. */}
        {open &&
          (bundle ? (
            <BundleVersionForm bundle={bundle} onDone={() => setOpen(false)} />
          ) : (
            <BundleForm onDone={() => setOpen(false)} />
          ))}
      </DialogContent>
    </Dialog>
  );
}

// BundleVersionForm starts the new version with the items of the current
// one, whose keys are their IDs in the app.
function BundleVersionForm({
  bundle,
  onDone,
}: {
  bundle: UserBundle;
  onDone: () => void;
}) {
  const template = useTemplate(bundle.id);
  if (!template) return null;

  // Messages that commands or event listeners use are added with them, so
  // they're only selected if they were chosen on their own.
  const used = new Set(
    [...template.commands, ...template.eventListeners].flatMap((i) =>
      i.flowSource.nodes.flatMap((n) => {
        const id = (n.data as { message_template_id?: string } | undefined)
          ?.message_template_id;
        return id ? [id] : [];
      })
    )
  );
  const keys = [
    ...template.commands,
    ...template.eventListeners,
    ...template.messages.filter((m) => !used.has(m.key)),
  ].map((i) => i.key);
  return (
    <BundleForm
      bundle={bundle}
      itemKeys={keys}
      defaults={parameterDefaults(template.parameters)}
      onDone={onDone}
    />
  );
}

function BundleForm({
  bundle,
  itemKeys = [],
  defaults = {},
  onDone,
}: {
  bundle?: UserBundle;
  itemKeys?: string[];
  // The defaults the current version shares.
  defaults?: Record<string, string>;
  onDone: () => void;
}) {
  const appId = useAppId();
  const sections = useBundleItemSections();
  const createMutation = useBundleCreateMutation(appId);
  const publishMutation = useBundleVersionCreateMutation(appId);

  const [name, setName] = useState(bundle?.name ?? "");
  const [description, setDescription] = useState(bundle?.description ?? "");
  const [notes, setNotes] = useState("");
  // A new version starts with the items of the current one.
  const [selected, setSelected] = useState(() => new Set(itemKeys));
  // Whether the owner switched sharing the value of a parameter on or off.
  const [toggled, setToggled] = useState<Record<string, boolean>>({});

  const items = sections ? selectedItems(sections, selected) : undefined;
  const shareable = useShareableParameters(items);
  // A value the current version shares is shared again, unless it changed
  // since. A changed one could be private now.
  const isShared = (p: ShareableParameter) =>
    toggled[p.key] ?? defaults[p.key] === p.value;
  const count = items
    ? items.command_ids.length +
      items.event_listener_ids.length +
      items.message_ids.length
    : 0;

  async function save() {
    if (!items) return;
    const req = {
      name: name.trim(),
      description: description.trim(),
      shared_parameter_keys: shareable.flatMap((p) =>
        isShared(p) ? [p.key] : []
      ),
      ...items,
    };
    const res = bundle
      ? await publishMutation.mutateAsync({
          bundleId: bundle.id,
          ...req,
          notes: notes.trim(),
        })
      : await createMutation.mutateAsync(req);
    if (!res.success) {
      toast.error(
        `Failed to ${bundle ? "publish" : "create"} the template: ${
          res.error.message
        } (${res.error.code})`
      );
      return;
    }
    toast.success(
      bundle
        ? `Published version ${res.data.version} of ${res.data.name}`
        : `Created ${res.data.name}`
    );
    onDone();
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {bundle
            ? `Publish version ${bundle.version + 1} of ${bundle.name}`
            : "Create template"}
        </DialogTitle>
        <DialogDescription>
          {bundle
            ? "Apps that added the template can update to it. Items you leave out stay in those apps without updates."
            : "Choose what's part of it. The message templates and stored variables they use are added. Apps that add it get your new versions as updates."}
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-4">
        <div className="space-y-1.5">
          <Label htmlFor="bundle-name">Name</Label>
          <Input
            id="bundle-name"
            value={name}
            maxLength={100}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="bundle-description">Description</Label>
          <Textarea
            id="bundle-description"
            value={description}
            maxLength={1000}
            placeholder="What it does"
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        {bundle && (
          <div className="space-y-1.5">
            <Label htmlFor="bundle-notes">What changed</Label>
            <Textarea
              id="bundle-notes"
              value={notes}
              maxLength={1000}
              placeholder="Shown to apps that update to this version"
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>
        )}
      </div>

      {sections && (
        <BundleItemPicker
          sections={sections}
          selected={selected}
          onChange={setSelected}
        />
      )}

      {!!shareable.length && (
        <div className="min-w-0">
          <div className="text-sm font-medium">Default values</div>
          <div className="text-xs text-muted-foreground mt-0.5 mb-2.5">
            Apps that add the template start with the values you share, and can
            change them. Don&apos;t share anything private, like an API key.
          </div>
          <div className="space-y-2.5">
            {shareable.map((p) => (
              <label
                key={p.key}
                className="flex items-center gap-2.5 text-sm cursor-pointer"
              >
                <Switch
                  checked={isShared(p)}
                  onCheckedChange={(checked) =>
                    setToggled({ ...toggled, [p.key]: checked })
                  }
                />
                <span className="truncate">{p.label}</span>
                <span className="text-xs text-muted-foreground truncate ml-auto max-w-[50%]">
                  {p.value}
                </span>
              </label>
            ))}
          </div>
        </div>
      )}

      <DialogFooter className="gap-2">
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <LoadingButton
          onClick={save}
          loading={createMutation.isPending || publishMutation.isPending}
          disabled={!name.trim() || !count}
        >
          {bundle ? "Publish" : "Create"} with {count}{" "}
          {count === 1 ? "item" : "items"}
        </LoadingButton>
      </DialogFooter>
    </>
  );
}

type ShareableParameter = { key: string; label: string; value: string };

// useShareableParameters are the text parameters of the selected items that
// have a value, which can be shared as defaults. The flows of the message
// templates the items use are part of the template too. A channel or role is
// from the owner's server, so it's never shared.
function useShareableParameters(
  items: BundleExportRequest | undefined
): ShareableParameter[] {
  const commands = useCommands();
  const eventListeners = useEventListeners();
  const messages = useMessages();
  if (!items || !commands || !eventListeners || !messages) return [];

  const selectedFlows: FlowData[] = [
    ...commands.flatMap((c) =>
      c && items.command_ids.includes(c.id) ? [c.flow_source] : []
    ),
    ...eventListeners.flatMap((l) =>
      l && items.event_listener_ids.includes(l.id) ? [l.flow_source] : []
    ),
    ...messages.flatMap((m) =>
      m && items.message_ids.includes(m.id) ? Object.values(m.flow_sources) : []
    ),
  ];
  const flows = [
    ...selectedFlows,
    ...referencedMessages(selectedFlows, messages, items.message_ids).flatMap(
      (m) => Object.values(m.flow_sources)
    ),
  ];

  const res: ShareableParameter[] = [];
  for (const flow of flows) {
    const used = referencedParameters(flow);
    for (const p of flow.parameters ?? []) {
      if (
        used.includes(p.key) &&
        !isPickedParameter(p) &&
        p.value.trim() &&
        !res.some((r) => r.key === p.key)
      ) {
        res.push({ key: p.key, label: p.label || p.key, value: p.value });
      }
    }
  }
  return res;
}
