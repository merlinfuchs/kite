import { useCallback } from "react";
import { toast } from "sonner";
import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
} from "@/lib/api/mutations";
import { useMarketplaceListingImportMutation } from "@/lib/api/marketplace";
import { useAppId } from "@/lib/hooks/params";
import { useMessages, useVariables } from "@/lib/hooks/api";
import { removeForeignReferences } from "@/lib/marketplace";

// Imports the selected items of a listing into the current app with the
// regular import routes, so plan limits apply like for any other import.
export function useMarketplaceImport() {
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();

  const listingImportMutation = useMarketplaceListingImportMutation();
  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);

  const ready = !!variables && !!messages;
  const pending =
    listingImportMutation.isPending ||
    commandsImportMutation.isPending ||
    eventListenersImportMutation.isPending;

  const importListing = useCallback(
    async (listingId: string, selected: Set<number>) => {
      const res = await listingImportMutation.mutateAsync(listingId);
      if (!res.success) {
        toast.error(
          `Failed to import listing: ${res.error.message} (${res.error.code})`
        );
        return false;
      }

      const variableIds = new Set(variables?.flatMap((v) => (v ? [v.id] : [])));
      const messageIds = new Set(messages?.flatMap((m) => (m ? [m.id] : [])));

      let removed = 0;
      const commands = [];
      const eventListeners = [];

      for (const [i, item] of res.data.items.entries()) {
        if (!selected.has(i) || !item.flow_source) continue;

        const sanitized = removeForeignReferences(
          item.flow_source,
          variableIds,
          messageIds
        );
        removed += sanitized.removed;

        if (item.type === "command") {
          commands.push({ flow_source: sanitized.flow, enabled: true });
        } else {
          eventListeners.push({
            source: item.source || "discord",
            flow_source: sanitized.flow,
            enabled: true,
          });
        }
      }

      let ok = true;

      if (commands.length) {
        const cmdRes = await commandsImportMutation.mutateAsync({ commands });
        if (cmdRes.success) {
          toast.success(
            `Imported ${commands.length} command${
              commands.length === 1 ? "" : "s"
            }`
          );
        } else {
          ok = false;
          toast.error(
            `Failed to import commands: ${cmdRes.error.message} (${cmdRes.error.code})`
          );
        }
      }

      if (eventListeners.length) {
        const elRes = await eventListenersImportMutation.mutateAsync({
          event_listeners: eventListeners,
        });
        if (elRes.success) {
          toast.success(
            `Imported ${eventListeners.length} event listener${
              eventListeners.length === 1 ? "" : "s"
            }`
          );
        } else {
          ok = false;
          toast.error(
            `Failed to import event listeners: ${elRes.error.message} (${elRes.error.code})`
          );
        }
      }

      if (ok && removed > 0) {
        toast.warning(
          `${removed} block(s) referenced variables or message templates from another app, reselect them in the editor.`
        );
      }
      if (ok && commands.length) {
        toast.info("Deploy your commands to make them show up in Discord.");
      }

      return ok;
    },
    [
      listingImportMutation,
      commandsImportMutation,
      eventListenersImportMutation,
      variables,
      messages,
    ]
  );

  return { importListing, ready, pending };
}
