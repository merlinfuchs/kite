import { describe, expect, it } from "vitest";
import { forecastUsage } from "./usageForecast";

// The days of October 2026 up to the given day, with the same credits each.
function october(untilDay: number, credits: number) {
  return Array.from({ length: untilDay }, (_, i) => ({
    date: new Date(Date.UTC(2026, 9, i + 1)).toISOString(),
    credits_used: credits,
  }));
}

describe("forecastUsage", () => {
  it("needs a few full days", () => {
    const now = new Date(Date.UTC(2026, 9, 3, 12));
    expect(forecastUsage(october(3, 100), 250, 10000, now)).toBeNull();
  });

  it("projects the rest of the month at the daily rate", () => {
    // Noon on October 11th leaves 20.5 days.
    const now = new Date(Date.UTC(2026, 9, 11, 12));
    const forecast = forecastUsage(october(11, 100), 1050, 10000, now);
    expect(forecast).toEqual({ dailyRate: 100, projected: 3100 });
  });

  it("uses the last week only", () => {
    const now = new Date(Date.UTC(2026, 9, 11, 12));
    const days = october(11, 100);
    days[0].credits_used = 5000;
    expect(forecastUsage(days, 6000, 100000, now)?.dailyRate).toBe(100);
  });

  it("starts with the first day that used credits", () => {
    const now = new Date(Date.UTC(2026, 9, 11, 12));
    const days = october(11, 100);
    for (const d of days.slice(0, 7)) d.credits_used = 0;
    expect(forecastUsage(days, 400, 10000, now)?.dailyRate).toBe(100);
  });

  it("says when the credits run out", () => {
    const now = new Date(Date.UTC(2026, 9, 11));
    const forecast = forecastUsage(october(11, 500), 5000, 10000, now);
    expect(forecast?.runsOutAt).toEqual(new Date(Date.UTC(2026, 9, 21)));
  });

  it("has nothing to say once they ran out", () => {
    const now = new Date(Date.UTC(2026, 9, 11));
    expect(forecastUsage(october(11, 1000), 10000, 10000, now)).toBeNull();
  });
});
