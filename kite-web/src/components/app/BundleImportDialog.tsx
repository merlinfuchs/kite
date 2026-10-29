import { ReactNode, useState } from "react";
import { Dialog, DialogContent, DialogTrigger } from "../ui/dialog";
import { ShareCodeImportForm } from "./ShareCodeImportForm";
import { ExportedBundle, parseExportedBundle } from "@/lib/flow/bundleExport";
import { toast } from "sonner";
import dynamic from "next/dynamic";

const ImportItems = dynamic(() => import("./BundleImportItems"), {
  ssr: false,
});

export default function BundleImportDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      {/* Wider while the previews of the items show. */}
      <DialogContent className="max-w-lg has-[[data-preview]]:max-w-3xl max-h-[90vh] overflow-y-auto">
        <BundleImportForm onImported={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so its state resets every time it's opened.
function BundleImportForm({ onImported }: { onImported: () => void }) {
  const [bundle, setBundle] = useState<{
    content: unknown;
    parsed: ExportedBundle;
  } | null>(null);

  if (bundle) {
    return (
      <ImportItems
        content={bundle.content}
        bundle={bundle.parsed}
        onBack={() => setBundle(null)}
        onImported={onImported}
      />
    );
  }

  return (
    <ShareCodeImportForm
      title="Import items"
      jsonPlaceholder='{"commands": ...}'
      submitLabel="Continue"
      onData={(content, type) => {
        if (type !== null && type !== "bundle") {
          toast.error(
            "This code is for a single command, event listener or message template, import it on its page"
          );
          return;
        }
        const parsed = parseExportedBundle(content);
        if (!parsed) {
          toast.error("Invalid share code, make sure it was made with Export");
          return;
        }
        setBundle({ content, parsed });
      }}
    />
  );
}
