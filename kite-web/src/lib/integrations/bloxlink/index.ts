import { Integration } from "../types";

export const bloxlink: Integration = {
  id: "bloxlink",
  name: "Bloxlink",
  description:
    "Look up the Roblox accounts members verified with the Bloxlink bot.",
  base_url: "https://api.blox.link",
  auth: {
    type: "header",
    name: "Authorization",
    label: "server API key",
    help_url: "https://blox.link/dashboard",
  },
  availability: "opt_in",
  data_shared:
    "Its blocks send Bloxlink your server's API key and the Discord IDs of the members you look up.",
  privacy_url: "https://blox.link/legal/privacy-policy",
};
