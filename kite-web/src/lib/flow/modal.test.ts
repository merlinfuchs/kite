import { describe, expect, it } from "vitest";
import {
  modalComponentNumber,
  nextModalInputNumber,
  normalizeModalComponents,
} from "./modal";
import { getNodeValues } from "./nodes";

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

describe("suspend_response_modal data schema", () => {
  const schema = getNodeValues("suspend_response_modal").dataSchema!;
  const parse = (components: unknown[]) =>
    schema.safeParse({
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

  // The paths of the issues of a modal with the given inputs, relative to
  // the first input.
  const issues = (...inputs: Record<string, unknown>[]) =>
    (
      schema.safeParse({
        modal_data: {
          title: "Form",
          components: inputs.map((input, i) => ({
            type: "label",
            label: `Input ${i}`,
            components: [input],
          })),
        },
      }).error?.issues ?? []
    ).map((i) => i.path.slice(2).join("."));

  const options = (...labels: string[]) => labels.map((label) => ({ label }));

  it("accepts valid limits", () => {
    expect(
      issues(
        {
          type: "string_select",
          custom_id: "a",
          max_values: 2,
          options: [
            { label: "A", default: true },
            { label: "B", default: true },
          ],
        },
        {
          type: "checkbox_group",
          custom_id: "b",
          options: [
            { label: "A", default: true },
            { label: "B", default: true },
          ],
        },
        { type: "user_select", custom_id: "c", min_values: 2, max_values: 3 },
        { type: "text_input", custom_id: "d", style: 1, placeholder: "a" }
      )
    ).toEqual([]);
  });

  it("needs unique identifiers", () => {
    expect(
      issues(
        { type: "text_input", custom_id: "a", style: 1 },
        { type: "checkbox", custom_id: "a" }
      )
    ).toEqual(["1.components.0.custom_id"]);
  });

  it("needs unique option values", () => {
    expect(
      issues({
        type: "string_select",
        custom_id: "a",
        options: [{ label: "A" }, { label: "B", value: "A" }],
      })
    ).toEqual(["0.components.0.options.1.value"]);
  });

  it("limits picks to the options", () => {
    expect(
      issues({
        type: "checkbox_group",
        custom_id: "a",
        max_values: 3,
        options: options("A", "B"),
      })
    ).toEqual(["0.components.0.max_values"]);
    expect(
      issues({ type: "role_select", custom_id: "a", min_values: 2 })
    ).toEqual(["0.components.0.min_values"]);
  });

  it("limits options picked by default", () => {
    const defaults = [
      { label: "A", default: true },
      { label: "B", default: true },
    ];
    expect(
      issues(
        { type: "radio_group", custom_id: "a", options: defaults },
        { type: "string_select", custom_id: "b", options: defaults },
        {
          type: "checkbox_group",
          custom_id: "c",
          max_values: 1,
          options: defaults,
        }
      )
    ).toEqual([
      "0.components.0.options",
      "1.components.0.options",
      "2.components.0.options",
    ]);
  });

  it("limits text input placeholders to 100 characters", () => {
    expect(
      issues(
        {
          type: "text_input",
          custom_id: "a",
          style: 1,
          placeholder: "a".repeat(101),
        },
        { type: "user_select", custom_id: "b", placeholder: "a".repeat(150) }
      )
    ).toEqual(["0.components.0.placeholder"]);
  });

  it("checks the option count of each type", () => {
    expect(
      issues({
        type: "checkbox_group",
        custom_id: "a",
        options: options(..."ABCDEFGHIJK"),
      })
    ).toEqual(["0.components.0.options"]);
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
