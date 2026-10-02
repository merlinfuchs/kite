import { Integration } from "../types";

// The app's bot token is its credential, and every app uses Discord.
export const discord: Integration = {
  id: "discord",
  name: "Discord",
  description: "Everything your bot does in Discord.",
  auth: { type: "discord_bot" },
  availability: "always",
};
