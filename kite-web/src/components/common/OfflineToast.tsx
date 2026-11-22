import { onlineManager } from "@tanstack/react-query";
import { useEffect } from "react";
import { toast } from "sonner";

// OfflineToast shows while the browser is offline. Requests wait until it's
// back online, so a save made meanwhile would otherwise seem to hang.
export default function OfflineToast() {
  useEffect(
    () =>
      onlineManager.subscribe((online) => {
        if (online) {
          toast.dismiss("offline");
          return;
        }
        toast.info(
          "You're offline. Changes you save are sent once you're back online.",
          { id: "offline", duration: Infinity }
        );
      }),
    []
  );
  return null;
}
