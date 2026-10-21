import { discord } from "./discord";
import { erlc } from "./erlc";
import { roblox } from "./roblox";
import { Integration } from "./types";

export { needsCredential } from "./types";

export const integrations: Integration[] = [discord, roblox, erlc];

export function getIntegration(id: string) {
  return integrations.find((i) => i.id === id);
}
