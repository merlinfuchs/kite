import { describe, expect, it } from "vitest";
import { nodeTypes } from "@/lib/flow/nodes";
import { dialogFields, settingInputs } from "./FlowNodeEditor";

describe("block editor", () => {
  // Blocks name their inputs in fields, and a misspelled one would be hidden.
  it("has an input for every setting", () => {
    for (const [type, values] of Object.entries(nodeTypes)) {
      for (const input of values.dataFields) {
        if (input.startsWith("field:")) continue;
        expect(settingInputs[input], `${type} ${input}`).toBeDefined();
      }
    }
  });

  // Previews can only open the dialogs of these, a renamed one would stay
  // closed.
  it("has an input for every field with a dialog", () => {
    for (const field of dialogFields) {
      expect(settingInputs[field], field).toBeDefined();
    }
  });
});
