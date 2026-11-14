import { BlockDefinition } from "./types";

export const discordInviteDelete: BlockDefinition = {
  type: "action_invite_delete",
  id_prefix: "invite",
  title: "Delete invite",
  description: "Delete an invite link",
  icon: "unlink",
  category: "Channels",
  permissions: ["MANAGE_CHANNELS"],
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "invite_revoke",
  },
  fields: [
    {
      name: "invite_code",
      target: "code",
      label: "Invite Code",
      description:
        "The invite code to delete, for example the part after discord.gg/.",
    },
  ],
};
