import { useComponentsV2Enabled, useDocument } from "@/lib/message/state";
import { cn } from "@/lib/utils";
import ConfirmDialog from "../common/ConfirmDialog";

export default function MessageComponentsV2Toggle() {
  const enabled = useComponentsV2Enabled();
  const setComponentsV2 = useDocument((state) => state.setComponentsV2);

  return (
    <ConfirmDialog
      title={
        enabled
          ? "Switch back to content and embeds?"
          : "Switch to components v2?"
      }
      description="The two formats can't share content, so this will clear the message."
      onConfirm={() => setComponentsV2(!enabled)}
    >
      <button className="flex bg-muted p-1 rounded-lg text-sm font-medium text-muted-foreground">
        <div
          className={cn(
            "py-1 px-3 rounded-md transition-colors",
            !enabled && "bg-background text-foreground shadow-sm"
          )}
        >
          Embeds
        </div>
        <div
          className={cn(
            "py-1 px-3 rounded-md transition-colors",
            enabled && "bg-background text-foreground shadow-sm"
          )}
        >
          Components V2
        </div>
      </button>
    </ConfirmDialog>
  );
}
