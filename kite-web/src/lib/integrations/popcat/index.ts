import { Integration } from "../types";

// Popcat needs no key, but it's an unaffiliated hobby project with no stated
// uptime or rate limits, and its blocks send it what users enter, so apps
// turn it on. See design/integrations.md.
export const popcat: Integration = {
  id: "popcat",
  name: "Popcat",
  description:
    "Jokes, facts, text effects, lookups like the weather or a GitHub user, and meme and welcome card images.",
  base_url: "https://api.popcat.xyz",
  auth: { type: "none" },
  availability: "opt_in",
  data_shared:
    "Its blocks send Popcat the text and image URLs you enter in them. Images are made by Popcat when Discord loads them.",
};
