import { getNodeValues, isKnownNodeType } from "@/lib/flow/nodes";
import { FlowData } from "@/lib/types/flow.gen";
import { MarketplaceListing, MarketplaceUser } from "@/lib/types/wire.gen";

export type BlockRisk = "destructive" | "external";

// Blocks that importers and moderators should look at twice. Destructive
// blocks remove or change things on the server, external blocks can send
// data to servers outside of Discord.
const riskyBlocks: Record<string, BlockRisk> = {
  action_member_ban: "destructive",
  action_member_kick: "destructive",
  action_member_timeout: "destructive",
  action_member_edit: "destructive",
  action_member_prune: "destructive",
  action_member_role_add: "destructive",
  action_member_role_remove: "destructive",
  action_channel_edit: "destructive",
  action_channel_delete: "destructive",
  action_role_delete: "destructive",
  action_message_delete: "destructive",
  action_message_reaction_clear: "destructive",
  action_emoji_delete: "destructive",
  action_sticker_delete: "destructive",
  action_soundboard_sound_delete: "destructive",
  action_guild_leave: "destructive",
  action_variable_delete: "destructive",
  action_http_request: "external",
  action_discord_api_request: "external",
};

export interface RiskyBlock {
  type: string;
  title: string;
  risk: BlockRisk;
}

export function getRiskyBlocks(blockTypes: string[]): RiskyBlock[] {
  return blockTypes
    .filter((type) => Object.hasOwn(riskyBlocks, type))
    .map((type) => ({
      type,
      title: getBlockTitle(type),
      risk: riskyBlocks[type],
    }));
}

export function getBlockTitle(type: string) {
  return isKnownNodeType(type) ? getNodeValues(type).defaultTitle : type;
}

export function listingKindLabel(kind: string) {
  switch (kind) {
    case "command":
      return "Command";
    case "event_listener":
      return "Event Listener";
    case "message":
      return "Message Template";
    default:
      return "Module";
  }
}

export function listingStatusLabel(status: string) {
  switch (status) {
    case "pending":
      return "In review";
    case "approved":
      return "Published";
    case "rejected":
      return "Rejected";
    case "removed":
      return "Removed";
    default:
      return status;
  }
}

export function listingSummary(
  listing: Pick<
    MarketplaceListing,
    "command_count" | "event_listener_count" | "message_count"
  >
) {
  const parts = [];
  if (listing.command_count) {
    parts.push(
      `${listing.command_count} command${
        listing.command_count === 1 ? "" : "s"
      }`
    );
  }
  if (listing.event_listener_count) {
    parts.push(
      `${listing.event_listener_count} event listener${
        listing.event_listener_count === 1 ? "" : "s"
      }`
    );
  }
  if (listing.message_count) {
    parts.push(
      `${listing.message_count} message template${
        listing.message_count === 1 ? "" : "s"
      }`
    );
  }
  return parts.join(", ");
}

export function userAvatarUrl(user: MarketplaceUser) {
  if (!user.discord_avatar) {
    return undefined;
  }
  return `https://cdn.discordapp.com/avatars/${user.discord_id}/${user.discord_avatar}.png?size=64`;
}

// Variable and message template IDs belong to the app the flow was exported
// from, so they are cleared when they don't exist in the current app.
// messageIdMap points templates that were imported together with the flow at
// their new IDs.
export function removeForeignReferences(
  flow: FlowData,
  variableIds: Set<string>,
  messageIds: Set<string>,
  messageIdMap?: Map<string, string>
) {
  let removed = 0;

  const nodes = (flow.nodes ?? []).map((node) => {
    const data = { ...node.data };
    let changed = false;

    if (data.variable_id && !variableIds.has(data.variable_id)) {
      delete data.variable_id;
      changed = true;
    }
    const mappedMessageId =
      data.message_template_id && messageIdMap?.get(data.message_template_id);
    if (mappedMessageId) {
      data.message_template_id = mappedMessageId;
    } else if (
      data.message_template_id &&
      !messageIds.has(data.message_template_id)
    ) {
      delete data.message_template_id;
      changed = true;
    }

    if (changed) removed++;
    return { ...node, data };
  });

  return { flow: { ...flow, nodes }, removed };
}

// Message template IDs that the flows reference, used to suggest adding the
// templates when publishing.
export function referencedMessageIds(flows: FlowData[]) {
  const res = new Set<string>();
  for (const flow of flows) {
    for (const node of flow.nodes ?? []) {
      if (node.data?.message_template_id) {
        res.add(node.data.message_template_id);
      }
    }
  }
  return res;
}

// Listings with more than one item are modules, like on the server.
export function listingKind(
  listing: Pick<
    MarketplaceListing,
    "command_count" | "event_listener_count" | "message_count"
  >
) {
  const total =
    listing.command_count +
    listing.event_listener_count +
    listing.message_count;
  if (total > 1) return "module";
  if (listing.command_count === 1) return "command";
  if (listing.message_count === 1) return "message";
  return "event_listener";
}
