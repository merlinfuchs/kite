import { describe, expect, it } from "vitest";
import { decodeSnowflake } from "./snowflake";

describe("decodeSnowflake", () => {
  it("decodes the example from the Discord docs", () => {
    expect(decodeSnowflake("175928847299117063")).toEqual({
      timestamp: new Date("2016-04-30T11:18:25.796Z").getTime(),
      workerId: 1,
      processId: 0,
      increment: 7,
    });
  });

  it("rejects values that aren't IDs", () => {
    expect(decodeSnowflake("123")).toBeNull();
    expect(decodeSnowflake("abc12345678901234567")).toBeNull();
    expect(decodeSnowflake("99999999999999999999")).toBeNull();
  });
});
