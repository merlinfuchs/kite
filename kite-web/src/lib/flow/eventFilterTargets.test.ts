import { describe, expect, it } from "vitest";
import { discordOptionEventFilter } from "../blocks/discordOptionEventFilter";
import { discordEventTypes } from "./eventTypes";
import {
  eventFilterTargetLabels,
  eventFilterTargets,
} from "./eventFilterTargets";

describe("eventFilterTargets", () => {
  const blockTargets = (
    discordOptionEventFilter.fields!.find(
      (f) => f.name === "event_filter_target"
    )!.schema as any
  ).options as string[];

  it("only lists targets the block knows, each with a label", () => {
    for (const targets of Object.values(eventFilterTargets)) {
      for (const target of targets) {
        expect(blockTargets).toContain(target);
        expect(eventFilterTargetLabels[target]).toBeTruthy();
      }
    }
    expect(Object.keys(eventFilterTargetLabels).sort()).toEqual(
      [...blockTargets].sort()
    );
  });

  it("has targets for every Discord event", () => {
    for (const { value } of discordEventTypes) {
      expect(eventFilterTargets[value], value).toBeTruthy();
    }
  });
});
