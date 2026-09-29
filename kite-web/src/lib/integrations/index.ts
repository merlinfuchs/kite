import { discord } from "./discord";
import { roblox } from "./roblox";
import { Integration } from "./types";

export const integrations: Integration[] = [discord, roblox];

export function getIntegration(id: string) {
  return integrations.find((i) => i.id === id);
}
