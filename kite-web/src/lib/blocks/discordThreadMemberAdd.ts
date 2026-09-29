import { nodeActionThreadMemberAddDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordThreadMemberAdd: BlockDefinition = {
  type: "action_thread_member_add",
  title: "Add member to thread",
  description: "Add a member to a thread",
  icon: "user-plus",
  category: "Channels",
  credits: 1,
  schema: nodeActionThreadMemberAddDataSchema,
  inputs: ["channel_target", "user_target", "audit_log_reason", "custom_label"],
  run: {
    kind: "request",
    integration: "discord",
    operation: "add_thread_member",
    method: "PUT",
    path: "/channels/{channel_id}/thread-members/{user_id}",
  },
  // Only describe the request. The settings keep their schema and inputs.
  fields: [
    {
      name: "channel_target",
      in: "path",
      target: "channel_id",
      type: "snowflake",
      label: "Channel",
      description: "ID of the channel.",
    },
    {
      name: "user_target",
      in: "path",
      target: "user_id",
      type: "snowflake",
      label: "User",
      description: "ID of the user.",
    },
  ],
};
