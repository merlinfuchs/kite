import { describe, expect, it } from "vitest";
import { formatResponseBody, httpRequestPlaceholders } from "./httpRequestTest";

describe("httpRequestPlaceholders", () => {
  it("lists each placeholder once, without secrets", () => {
    expect(
      httpRequestPlaceholders({
        url: "https://api.example.com/users/{{ user.id }}?key={{secrets.KEY}}",
        method: "POST",
        headers: [{ key: "Authorization", value: "Bearer {{secrets.TOKEN}}" }],
        query: [{ key: "server", value: "{{server.id}}" }],
        body_json: {
          name: "{{user.name}}",
          "{{user.id}}": ["{{result('get').body.count}}", 1],
        },
      })
    ).toEqual([
      "user.id",
      "server.id",
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
