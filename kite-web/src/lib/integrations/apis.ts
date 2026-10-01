import { IntegrationApi } from "./api";
import discordApi from "./discord/api.json";

// The API of each integration with request blocks. Kept apart from the
// integrations, which pages load that never need Discord's large one.
export const integrationApis: Record<string, IntegrationApi> = {
  discord: discordApi,
};
