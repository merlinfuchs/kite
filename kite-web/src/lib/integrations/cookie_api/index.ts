import { Integration } from "../types";

export const cookieApi: Integration = {
  id: "cookie_api",
  name: "Cookie API",
  description: "QR codes and other tools from cookie-api.com.",
  base_url: "https://api.cookie-api.com",
  auth: {
    type: "header",
    name: "Authorization",
    label: "API key",
    help_url:
      "https://docs.cookie-api.com/en/docs/getting-started/faq/api-key/",
  },
  availability: "opt_in",
  test_path: "/api/time/current-time",
};
