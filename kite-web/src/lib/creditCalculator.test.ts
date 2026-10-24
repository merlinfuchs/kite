import { describe, expect, it } from "vitest";
import { suggestedActivity } from "./creditCalculator";
import { getCalculatorTemplates } from "./creditCalculatorTemplates";

describe("getCalculatorTemplates", () => {
  // Fails for a new template until it's added to calculatorActivities.
  it("has every template", () => {
    expect(getCalculatorTemplates().length).toBeGreaterThan(0);
  });

  it("prices AI templates per model", () => {
    const askAi = getCalculatorTemplates().find((t) => t.id === "ask-ai")!;
    expect(askAi.aiCredits!.small).toBeLessThan(askAi.aiCredits!.large);
    expect(askAi.credits).toBe(askAi.aiCredits!.small);
  });

  it("prices other templates once", () => {
    const welcomer = getCalculatorTemplates().find((t) => t.id === "welcomer")!;
    expect(welcomer).toMatchObject({ credits: 1, aiCredits: null });
  });
});

describe("suggestedActivity", () => {
  it("rounds to two digits", () => {
    expect(suggestedActivity({ perThousandMembers: 120 }, 2345)).toBe(280);
    expect(suggestedActivity({ perThousandMembers: 20 }, 100)).toBe(2);
    expect(suggestedActivity({ perThousandMembers: 150 }, 50000)).toBe(7500);
  });
});
