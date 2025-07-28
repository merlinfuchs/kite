import { z } from "zod";

export const userResultSchema = z.object({
  id: z.string().describe("The ID of the user"),
  username: z.string().describe("The username of the user"),
  discriminator: z.string().describe("The discriminator of the user"),
  display_name: z.string().describe("The display name of the user"),
  avatar_url: z.string().describe("The avatar URL of the user"),
});

export const memberResultSchema = userResultSchema.extend({
  nick: z.string().describe("The nickname of the member"),
  avatar_url: z.string().describe("The avatar URL of the member"),
});

export const channelResultSchema = z.object({
  id: z.string().describe("The ID of the channel"),
  name: z.string().describe("The name of the channel"),
  type: z
    .enum(["text", "voice", "category", "news", "store", "stage"])
    .describe("The type of the channel"),
});

export const guildResultSchema = z.object({
  id: z.string().describe("The ID of the guild"),
  name: z.string().describe("The name of the guild"),
  icon_url: z.string().describe("The icon URL of the guild"),
});

export const messageResultSchema = z.object({
  id: z.string().describe("The ID of the message"),
  channel_id: z
    .string()
    .describe("The ID of the channel the message was sent in"),
  content: z.string().describe("The content of the message"),
  author: userResultSchema.describe("The author of the message if it's a DM"),
  member: memberResultSchema
    .optional()
    .describe("The author of the message if it's in a server"),
});

export const roleResultSchema = z.object({
  id: z.string().describe("The ID of the role"),
  name: z.string().describe("The name of the role"),
  color: z.string().describe("The color of the role"),
  hoist: z.boolean().describe("Whether the role is hoisted"),
  mentionable: z.boolean().describe("Whether the role is mentionable"),
});

export const nodeActionResponseCreateResultSchema = messageResultSchema;

export const nodeActionResponseEditResultSchema = messageResultSchema;

export const nodeActionMessageCreateResultSchema = messageResultSchema;

export const nodeActionMessageEditResultSchema = messageResultSchema;

export const nodeActionPrivateMessageCreateResultSchema = messageResultSchema;

export const nodeActionMessageGetResultSchema = messageResultSchema;

export const nodeActionUserGetResultSchema = userResultSchema;

export const nodeActionMemberGetResultSchema = memberResultSchema;

export const nodeActionChannelGetResultSchema = channelResultSchema;

export const nodeActionGuildGetResultSchema = guildResultSchema;

export const nodeActionRoleGetResultSchema = roleResultSchema;

export const nodeActionRobloxUserGetResultSchema = z.object({
  description: z.string().describe("The description of the Roblox user"),
  created: z
    .string()
    .describe("The datetime when the Roblox account was created"),
  is_banned: z.boolean().describe("Whether the Roblox user is banned"),
  external_app_display_Name: z
    .string()
    .describe("The display name shown in external applications"),
  has_verified_badge: z
    .boolean()
    .describe("Whether the Roblox user has a verified badge"),
  id: z.number().describe("The ID of the Roblox user"),
  name: z.string().describe("The username of the Roblox user"),
  display_name: z.string().describe("The display name of the Roblox user"),
});
