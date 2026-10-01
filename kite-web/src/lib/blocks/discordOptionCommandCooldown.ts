import { z } from "zod";
import { numericRegex, placeholderRegex, templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

// Cooldowns are kept in memory and reset when Kite restarts, so they're kept
// short. Matches maxCooldownDuration in kite-service/pkg/flow/data.go.
export const maxCooldownDurationSeconds = 60 * 60;

const durationMessage =
  "Must be a whole number of seconds, or a single {{ }} placeholder";

export const discordOptionCommandCooldown: BlockDefinition = {
  type: "option_command_cooldown",
  title: "Command Cooldown",
  description:
    "Limit how often the command can be used, per user, per server, or globally, for up to 1 hour. Cooldowns reset when Kite restarts.",
  icon: "hourglass",
  category: "Commands",
  component: "option",
  requires: ["discord"],
  custom_label: false,
  strict_settings: true,
  fields: [
    {
      name: "cooldown_scope",
      schema: z
        .enum(["user", "server", "global"])
        .optional()
        .describe(
          "Who the cooldown applies to: the user who ran the command, everyone in the server, or everyone everywhere. Defaults to user."
        ),
    },
    {
      name: "cooldown_duration_seconds",
      schema: z
        .string()
        .regex(numericRegex, durationMessage)
        .refine(
          (v) => Number(v) >= 1 && Number(v) <= maxCooldownDurationSeconds,
          `Must be between 1 and ${maxCooldownDurationSeconds} seconds`
        )
        .or(z.string().regex(placeholderRegex, durationMessage))
        .describe(
          `How many whole seconds the cooldown lasts for, from 1 to ${maxCooldownDurationSeconds} (1 hour). Cooldowns reset when Kite restarts.`
        ),
    },
    {
      name: "cooldown_message",
      schema: templated(
        z.string().max(2000).optional(),
        "Message shown when someone uses the command while it's on cooldown. Use {{var('cooldown_remaining')}} to show how many seconds are left. Leave empty for a default message."
      ),
    },
  ],
  run: { kind: "custom" },
};
