import { cookieApi } from "./cookie_api";
import { discord } from "./discord";
import { roblox } from "./roblox";
import { Integration } from "./types";

export { needsCredential } from "./types";

export const integrations: Integration[] = [discord, roblox, cookieApi];

export function getIntegration(id: string) {
  return integrations.find((i) => i.id === id);
}
