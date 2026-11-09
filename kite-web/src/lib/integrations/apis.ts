import { IntegrationApi } from "./api";
import cookieApi from "./cookie_api/api.json";
import discordApi from "./discord/api.json";
import erlcApi from "./erlc/api.json";

// The API of each integration with request blocks. Kept apart from the
// integrations, which pages load that never need Discord's large one.
export const integrationApis: Record<string, IntegrationApi> = {
  discord: discordApi,
  cookie_api: cookieApi,
  erlc: erlcApi,
};
