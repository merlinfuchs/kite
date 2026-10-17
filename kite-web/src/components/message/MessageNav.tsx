import { useCallback, useEffect } from "react";
import {
  ArrowLeftIcon,
  CheckIcon,
  RefreshCwIcon,
  SendIcon,
} from "lucide-react";
import ThemeSwitch from "../common/ThemeSwitch";
import MessageSendDialog from "./MessageSendDialog";

interface Props {
  hasUnsavedChanges: boolean;
  isSaving: boolean;
  onSave: (d: {}) => void;
  onExit: () => void;
}

export default function MessageNav({
  hasUnsavedChanges,
  isSaving,
  onSave,
  onExit,
}: Props) {
  const save = useCallback(() => {
    onSave({});
  }, [onSave]);

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
          aria-label="Back to App"
        >
          <ArrowLeftIcon className="h-5 w-5" />
          <div className="hidden sm:block">Back to App</div>
        </button>
        {isSaving ? (
          <div
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
            onClick={save}
          >
            <RefreshCwIcon className="h-5 w-5 animate-spin" />
            <div className="hidden sm:block">Saving Changes</div>
          </div>
        ) : hasUnsavedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
            onClick={save}
          >
            <div className="h-3 w-3 rounded-full bg-foreground/80"></div>
            <div>Save Changes</div>
          </button>
        ) : (
          <div className="flex space-x-2 text-foreground/80 items-center">
            <CheckIcon className="h-5 w-5" />
            <div className="hidden sm:block">No Unsaved Changes</div>
          </div>
        )}
        {hasUnsavedChanges ? (
          <div className="flex space-x-2 text-foreground/60 items-center">
            <SendIcon className="h-5 w-5" />
            <div className="hidden sm:block">Send Message</div>
          </div>
        ) : (
          <MessageSendDialog>
            <button
              className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
              aria-label="Send Message"
            >
              <SendIcon className="h-5 w-5" />
              <div className="hidden sm:block">Send Message</div>
            </button>
          </MessageSendDialog>
        )}
      </div>
      <ThemeSwitch />
    </div>
  );
}
