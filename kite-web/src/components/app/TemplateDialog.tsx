import { LoaderCircleIcon } from "lucide-react";
import { useUserBundles } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import { Dialog, DialogContent, DialogTitle } from "../ui/dialog";
import { SourceAppCard } from "./BundleManageCards";
import { TemplateDetails } from "./TemplateDetails";
import { TemplateInstallCard } from "./TemplateInstallCard";
import { TemplateInstalledCard } from "./TemplateInstalledCard";

// TemplateDialog shows the page of a template with the card for adding it, so
// it can be added without leaving the editor.
export default function TemplateDialog({
  templateId,
  open,
  onOpenChange,
}: {
  templateId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl w-[95vw] max-h-[90dvh] overflow-y-auto">
        {open && <TemplateDialogContent templateId={templateId} />}
      </DialogContent>
    </Dialog>
  );
}

function TemplateDialogContent({ templateId }: { templateId: string }) {
  const appId = useAppId();
  const template = useTemplate(templateId);
  // The app a listed bundle of the user is made from can't add it.
  const own = useUserBundles()?.find((b) => b?.id === templateId);
  const fromHere = own?.source_app_id === appId;

  return (
    <>
      <DialogTitle className="sr-only">
        {template?.name ?? "Template"}
      </DialogTitle>
      {template ? (
        <TemplateDetails
          template={template}
          aside={
            own && fromHere ? (
              <SourceAppCard bundle={own} />
            ) : (
              <div className="space-y-5">
                <TemplateInstalledCard template={template} />
                <TemplateInstallCard template={template} />
              </div>
            )
          }
        />
      ) : template === null ? (
        <p className="text-muted-foreground text-sm">
          This template doesn&apos;t exist anymore.
        </p>
      ) : (
        <LoaderCircleIcon className="size-5 animate-spin text-muted-foreground mx-auto my-10" />
      )}
    </>
  );
}
