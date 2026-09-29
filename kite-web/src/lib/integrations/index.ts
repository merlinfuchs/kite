import { discord } from "./discord";
import { Integration } from "./types";

export const integrations: Integration[] = [discord];

export function getIntegration(id: string) {
  return integrations.find((i) => i.id === id);
}
