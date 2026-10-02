import { Integration } from "../types";

// Roblox's public APIs need no credential, so it's on until the app turns it
// off.
export const roblox: Integration = {
  id: "roblox",
  name: "Roblox",
  description: "Look up Roblox users.",
  auth: { type: "none" },
  availability: "default",
};
