import { ReactNode, useState } from "react";
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
import { Checkbox } from "../ui/checkbox";
import LoadingButton from "../common/LoadingButton";
import { ExportForm } from "./FlowExportDialog";
import { useBundleExportMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useCommands, useEventListeners, useMessages } from "@/lib/hooks/api";
import { parseExportedBundle } from "@/lib/flow/bundleExport";
import { eventListenerLabel } from "@/lib/flow/eventTypes";
import { BundleExportResponse } from "@/lib/types/wire.gen";
import { toast } from "sonner";

export default function BundleExportDialog({
  children,
}: {
  children: ReactNode;
}) {
  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <BundleExportForm />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so its state resets every time it's opened.
function BundleExportForm() {
  const appId = useAppId();
  const commands = useCommands();
  const eventListeners = useEventListeners();
  const messages = useMessages();
  const exportMutation = useBundleExportMutation(appId);

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [result, setResult] = useState<BundleExportResponse | null>(null);

  if (result) {
    const bundle = parseExportedBundle(result.content);
    const parameters = bundle?.parameters ?? [];
    const selectedMessages = (messages ?? []).filter(
      (m) => m && selected.has(m.id)
    ).length;
    const added =
      (bundle?.messages.length ?? 0) -
      selectedMessages +
      (bundle?.variables.length ?? 0);
    const exported = selected.size;

    return (
      <ExportForm
        title={`Export ${exported} ${exported === 1 ? "item" : "items"}`}
        type="bundle"
        shareData={result.content ?? {}}
        flows={[]}
        note={
          <div className="space-y-2 text-sm text-muted-foreground">
            {added > 0 && (
              <p>
                The message templates and stored variables they use are
                included.
              </p>
            )}
            {!!parameters.length && (
              <p>
                Whoever imports it fills in their own{" "}
                {new Intl.ListFormat("en").format(
                  parameters.map((p) => p.label)
                )}
                .
              </p>
            )}
          </div>
        }
      />
    );
  }

  function toggle(id: string, checked: boolean) {
    const next = new Set(selected);
    if (checked) next.add(id);
    else next.delete(id);
    setSelected(next);
  }

  function onExport() {
    const pickIds = (items?: ({ id: string } | undefined)[]) =>
      (items ?? []).flatMap((i) => (i && selected.has(i.id) ? [i.id] : []));

    exportMutation.mutate(
      {
        command_ids: pickIds(commands),
        event_listener_ids: pickIds(eventListeners),
        message_ids: pickIds(messages),
      },
      {
        onSuccess: (res) => {
          if (res.success) setResult(res.data);
          else
            toast.error(
              `Failed to export: ${res.error.message} (${res.error.code})`
            );
        },
      }
    );
  }

  const sections = [
    {
      title: "Commands",
      items: (commands ?? []).flatMap((c) =>
        c ? [{ id: c.id, label: `/${c.name}`, description: c.description }] : []
      ),
    },
    {
      title: "Event listeners",
      items: (eventListeners ?? []).flatMap((l) =>
        l
          ? [
              {
                id: l.id,
                label: eventListenerLabel(l.source, l.type),
                description: l.description,
              },
            ]
          : []
      ),
    },
    {
      title: "Message templates",
      items: (messages ?? []).flatMap((m) =>
        m ? [{ id: m.id, label: m.name, description: m.description ?? "" }] : []
      ),
    },
  ].filter((s) => s.items.length);

  return (
    <>
      <DialogHeader>
        <DialogTitle>Export items</DialogTitle>
        <DialogDescription>
          Choose what to share as one code. Importing it makes copies in the
          other app, and channels and roles of your server are picked there.
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-5">
        {sections.map((section) => (
          <div key={section.title}>
            <div className="text-sm font-medium mb-2">{section.title}</div>
            <div className="space-y-2.5">
              {section.items.map((item) => (
                <label
                  key={item.id}
                  className="flex items-center gap-2.5 text-sm cursor-pointer"
                >
                  <Checkbox
                    checked={selected.has(item.id)}
                    onCheckedChange={(checked) => toggle(item.id, !!checked)}
                  />
                  <span className="truncate">{item.label}</span>
                  <span className="text-xs text-muted-foreground truncate ml-auto max-w-[50%]">
                    {item.description}
                  </span>
                </label>
              ))}
            </div>
          </div>
        ))}
        {commands && eventListeners && messages && !sections.length && (
          <p className="text-sm text-muted-foreground">
            Your app has nothing to export yet.
          </p>
        )}
      </div>

      <DialogFooter className="gap-2">
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        {selected.size ? (
          <LoadingButton onClick={onExport} loading={exportMutation.isPending}>
            Export {selected.size} {selected.size === 1 ? "item" : "items"}
          </LoadingButton>
        ) : (
          <Button disabled>Nothing selected</Button>
        )}
      </DialogFooter>
    </>
  );
}
