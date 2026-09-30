import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordInviteCreate: BlockDefinition = {
  type: "action_invite_create",
  title: "Create invite",
  description: "Create an invite link for a channel",
  icon: "link",
  category: "Channels",
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "create_channel_invite",
    method: "POST",
    path: "/channels/{channel_id}/invites",
  },
  fields: [
    {
      name: "channel_target",
      in: "path",
      target: "channel_id",
      type: "snowflake",
      label: "Channel",
      description:
        "ID of the channel the invite leads to. Leave empty to use the channel the flow runs in.",
      fallback: "channel",
    },
    {
      name: "max_age",
      in: "body",
      type: "integer",
      label: "Expires After",
      description:
        "Seconds until the invite expires, up to 5184000 (60 days). 0 means never. Defaults to 1 day.",
      min: 0,
      max: 5184000,
    },
    {
      name: "max_uses",
      in: "body",
      type: "integer",
      label: "Max Uses",
      description:
        "How often the invite can be used, up to 100. 0 means unlimited.",
      min: 0,
      max: 100,
    },
    {
      name: "temporary",
      in: "body",
      type: "boolean",
      label: "Temporary Membership",
      description:
        "Whether members who joined with the invite are kicked when they go offline, unless they got a role.",
    },
    {
      name: "unique",
      in: "body",
      type: "boolean",
      label: "Always New",
      description:
        "Whether to always create a new invite instead of reusing a similar existing one.",
    },
  ],
  result: {
    schema: z
      .object({
        code: z
          .string()
          .describe(
            "The invite code. The link is https://discord.gg/ followed by the code."
          ),
        max_age: z.number().describe("Seconds until the invite expires"),
        max_uses: z.number().describe("How often the invite can be used"),
      })
      .describe("The created invite"),
  },
};
