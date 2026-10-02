import { useCallback } from "react";
import { toast } from "sonner";
import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
  useMessagesImportMutation,
} from "@/lib/api/mutations";
import { useMarketplaceListingImportMutation } from "@/lib/api/marketplace";
import { useAppId } from "@/lib/hooks/params";
import { useMessages, useVariables } from "@/lib/hooks/api";
import { removeForeignReferences } from "@/lib/marketplace";
import { FlowData } from "@/lib/types/flow.gen";

function plural(count: number, word: string) {
  return `${count} ${word}${count === 1 ? "" : "s"}`;
}

// Imports the selected items of a listing into the current app with the
// regular import routes, so plan limits apply like for any other import.
// Message templates go first, so blocks from the same listing can be pointed
// at the imported copies.
export function useMarketplaceImport() {
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();

  const listingImportMutation = useMarketplaceListingImportMutation();
  const messagesImportMutation = useMessagesImportMutation(appId);
  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);

  const ready = !!variables && !!messages;
  const pending =
    listingImportMutation.isPending ||
    messagesImportMutation.isPending ||
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
      const messageIdMap = new Map<string, string>();

      let removed = 0;
      const sanitize = (flow: FlowData) => {
        const res = removeForeignReferences(
          flow,
          variableIds,
          messageIds,
          messageIdMap
        );
        removed += res.removed;
        return res.flow;
      };

      const items = res.data.items
        .map((item, i) => ({ item, i }))
        .filter(({ i }) => selected.has(i))
        .map(({ item }) => item);

      // Message templates first, their new IDs are needed for the flows.
      const messageItems = items.filter(
        (item) => item.type === "message" && item.message_data
      );
      if (messageItems.length) {
        const msgRes = await messagesImportMutation.mutateAsync({
          messages: messageItems.map((item) => ({
            name: item.name,
            description: item.description || null,
            data: item.message_data!,
            flow_sources: Object.fromEntries(
              Object.entries(item.message_flow_sources ?? {}).map(
                ([id, flow]) => [id, sanitize(flow)]
              )
            ),
          })),
        });
        if (!msgRes.success) {
          toast.error(
            `Failed to import message templates: ${msgRes.error.message} (${msgRes.error.code})`
          );
          return false;
        }

        // The response keeps the order of the request.
        msgRes.data.forEach((msg, i) => {
          const sourceId = messageItems[i]?.source_id;
          if (msg && sourceId) {
            messageIdMap.set(sourceId, msg.id);
          }
        });
        toast.success(
          `Imported ${plural(messageItems.length, "message template")}`
        );
      }

      const commands = [];
      const eventListeners = [];
      for (const item of items) {
        if (!item.flow_source || item.type === "message") continue;

        if (item.type === "command") {
          commands.push({
            flow_source: sanitize(item.flow_source),
            enabled: true,
          });
        } else {
          eventListeners.push({
            source: item.source || "discord",
            flow_source: sanitize(item.flow_source),
            enabled: true,
          });
        }
      }

      let ok = true;

      if (commands.length) {
        const cmdRes = await commandsImportMutation.mutateAsync({ commands });
        if (cmdRes.success) {
          toast.success(`Imported ${plural(commands.length, "command")}`);
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
            `Imported ${plural(eventListeners.length, "event listener")}`
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
      messagesImportMutation,
      commandsImportMutation,
      eventListenersImportMutation,
      variables,
      messages,
    ]
  );

  return { importListing, ready, pending };
}
