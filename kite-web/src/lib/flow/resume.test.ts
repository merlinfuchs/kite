import { describe, expect, it } from "vitest";
import { ComponentData } from "../types/message.gen";
import { getComponentHandleIds, selectOptionHandleId } from "./resume";

describe("getComponentHandleIds", () => {
  it("adds an output per string select option", () => {
    const components: ComponentData[] = [
      {
        type: 1,
        components: [
          { id: 1, type: 2, style: 1, label: "Click" },
          {
            id: 2,
            type: 3,
            options: [
              { id: 3, label: "A", value: "a" },
              { id: 4, label: "B", value: "b" },
            ],
          },
          { id: 5, type: 5 },
        ],
      },
    ];

    expect(getComponentHandleIds(components)).toEqual([
      "component_1",
      "component_2",
      selectOptionHandleId(2, 3),
      selectOptionHandleId(2, 4),
      "component_5",
    ]);
    expect(selectOptionHandleId(2, 3)).toBe("component_2_option_3");
  });
});
