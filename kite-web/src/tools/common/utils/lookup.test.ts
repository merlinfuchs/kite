import { describe, expect, it } from "vitest";
import { parseLookupId } from "./lookup";

describe("parseLookupId", () => {
  it.each([
    ["80351110224678912", "80351110224678912"],
    ["<@80351110224678912>", "80351110224678912"],
    ["<@!80351110224678912>", "80351110224678912"],
    ["https://discord.com/users/80351110224678912", "80351110224678912"],
    [
      "https://discord.com/oauth2/authorize?client_id=235148962103951360&scope=bot",
      "235148962103951360",
    ],
    ["123456789012345678901", null],
    [
      "https://discord.com/channels/613425648685547541/613430047285706767/1083038298513887292",
      null,
    ],
    ["hello", null],
  ])("%s", (value, id) => {
    expect(parseLookupId(value)).toBe(id);
  });
});
