import { describe, expect, it } from "vitest";
import spec from "./discordApi.json";
import { similarDiscordApiOperations } from "./discordApi";

describe("discord api", () => {
  it("matches the file embedded in the service", async () => {
    await expect(JSON.stringify(spec, null, 2) + "\n").toMatchFileSnapshot(
      "../../../../kite-service/pkg/flow/discord_api.json"
    );
  });

  it("suggests endpoints for names from the docs", () => {
    expect(similarDiscordApiOperations("modify_guild")).toContain(
      "update_guild"
    );
    expect(similarDiscordApiOperations("get_channel_messages")).toContain(
      "list_messages"
    );
  });
});
