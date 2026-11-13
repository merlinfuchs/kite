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
import LoadingButton from "../common/LoadingButton";
import { ExportForm } from "./FlowExportDialog";
import { useBundleExportMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useMessages } from "@/lib/hooks/api";
import { parseExportedBundle } from "@/lib/flow/bundleExport";
import { BundleExportResponse } from "@/lib/types/wire.gen";
import { toast } from "sonner";
import {
  BundleItemPicker,
  selectedItems,
  useBundleItemSections,
} from "./BundleItemPicker";

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
  const sections = useBundleItemSections();
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

  function onExport() {
    if (!sections) return;
    exportMutation.mutate(selectedItems(sections, selected), {
      onSuccess: (res) => {
        if (res.success) setResult(res.data);
        else
          toast.error(
            `Failed to export: ${res.error.message} (${res.error.code})`
          );
      },
    });
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Export items</DialogTitle>
        <DialogDescription>
          Choose what to share as one code. Importing it makes copies in the
          other app, and channels and roles of your server are picked there.
        </DialogDescription>
      </DialogHeader>

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
