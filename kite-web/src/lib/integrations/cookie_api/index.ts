import { Integration } from "../types";

// cookie-api.com publishes no spec, so openapi.json is written by hand from
// its docs, for the endpoints Kite's blocks use.
export const cookieApi: Integration = {
  id: "cookie_api",
  name: "Cookie API",
  description: "QR codes, AI images and other tools from cookie-api.com.",
  base_url: "https://api.cookie-api.com",
  auth: {
    type: "header",
    name: "Authorization",
    label: "API key",
    help_url:
      "https://docs.cookie-api.com/en/docs/getting-started/faq/api-key/",
  },
  test_path: "/api/api-key",
};
