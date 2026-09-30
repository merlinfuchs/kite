import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { channelTargetField, userTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordThreadMemberRemove: BlockDefinition = {
  type: "action_thread_member_remove",
  title: "Remove member from thread",
  description: "Remove a member from a thread",
  icon: "user-minus",
  category: "Channels",
  credits: 1,
  run: {
    kind: "request",
    integration: "discord",
    operation: "delete_thread_member",
    method: "DELETE",
    path: "/channels/{channel_id}/thread-members/{user_id}",
  },
  fields: [
    {
      ...channelTargetField,
      schema: numericOrPlaceholder("ID of the thread."),
    },
    { ...userTargetField, schema: templated(z.string(), "ID of the user.") },
  ],
};
