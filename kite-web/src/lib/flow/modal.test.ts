import { describe, expect, it } from "vitest";
import { nodeSuspendResponseModalDataSchema } from "./dataSchema";
import {
  modalComponentNumber,
  nextModalInputNumber,
  normalizeModalComponents,
} from "./modal";

describe("normalizeModalComponents", () => {
  it("turns legacy rows into one label per text input", () => {
    expect(
      normalizeModalComponents([
        {
          components: [
            { custom_id: "a", label: "A", style: 1 },
            { custom_id: "b", label: "B", style: 2 },
          ],
        },
        { type: "text_display", content: "Hi" },
      ])
    ).toEqual([
      {
        type: "label",
        label: "A",
        components: [{ type: "text_input", custom_id: "a", style: 1 }],
      },
      {
        type: "label",
        label: "B",
        components: [{ type: "text_input", custom_id: "b", style: 2 }],
      },
      { type: "text_display", content: "Hi" },
    ]);
  });
});

describe("nodeSuspendResponseModalDataSchema", () => {
  const parse = (components: unknown[]) =>
    nodeSuspendResponseModalDataSchema.safeParse({
      modal_data: { title: "Form", components },
    }).success;

  it("accepts legacy modals", () => {
    expect(
      parse([{ components: [{ custom_id: "a", label: "A", style: 1 }] }])
    ).toBe(true);
  });

  it("accepts the new components", () => {
    expect(
      parse([
        { type: "text_display", content: "Hi" },
        {
          type: "label",
          label: "Size",
          components: [
            {
              type: "radio_group",
              custom_id: "size",
              options: [{ label: "S" }, { label: "M" }],
            },
          ],
        },
        {
          type: "label",
          label: "Agree",
          components: [{ type: "checkbox", custom_id: "agree" }],
        },
      ])
    ).toBe(true);
  });

  it("needs an input", () => {
    expect(parse([{ type: "text_display", content: "Hi" }])).toBe(false);
  });

  it("needs two options in a radio group", () => {
    expect(
      parse([
        {
          type: "label",
          label: "Size",
          components: [
            {
              type: "radio_group",
              custom_id: "size",
              options: [{ label: "S" }],
            },
          ],
        },
      ])
    ).toBe(false);
  });
});

describe("modal numbering", () => {
  const text = { type: "text_display", content: "Hi" };
  const input = (id: string) => ({
    type: "label",
    label: id,
    components: [{ type: "text_input", custom_id: id }],
  });

  it("counts inputs and texts separately", () => {
    const components = [text, input("input_1"), text, input("input_2")];
    expect(
      components.map((_, i) => modalComponentNumber(components, i))
    ).toEqual([1, 1, 2, 2]);
  });

  it("numbers a new input after a text like the editor shows it", () => {
    expect(nextModalInputNumber([text])).toBe(1);
    expect(nextModalInputNumber([text, input("input_1")])).toBe(2);
  });

  it("skips identifiers that are taken", () => {
    expect(nextModalInputNumber([input("input_2")])).toBe(3);
  });
});
