import { BlockDefinition } from "./types";

export const discordBotProfileEdit: BlockDefinition = {
  type: "action_bot_profile_edit",
  title: "Edit bot server profile",
  description: "Change the bot's name, avatar, banner and bio in a server",
  icon: "bot",
  category: "Bot",
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "update_my_guild_member",
    method: "PATCH",
    path: "/guilds/{guild_id}/members/@me",
  },
  fields: [
    {
      name: "guild_target",
      in: "path",
      target: "guild_id",
      type: "snowflake",
      label: "Server",
      description:
        "ID of the server to change the bot's profile in. Leave empty to use the server the flow runs in.",
      fallback: "guild",
    },
    {
      name: "nick",
      in: "body",
      type: "string",
      label: "Display Name",
      description:
        "Name of the bot in the server, up to 32 characters. Leave empty to keep the current one.",
      max_length: 32,
    },
    {
      name: "bio",
      in: "body",
      type: "string",
      label: "Bio",
      description:
        "About Me text on the bot's profile in the server, up to 190 characters. Leave empty to keep the current one.",
      max_length: 190,
    },
    {
      name: "avatar",
      in: "body",
      type: "image",
      label: "Avatar URL",
      description:
        "URL of a PNG, JPEG, GIF or WebP image, up to 4 MB, to use as the bot's avatar in the server. Leave empty to keep the current one.",
    },
    {
      name: "banner",
      in: "body",
      type: "image",
      label: "Banner URL",
      description:
        "URL of a PNG, JPEG, GIF or WebP image, up to 4 MB, to use as the bot's profile banner in the server. Leave empty to keep the current one.",
    },
  ],
};
