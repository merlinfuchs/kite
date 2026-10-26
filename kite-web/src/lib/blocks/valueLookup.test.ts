import { describe, expect, it } from "vitest";
import { parseLookupLines } from "./valueLookup";

describe("parseLookupLines", () => {
  it("splits key = value and tab separated lines", () => {
    expect(
      parseLookupLines(
        "hello = Hi there!\r\n\nbye=See you\nid\tyour id is {{user.id}}\tignored\nonly key"
      )
    ).toEqual([
      { key: "hello", value: "Hi there!" },
      { key: "bye", value: "See you" },
      { key: "id", value: "your id is {{user.id}}" },
      { key: "only key", value: "" },
    ]);
  });

  it("keeps = in values", () => {
    expect(parseLookupLines("a = b = c")).toEqual([
      { key: "a", value: "b = c" },
    ]);
  });
});
