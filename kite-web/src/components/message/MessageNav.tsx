import { ReactNode, useCallback, useEffect } from "react";
import {
  ArrowLeftIcon,
  CheckIcon,
  RefreshCwIcon,
  SendIcon,
} from "lucide-react";
import ThemeSwitch from "../common/ThemeSwitch";
import MessageSendDialog from "./MessageSendDialog";
import { BundleInstallNavItem } from "../app/BundleInstallBadge";

interface Props {
  hasUnsavedChanges: boolean;
  isSaving: boolean;
  bundleInstallId?: string | null;
  onSave: (d: {}) => void;
  onExit: () => void;
  // The save history button.
  history?: ReactNode;
}

export default function MessageNav({
  hasUnsavedChanges,
  isSaving,
  bundleInstallId,
  onSave,
  onExit,
  history,
}: Props) {
  const save = useCallback(() => {
    if (!isSaving) onSave({});
  }, [onSave, isSaving]);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        save();
      }
    }

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onSave, save]);

  return (
    <div className="h-12 flex items-center justify-between px-4 select-none bg-muted/50">
      <div className="flex items-center space-x-4 sm:space-x-8">
        <button
          className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
          onClick={onExit}
          aria-label="Back to app"
        >
          <ArrowLeftIcon className="h-5 w-5" />
          <div className="hidden sm:block">Back to app</div>
        </button>
        {isSaving ? (
          <div
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
            onClick={save}
          >
            <RefreshCwIcon className="h-5 w-5 animate-spin" />
            <div className="hidden sm:block">Saving changes</div>
          </div>
        ) : hasUnsavedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
            onClick={save}
          >
            <div className="h-3 w-3 rounded-full bg-foreground/80"></div>
            <div>Save changes</div>
          </button>
        ) : (
          <div className="flex space-x-2 text-foreground/80 items-center">
            <CheckIcon className="h-5 w-5" />
            <div className="hidden sm:block">No unsaved changes</div>
          </div>
        )}
        {hasUnsavedChanges ? (
          <div className="flex space-x-2 text-foreground/60 items-center">
            <SendIcon className="h-5 w-5" />
            <div className="hidden sm:block">Send message</div>
          </div>
        ) : (
          <MessageSendDialog>
            <button
              className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
              aria-label="Send message"
            >
              <SendIcon className="h-5 w-5" />
              <div className="hidden sm:block">Send message</div>
            </button>
          </MessageSendDialog>
        )}
        <BundleInstallNavItem bundleInstallId={bundleInstallId} />
      </div>
      <div className="flex items-center space-x-4 sm:space-x-6">
        {history}
        <ThemeSwitch />
      </div>
    </div>
  );
}
