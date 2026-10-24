import { describe, expect, it } from "vitest";
import { needsValue, syncParameters } from "./parameters";

describe("parameters", () => {
  it("adds the ones blocks use as required", () => {
    const nodes = [{ data: { channel_target: "{{params.log_channel}}" } }];
    const synced = syncParameters({ nodes } as any, [
      { key: "reason", type: "text", value: "" },
    ]);
    expect(synced).toEqual([
      { key: "reason", type: "text", value: "" },
      { key: "log_channel", type: "text", required: true, value: "" },
    ]);
  });

  it("needs a value when required", () => {
    expect(needsValue({ key: "a", type: "channel", value: "" })).toBe(false);
    expect(
      needsValue({ key: "a", type: "text", required: true, value: " " })
    ).toBe(true);
    expect(
      needsValue({ key: "a", type: "role", required: true, value: "1" })
    ).toBe(false);
  });
});
