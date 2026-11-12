import { describe, expect, it } from "vitest";
import { creatorBadges } from "./creatorBadges";

describe("creatorBadges", () => {
  it("gives early creators and milestones their badges", () => {
    expect(
      creatorBadges({
        installerCount: 120,
        firstListedAt: "2026-10-28T10:00:00Z",
      })
    ).toEqual([{ kind: "early" }, { kind: "milestone", installers: 100 }]);
  });

  it("gives creators after the launch and below 10 users none", () => {
    expect(
      creatorBadges({
        installerCount: 9,
        firstListedAt: "2026-11-20T10:00:00Z",
      })
    ).toEqual([]);
  });

  it("counts a milestone from exactly its number of users", () => {
    expect(creatorBadges({ installerCount: 1000 })).toEqual([
      { kind: "milestone", installers: 1000 },
    ]);
  });
});
