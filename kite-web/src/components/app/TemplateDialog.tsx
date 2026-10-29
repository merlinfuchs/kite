import { Template } from "@/lib/flow/templates";
import { Dialog, DialogContent, DialogTitle } from "../ui/dialog";
import AppTemplateAside from "./AppTemplateAside";
import { TemplateDetails } from "./TemplateDetails";

// TemplateDialog shows the page of a template with the cards for adding it,
// so it can be added without leaving the editor.
export default function TemplateDialog({
  template,
  from,
  open,
  onOpenChange,
}: {
  template: Template;
  // Where the dialog was opened from, for analytics.
  from: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl w-[95vw] max-h-[90dvh] overflow-y-auto">
        <DialogTitle className="sr-only">{template.name}</DialogTitle>
        <TemplateDetails
          template={template}
          aside={<AppTemplateAside template={template} from={from} />}
        />
      </DialogContent>
    </Dialog>
  );
}
