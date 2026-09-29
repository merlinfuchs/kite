import { flowChannelField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageBulkDelete: BlockDefinition = {
  type: "action_message_bulk_delete",
  title: "Bulk delete messages",
  description: "Delete 2 to 100 messages of a channel at once",
  icon: "trash-2",
  category: "Messages",
  credits: 1,
  audit_log_reason: true,
  // Destructive, but Discord only deletes up to 100 messages that are less
  // than 2 weeks old, and only the ones given by ID.
  run: {
    kind: "request",
    integration: "discord",
    operation: "bulk_delete_messages",
    method: "POST",
    path: "/channels/{channel_id}/messages/bulk-delete",
  },
  fields: [
    flowChannelField,
    {
      name: "message_ids",
      in: "body",
      target: "messages",
      type: "snowflake_list",
      label: "Messages",
      description:
        "2 to 100 message IDs separated by commas, or a placeholder with a list of messages like the result of List channel messages. Messages older than 2 weeks can't be deleted.",
      required: true,
      min: 2,
      max: 100,
    },
  ],
};
