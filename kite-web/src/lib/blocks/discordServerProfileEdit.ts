import { z } from "zod";
import { templated } from "../flow/dataSchema";
import {
  maxNameStyleColors,
  nameStyleEffects,
  nameStyleFonts,
} from "../nameStyle";
import { guildTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

const imageUrl = (what: string) =>
  templated(
    z.string(),
    `URL of a PNG, JPEG, GIF or WebP image up to 10 MB for the bot's ${what} in the server, like {{user.avatar_url}}. Empty resets it to the bot's own ${what}.`
  ).optional();

const nameStyle = z
  .object({
    font: z
      .enum(nameStyleFonts.map((f) => f.value))
      .describe("Font of the name, shown in servers and in the bot's profile."),
    effect: z
      .enum(nameStyleEffects.map((e) => e.value))
      .describe(
        "Effect of the name in the bot's profile. Solid, neon, toon and pop use one color, gradient two, gummy four and prism five."
      ),
    colors: z
      .array(
        templated(
          z.string(),
          "Hex color like #5865F2, or a number like 5793266."
        )
      )
      .min(1)
      .max(maxNameStyleColors)
      .describe(
        "Colors of the name in the bot's profile. Servers show the role color instead."
      ),
  })
  .optional()
  .describe(
    "Font, effect and colors of the bot's name. Reset profile removes it."
  );

export const discordServerProfileEdit: BlockDefinition = {
  type: "action_server_profile_edit",
  id_prefix: "profile",
  title: "Edit server profile",
  description:
    "Change the bot's nickname, avatar, banner, bio, or name style in a server",
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
          name_style: nameStyle,
          reset: z
            .boolean()
            .optional()
            .describe(
              "Reset the nickname, avatar, banner, about me and name style to the bot's own. Can't be combined with the other fields."
            ),
        })
        .refine((d) => Object.values(d).some((v) => v !== undefined), {
          message: "Set at least one thing to change",
        })
        .refine(
          (d) =>
            !d.reset ||
            [d.nick, d.avatar_url, d.banner_url, d.bio, d.name_style].every(
              (v) => v === undefined
            ),
          { message: "Reset can't be combined with other changes" }
        )
        // Validates the fields even before any was set, so the error shows up
        .prefault({})
        .describe(
          "The parts of the profile to change. Leave out the ones that stay as they are."
        ),
    },
  ],
  run: { kind: "custom" },
};
