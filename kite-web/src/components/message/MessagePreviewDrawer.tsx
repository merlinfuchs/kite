import { ViewIcon } from "lucide-react";
import { Button } from "../ui/button";
import { Drawer, DrawerContent, DrawerTrigger } from "../ui/drawer";
import MessageEditorPreview from "./MessageEditorPreview";

/**
 * Floating button that opens the message preview in a drawer on screens where the preview isn't shown next to the editor.
 * The editor needs `pb-24 xl:pb-8` so the button doesn't cover its last controls.
 */
export default function MessagePreviewDrawer() {
  return (
    <Drawer>
      <DrawerTrigger asChild>
        <Button
          size="icon"
          className="fixed bottom-5 right-5 xl:hidden"
          aria-label="Preview message"
        >
          <ViewIcon />
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <div className="max-h-[80dvh] overflow-x-hidden overflow-y-auto mt-3">
          <MessageEditorPreview reducePadding />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
