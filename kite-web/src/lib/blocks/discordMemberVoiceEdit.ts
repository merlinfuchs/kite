import { AnyZodObject, z } from "zod";
import { guildTargetField, userTargetField } from "./fields";
import { BlockDefinition } from "./types";

// Empty leaves the setting unchanged.
const voiceStateChangeSchema = (description: string) =>
  z.enum(["true", "false"]).optional().describe(description);

const isChange = (value: unknown) => value === "true" || value === "false";

// A request without changes does nothing, so the editor says so.
function requireChange(schema: AnyZodObject) {
  return schema
    .refine(
      (data) =>
        isChange(data.member_voice_mute) ||
        isChange(data.member_voice_deaf) ||
        !!data.channel_target,
      "Set mute, deafen or a channel to move the member to"
    )
    .describe(
      "Set at least one of member_voice_mute, member_voice_deaf or channel_target."
    );
}

export const discordMemberVoiceEdit: BlockDefinition = {
  type: "action_member_voice_edit",
  title: "Edit voice state",
  description: "Server-mute, deafen or move a member to another voice channel",
  icon: "audio-lines",
  category: "Voice",
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "update_guild_member",
    method: "PATCH",
    path: "/guilds/{guild_id}/members/{user_id}",
    partial: true,
  },
  fields: [
    guildTargetField,
    userTargetField,
    {
      name: "member_voice_mute",
      in: "body",
      target: "mute",
      type: "boolean",
      schema: voiceStateChangeSchema(
        'Whether to server-mute the member, "true" or "false". Empty leaves it unchanged.'
      ),
    },
    {
      name: "member_voice_deaf",
      in: "body",
      target: "deaf",
      type: "boolean",
      schema: voiceStateChangeSchema(
        'Whether to server-deafen the member, "true" or "false". Empty leaves it unchanged.'
      ),
    },
    {
      name: "channel_target",
      in: "body",
      target: "channel_id",
      type: "snowflake",
      label: "Move to Channel",
      description:
        "ID of the voice channel to move the member to. Leave empty to keep the member where they are.",
    },
  ],
  refine: requireChange,
};
