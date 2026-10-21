import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSaveAndExit?: () => void;
  onDiscardAndExit: () => void;
  isSaving?: boolean;
}

export default function UnsavedChangesDialog({
  open,
  onOpenChange,
  onSaveAndExit,
  onDiscardAndExit,
  isSaving,
}: Props) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="rounded-lg w-[calc(100%-2rem)] sm:max-w-[480px]">
        <AlertDialogHeader>
          <AlertDialogTitle>Unsaved Changes</AlertDialogTitle>
          <AlertDialogDescription>
            You have unsaved changes. If you leave now, they will be lost.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="flex-col sm:flex-row gap-2 sm:gap-2">
          <AlertDialogCancel disabled={isSaving}>Cancel</AlertDialogCancel>
          <button
            type="button"
            className={cn(
              buttonVariants({ variant: "destructive" }),
              "w-full sm:w-auto"
            )}
            onClick={() => {
              onOpenChange(false);
              onDiscardAndExit();
            }}
            disabled={isSaving}
          >
            Leave without saving
          </button>
          {onSaveAndExit && (
            <button
              type="button"
              className={cn(
                buttonVariants({ variant: "default" }),
                "w-full sm:w-auto"
              )}
              onClick={() => {
                onSaveAndExit();
              }}
              disabled={isSaving}
            >
              {isSaving ? "Saving..." : "Save and exit"}
            </button>
          )}
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
