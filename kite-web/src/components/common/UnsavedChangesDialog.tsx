import {
  AlertDialog,
  AlertDialogAction,
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
      <AlertDialogContent className="sm:max-w-[480px]">
        <AlertDialogHeader>
          <AlertDialogTitle>Unsaved Changes</AlertDialogTitle>
          <AlertDialogDescription>
            Looks like you forgot to save and are trying to leave, do you want
            to do this? If you do not save, all work since last save will be
            lost.
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
