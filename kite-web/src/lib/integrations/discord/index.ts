import { Integration } from "../types";
import api from "./api.json";

// The app's bot token is its credential, so Discord is always connected.
export const discord: Integration = {
  id: "discord",
  name: "Discord",
  description: "Everything your bot does in Discord.",
  auth: { type: "discord_bot" },
  api,
};
