import { describe, expect, it } from "vitest";
import {
  formatHttpTestBody,
  getHttpTestSecretNames,
  getHttpTestValueKeys,
} from "./httpTest";

describe("getHttpTestValueKeys", () => {
  it("finds placeholders in every field", () => {
    expect(
      getHttpTestValueKeys({
        url: "https://api.example.com/users/{{user.id}}",
        query: [{ key: "q", value: "{{arg('term')}}" }],
        headers: [{ key: "Authorization", value: 'Bearer {{var("token")}}' }],
        body: '{"n": {{result(\'count\').data().n}}, "g": "{{interaction.guild.id}}"}',
      })
    ).toEqual([
      "user.id",
      "arg('term')",
      "var('token')",
      "result('count')",
      "interaction.guild.id",
    ]);
  });

  it("maps nodes.x.result to result('x')", () => {
    expect(getHttpTestValueKeys({ url: "{{nodes.abc.result.id}}" })).toEqual([
      "result('abc')",
    ]);
  });

  it("skips members of calls, strings and method calls", () => {
    expect(
      getHttpTestValueKeys({
        url: "{{arg('user').id}} {{'a.b'}} {{user.name.upper()}} {{plain}}",
      })
    ).toEqual(["arg('user')", "user.name"]);
  });

  it("deduplicates keys", () => {
    expect(
      getHttpTestValueKeys({ url: "{{user.id}}/{{user.id}}/{{ user.id }}" })
    ).toEqual(["user.id"]);
  });

  it("leaves secrets to the app", () => {
    const request = {
      url: "https://api.example.com/?key={{secrets.API_KEY}}",
      headers: [{ key: "Authorization", value: "Bearer {{secrets.TOKEN}}" }],
      body: '{"id": "{{user.id}}", "k": "{{secrets.API_KEY}}"}',
    };
    expect(getHttpTestValueKeys(request)).toEqual(["user.id"]);
    expect(getHttpTestSecretNames(request)).toEqual(["API_KEY", "TOKEN"]);
  });

  it("handles a missing request", () => {
    expect(getHttpTestValueKeys(undefined)).toEqual([]);
  });
});

describe("formatHttpTestBody", () => {
  it("pretty-prints JSON", () => {
    expect(formatHttpTestBody('{"a":1}')).toEqual({
      text: '{\n  "a": 1\n}',
      json: true,
    });
  });

  it("leaves other text alone", () => {
    expect(formatHttpTestBody("{not json")).toEqual({
      text: "{not json",
      json: false,
    });
    expect(formatHttpTestBody("hello")).toEqual({ text: "hello", json: false });
  });
});
