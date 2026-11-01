import { Integration } from "../types";

export const topgg: Integration = {
  id: "topgg",
  name: "Top.gg",
  description:
    "Check who voted for your bot or server on Top.gg, read its listing and post its server count.",
  base_url: "https://top.gg/api/v1",
  auth: {
    type: "header",
    name: "Authorization",
    prefix: "Bearer ",
    label: "API token",
    help_url: "https://docs.top.gg/authentication",
  },
  availability: "opt_in",
  test_path: "/projects/@me",
  data_shared:
    "Its blocks send Top.gg your API token, the Discord IDs of the users you check votes of and the server count you post.",
  privacy_url: "https://top.gg/privacy",
  terms_url: "https://top.gg/terms",
};
