import { Integration } from "../types";

export const robloxOpenCloud: Integration = {
  id: "roblox_open_cloud",
  name: "Roblox Open Cloud",
  description:
    "Change the roles of your Roblox group's members and answer join requests.",
  base_url: "https://apis.roblox.com",
  // Ranking bots call Roblox's old group API with an account's cookie.
  other_hosts: ["groups.roblox.com"],
  auth: {
    type: "header",
    name: "x-api-key",
    label: "API key",
    help_url: "https://create.roblox.com/docs/cloud/auth/api-keys",
  },
  availability: "opt_in",
  data_shared:
    "Its blocks send Roblox your API key and the users and roles you change. The key can do what its owner can in the group, so give it only the group:read and group:write permissions.",
  privacy_url: "https://www.roblox.com/info/privacy",
  terms_url: "https://www.roblox.com/info/terms",
};
