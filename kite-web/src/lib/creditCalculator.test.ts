import { describe, expect, it } from "vitest";
import { suggestedActivity } from "./creditCalculator";
import { getCalculatorTemplates } from "./creditCalculatorTemplates";

describe("getCalculatorTemplates", () => {
  it("prices AI templates per model", () => {
    const askAi = getCalculatorTemplates().find((t) => t.id === "ask-ai")!;
    expect(askAi.credits.small).toBeLessThan(askAi.credits.large);
  });

  it("prices other templates the same for every model", () => {
    const welcomer = getCalculatorTemplates().find((t) => t.id === "welcomer")!;
    expect(welcomer.credits).toEqual({ small: 1, medium: 1, large: 1 });
  });
});

describe("suggestedActivity", () => {
  it("rounds to two digits", () => {
    expect(suggestedActivity({ perThousandMembers: 120 }, 2345)).toBe(280);
    expect(suggestedActivity({ perThousandMembers: 20 }, 100)).toBe(2);
    expect(suggestedActivity({ perThousandMembers: 150 }, 50000)).toBe(7500);
  });
});
