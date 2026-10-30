import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { guildTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

const imageUrl = (what: string) =>
  templated(
    z.string(),
    `URL of a PNG, JPEG, GIF or WebP image up to 10 MB for the bot's ${what} in the server, like {{user.avatar_url}}. Empty resets it to the bot's own ${what}.`
  ).optional();

export const discordServerProfileEdit: BlockDefinition = {
  type: "action_server_profile_edit",
  id_prefix: "profile",
  title: "Edit server profile",
  description: "Change the bot's nickname, avatar, banner or bio in a server",
  icon: "id-card",
  category: "Bot",
  requires: ["discord"],
  premium_feature: "server_profiles",
  credits: 1,
  strict_settings: true,
  fields: [
    guildTargetSetting,
    {
      name: "server_profile_data",
      schema: z
        .object({
          nick: templated(
            z.string().max(32),
            "Nickname of the bot in the server. Empty resets it to the bot's name."
          ).optional(),
          avatar_url: imageUrl("avatar"),
          banner_url: imageUrl("banner"),
          bio: templated(
            z.string().max(190),
            "About me of the bot in the server. Empty resets it to the bot's own."
          ).optional(),
        })
        .refine((d) => Object.values(d).some((v) => v !== undefined), {
          message: "Set at least one thing to change",
        })
        // Validates the fields even before any was set, so the error shows up
        .prefault({})
        .describe(
          "The parts of the profile to change. Leave out the ones that stay as they are."
        ),
    },
  ],
  run: { kind: "custom" },
};
