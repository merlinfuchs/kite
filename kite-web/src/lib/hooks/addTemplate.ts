import {
  useBundleInstallCreateMutation,
  useMessageInstanceCreateMutation,
} from "../api/mutations";
import { track } from "../analytics";
import { InstallSelection, installRequest } from "../flow/templateInstall";
import { Template } from "../flow/templates";
import { MessageData } from "../types/message.gen";
import { useAppId } from "./params";
import { toast } from "sonner";

// A message of a template that couldn't be posted, with Discord's reason.
export type UnpostedMessage = { name: string; error: string };

// useAddTemplate adds what's selected of a template to the app, then posts
// its messages, like the verification message, in the channels of their
// parameters. A message that can't be posted can be posted again later.
export function useAddTemplate() {
  const appId = useAppId();
  const installMutation = useBundleInstallCreateMutation(appId);
  const postMutation = useMessageInstanceCreateMutation(appId);

  // Returns the messages that couldn't be posted, or undefined if the
  // template couldn't be added, which a toast says.
  async function add({
    template,
    selection,
    values,
    customMessages = {},
    guildId,
    first,
    from,
  }: {
    template: Template;
    selection: InstallSelection;
    values: Record<string, string>;
    customMessages?: Record<string, MessageData>;
    // The server the channels of the messages are in.
    guildId: string | null;
    // Whether the app didn't have the template yet.
    first: boolean;
    // Where it's added from, like "setup", for analytics.
    from?: string;
  }): Promise<UnpostedMessage[] | undefined> {
    const res = await installMutation.mutateAsync(
      installRequest(template, selection, values, customMessages)
    );
    if (!res.success) {
      toast.error(
        `Failed to add ${template.name}: ${res.error.message} (${res.error.code})`
      );
      return undefined;
    }

    const { commands, eventListeners, messages, variables } = selection;
    const analytics = {
      source: "template",
      template: template.id,
      template_version: template.version,
    };
    track("template_added", {
      ...analytics,
      count:
        commands.length +
        eventListeners.length +
        messages.length +
        variables.length,
      first,
      ...(from && { from }),
    });
    if (commands.length) {
      track("command_created", { ...analytics, count: commands.length });
    }
    if (eventListeners.length) {
      track("event_listener_created", {
        ...analytics,
        count: eventListeners.length,
      });
    }

    const unposted: UnpostedMessage[] = [];
    for (const message of messages) {
      const channelId = message.sendTo && values[message.sendTo];
      const item = res.data.items.find(
        (i) => i.type === "message" && i.key === message.key
      );
      if (!channelId || !guildId || !item) continue;

      const postRes = await postMutation.mutateAsync({
        messageId: item.id,
        discord_guild_id: guildId,
        discord_channel_id: channelId,
      });
      if (postRes.success) {
        track("message_sent", analytics);
      } else {
        unposted.push({ name: message.name, error: postRes.error.message });
      }
    }
    return unposted;
  }

  return {
    add,
    pending: installMutation.isPending || postMutation.isPending,
  };
}
