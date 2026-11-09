import { PermissionFlag } from "../discord/permissions";
import { BlockDefinition } from "./types";

export const discordSoundboardSoundPlay: BlockDefinition = {
  type: "action_soundboard_sound_play",
  id_prefix: "soundboard",
  title: "Play soundboard sound",
  description: "Bot plays a soundboard sound in its voice channel",
  icon: "audio-lines",
  category: "Voice",
  credits: 1,
  permissions: (data) => {
    const perms: PermissionFlag[] = ["SPEAK", "USE_SOUNDBOARD"];
    if (data?.source_guild_id) {
      perms.push("USE_EXTERNAL_SOUNDS");
    }
    return perms;
  },
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "send_soundboard_sound",
    method: "POST",
    path: "/channels/{channel_id}/send-soundboard-sound",
  },
  fields: [
    {
      name: "channel_target",
      in: "path",
      target: "channel_id",
      type: "snowflake",
      label: "Voice Channel",
      description:
        "ID of the voice channel to play the sound in. The bot has to be in it, use Join voice channel first. Leave empty to use the channel the flow runs in, like a command used in a voice channel's chat.",
      fallback: "channel",
    },
    {
      name: "sound_id",
      in: "body",
      type: "snowflake",
      label: "Sound",
      description:
        "ID of the soundboard sound to play, of this server or one of Discord's default sounds.",
      required: true,
    },
    {
      name: "source_guild_id",
      in: "body",
      type: "snowflake",
      label: "Sound From Server",
      description:
        "ID of the server the sound is from, only for sounds of another server the bot is in. Leave empty for sounds of this server and default sounds.",
    },
  ],
};
