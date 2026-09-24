import { CodeIcon, PaintbrushIcon } from "lucide-react";
import { useDocument } from "@/lib/message/state";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import MessageControlsButton from "./MessageControlsButton";
import MessageControlsUndo from "./MessageControlsUndo";
import MessageJSONDialog from "./MessageJSONDialog";

export default function MessageControls() {
  const clearMessage = useDocument((state) => state.clear);

  return (
    <div className="flex items-center justify-between space-x-3">
      <div className="flex items-center space-x-3">
        <MessageControlsUndo />
      </div>
      <div className="flex items-center space-x-3">
        <MessageJSONDialog>
          <MessageControlsButton
            icon={CodeIcon}
            label="JSON Code"
            onClick={() => {}}
          />
        </MessageJSONDialog>
        <ConfirmDialog
          title="Are you sure that you want to clear the message?"
          description="This will clear everything and cannot be undone."
          onConfirm={clearMessage}
        >
          <MessageControlsButton
            icon={PaintbrushIcon}
            label="Clear Message"
            onClick={() => {}}
          />
        </ConfirmDialog>
      </div>
    </div>
  );
}
