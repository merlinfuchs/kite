import { CircleAlertIcon } from "lucide-react";
import { ReactNode } from "react";
import { Button } from "../ui/button";
import { DialogTrigger } from "../ui/dialog";

// The button of a block setting that is edited in a dialog, like a modal or a
// channel. Below it sums up the settings and how many have an error, as
// they're hidden until the dialog is opened.
export default function FlowDialogTrigger({
  label,
  summary,
  errorCount,
}: {
  label: string;
  summary: ReactNode;
  errorCount: number;
}) {
  return (
    <div className="space-y-2">
      <DialogTrigger asChild>
        <Button className="w-full">{label}</Button>
      </DialogTrigger>
      <div className="text-sm text-muted-foreground flex items-center gap-2">
        <div className="truncate flex-auto">{summary}</div>
        {errorCount > 0 && (
          <div className="flex items-center gap-1 text-red-600 dark:text-red-400 flex-none">
            <CircleAlertIcon className="h-4 w-4" />
            {errorCount} {errorCount === 1 ? "issue" : "issues"}
          </div>
        )}
      </div>
    </div>
  );
}
