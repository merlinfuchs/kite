import { Integration } from "../types";

export const erlc: Integration = {
  id: "erlc",
  name: "ER:LC",
  description:
    "Read the status of your Emergency Response: Liberty County private server and run commands in it.",
  base_url: "https://api.erlc.gg",
  auth: {
    type: "header",
    name: "server-key",
    label: "server key",
    help_url: "https://apidocs.erlc.gg/how-to-obtain-your-server-key",
    kite_key_header: "Authorization",
  },
  availability: "opt_in",
  test_path: "/v2/server",
  data_shared:
    "Its blocks send ER:LC your server key and the commands you run. To run commands, authorize Kite for your server once after enabling it.",
  authorize_url:
    "https://api.erlc.gg/server-owners/server/{credential_id}/authorize/{app_id}",
  rate_limit_headers: true,
};
