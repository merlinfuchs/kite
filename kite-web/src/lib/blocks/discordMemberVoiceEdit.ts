import { AnyZodObject, z } from "zod";
import { numericOrPlaceholder } from "../flow/dataSchema";
import { guildTargetSetting, userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

const voiceStateChangeSchema = (description: string) =>
  z.enum(["unchanged", "on", "off"]).optional().describe(description);

const isChange = (value: unknown) => value === "on" || value === "off";

// The service fails a block that changes nothing, so the editor says so first.
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
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [
    guildTargetSetting,
    userTargetSetting,
    {
      name: "member_voice_mute",
      schema: voiceStateChangeSchema(
        "Whether to server-mute the member. Empty means unchanged."
      ),
    },
    {
      name: "member_voice_deaf",
      schema: voiceStateChangeSchema(
        "Whether to server-deafen the member. Empty means unchanged."
      ),
    },
    {
      name: "channel_target",
      input: "member_voice_channel",
      schema: numericOrPlaceholder(
        "ID of the voice channel to move the member to. Empty leaves the member where they are."
      ).optional(),
    },
  ],
  refine: requireChange,
  audit_log_reason: true,
  run: { kind: "custom" },
};
