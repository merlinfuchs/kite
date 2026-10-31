import { Integration } from "../types";

export const cookieApi: Integration = {
  id: "cookie_api",
  name: "Cookie API",
  description:
    "Transcripts, cards, QR codes, captchas and Minecraft verification from cookie-api.com.",
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
  data_shared:
    "Its blocks send Cookie API the settings you fill in. Create transcript also sends your bot's token, which Cookie API uses to read the channel's messages.",
  privacy_url: "https://www.cookie-api.com/privacy-policy",
  terms_url: "https://www.cookie-api.com/terms-of-service",
};
