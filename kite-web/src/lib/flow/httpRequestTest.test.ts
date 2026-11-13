import { describe, expect, it } from "vitest";
import { formatResponseBody, httpRequestPlaceholders } from "./httpRequestTest";

describe("httpRequestPlaceholders", () => {
  it("lists each placeholder once, without the ones that are a secret", () => {
    expect(
      httpRequestPlaceholders({
        url: "https://api.example.com/users/{{ user.id }}?key={{secrets.KEY}}",
        method: "POST",
        headers: [{ key: "Authorization", value: "Bearer {{secrets.TOKEN}}" }],
        query: [
          { key: "server", value: "{{server.id}}" },
          { key: "count", value: "{{webhook.data.secrets.count}}" },
        ],
        body_json: {
          name: "{{user.name}}",
          "{{user.id}}": ["{{result('get').body.count}}", 1],
        },
      })
    ).toEqual([
      "user.id",
      "server.id",
      "webhook.data.secrets.count",
      "user.name",
      "result('get').body.count",
    ]);
  });
});

describe("formatResponseBody", () => {
  it("indents JSON and keeps other text", () => {
    expect(formatResponseBody('{"a":1}')).toBe('{\n  "a": 1\n}');
    expect(formatResponseBody("plain text")).toBe("plain text");
  });
});
