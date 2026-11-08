import { bloxlink } from "./bloxlink";
import { cookieApi } from "./cookie_api";
import { discord } from "./discord";
import { erlc } from "./erlc";
import { roblox } from "./roblox";
import { robloxOpenCloud } from "./roblox_open_cloud";
import { Integration } from "./types";

export { needsCredential } from "./types";

export const integrations: Integration[] = [
  discord,
  roblox,
  robloxOpenCloud,
  bloxlink,
  cookieApi,
  erlc,
];

export function getIntegration(id: string) {
  return integrations.find((i) => i.id === id);
}
