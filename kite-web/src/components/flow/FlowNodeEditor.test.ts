import { describe, expect, it } from "vitest";
import { nodeTypes } from "@/lib/flow/nodes";
import { intputs } from "./FlowNodeEditor";

describe("block editor", () => {
  // Blocks name their inputs in fields, and a misspelled one would be hidden.
  it("has an input for every setting", () => {
    for (const [type, values] of Object.entries(nodeTypes)) {
      for (const input of values.dataFields) {
        if (input.startsWith("field:")) continue;
        expect(intputs[input], `${type} ${input}`).toBeDefined();
      }
    }
  });
});
