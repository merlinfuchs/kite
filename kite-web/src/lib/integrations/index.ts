import { cookieApi } from "./cookie_api";
import { discord } from "./discord";
import { erlc } from "./erlc";
import { roblox } from "./roblox";
import { topgg } from "./topgg";
import { Integration } from "./types";

export { needsCredential } from "./types";

export const integrations: Integration[] = [
  discord,
  roblox,
  cookieApi,
  erlc,
  topgg,
];

export function getIntegration(id: string) {
  return integrations.find((i) => i.id === id);
}
