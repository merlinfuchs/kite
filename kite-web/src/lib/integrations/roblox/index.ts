import { Integration } from "../types";

// Roblox's public APIs need no credential and only see what the blocks look
// up, so like Discord there's nothing to turn on or off.
export const roblox: Integration = {
  id: "roblox",
  name: "Roblox",
  description: "Look up Roblox users.",
  auth: { type: "none" },
  availability: "always",
};
