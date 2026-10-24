import { describe, expect, it } from "vitest";
import { getTemplates, templateItemKey } from "./templates";

describe("templates", () => {
  it("have unique ids", () => {
    const ids = getTemplates().map((t) => t.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  // The keys match what an app installed to a later version of the template.
  it("have unique item keys", () => {
    for (const template of getTemplates()) {
      const keys = [...template.commands, ...template.eventListeners].map(
        templateItemKey
      );
      expect(new Set(keys).size, template.id).toBe(keys.length);
    }
  });
});
