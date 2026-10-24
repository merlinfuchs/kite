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
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import {
  useBundleCreateMutation,
  useBundleVersionCreateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import { UserBundle } from "@/lib/types/wire.gen";
import {
  BundleItemPicker,
  selectedItems,
  useBundleItemSections,
} from "./BundleItemPicker";

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
  return <BundleForm bundle={bundle} itemKeys={keys} onDone={onDone} />;
}

function BundleForm({
  bundle,
  itemKeys = [],
  onDone,
}: {
  bundle?: UserBundle;
  itemKeys?: string[];
  onDone: () => void;
}) {
  const appId = useAppId();
  const sections = useBundleItemSections();
  const createMutation = useBundleCreateMutation(appId);
  const publishMutation = useBundleVersionCreateMutation(appId);

  const [name, setName] = useState(bundle?.name ?? "");
  const [description, setDescription] = useState(bundle?.description ?? "");
  // A new version starts with the items of the current one.
  const [selected, setSelected] = useState(() => new Set(itemKeys));

  const items = sections ? selectedItems(sections, selected) : undefined;
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
      ...items,
    };
    const res = bundle
      ? await publishMutation.mutateAsync({ bundleId: bundle.id, ...req })
      : await createMutation.mutateAsync(req);
    if (!res.success) {
      toast.error(
        `Failed to ${bundle ? "publish" : "create"} the bundle: ${
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
            : "Create bundle"}
        </DialogTitle>
        <DialogDescription>
          {bundle
            ? "Apps that added the bundle can update to it. Items you leave out stay in those apps without updates."
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
      </div>

      {sections && (
        <BundleItemPicker
          sections={sections}
          selected={selected}
          onChange={setSelected}
        />
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
