import { describe, expect, it } from "vitest";
import { formatDiscordTimestamp } from "./timestamp";

describe("formatDiscordTimestamp", () => {
  it("names today and yesterday", () => {
    const today = new Date();
    today.setHours(15, 4);
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);

    expect(formatDiscordTimestamp(today, "en-US")).toBe("Today at 3:04 PM");
    expect(formatDiscordTimestamp(yesterday, "en-US")).toBe(
      "Yesterday at 3:04 PM"
    );
  });

  it("shows older dates in the locale's format", () => {
    const date = new Date(2026, 9, 4, 15, 4);

    expect(formatDiscordTimestamp(date, "en-US")).toBe("10/04/2026 3:04 PM");
    expect(formatDiscordTimestamp(date, "en-GB")).toBe("04/10/2026 15:04");
    expect(formatDiscordTimestamp(date, "de-DE")).toBe("04.10.2026 15:04");
  });
});
