import { describe, expect, it } from "vitest";
import { describeList } from "./list";

describe("describeList", () => {
  it("counts lines that aren't empty", () => {
    expect(describeList("", "item")).toBe("0 items");
    expect(describeList(" a \n\n b\r\n", "item")).toBe("2 items");
    expect(describeList("{{user.name}} is here", "value")).toBe("1 value");
  });

  it("can't count a placeholder that fills in the whole list", () => {
    expect(describeList(" {{result('words')}} ", "value")).toBe(
      "Filled in by the placeholder"
    );
  });
});
