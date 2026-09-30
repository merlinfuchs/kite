import { describe, expect, it } from "vitest";
import spec from "./discordApi.json";
import { similarDiscordApiOperations } from "./discordApi";
import { nodeActionDiscordApiRequestDataSchema } from "./dataSchema";

function issues(data: Record<string, unknown>) {
  const res = nodeActionDiscordApiRequestDataSchema.safeParse({
    discord_api_request_data: data,
  });
  return res.error?.issues.map((i) => `${i.path.join(".")}: ${i.message}`);
}

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

  it("validates requests against the endpoint", () => {
    expect(
      issues({
        operation: "list_messages",
        path_params: [
          { key: "channel_id", value: "{{channel.id}}" },
          { key: "guild_id", value: "1" },
        ],
        query: [{ key: "limit", value: "5" }],
      })
    ).toBeUndefined();
    expect(
      issues({
        operation: "bulk_update_guild_roles",
        path_params: [{ key: "guild_id", value: "1" }],
        body_json: [{ id: "2", position: 1 }],
      })
    ).toBeUndefined();

    expect(
      issues({
        operation: "list_messages",
        path_params: [{ key: "channel_id", value: "abc" }],
        query: [
          { key: "limit", value: "five" },
          { key: "foo", value: "1" },
        ],
      })
    ).toEqual([
      "discord_api_request_data.path_params.channel_id: Must be a number or ID, or a single {{ }} placeholder",
      "discord_api_request_data.query.limit: Must be a whole number, or a single {{ }} placeholder",
      "discord_api_request_data.query.foo: The endpoint has no parameter foo",
    ]);
    expect(issues({ operation: "get_channel" })).toEqual([
      "discord_api_request_data.path_params.channel_id: channel_id is required",
    ]);
  });
});
