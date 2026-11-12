import { describe, expect, it } from "vitest";
import {
  changeModalInputType,
  duplicateModalComponent,
  importModalData,
  modalComponentNumber,
  modalIdentifierFromLabel,
  newModalOption,
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
        { type: "text_input", custom_id: "d", style: 1, placeholder: "a" },
        { type: "file_upload", custom_id: "e", min_values: 2, max_values: 10 }
      )
    ).toEqual([]);
  });

  it("limits the files of a file upload", () => {
    expect(
      issues({ type: "file_upload", custom_id: "a", max_values: 11 })
    ).toEqual(["0.components.0.max_values"]);
    expect(
      issues({ type: "file_upload", custom_id: "a", min_values: 2 })
    ).toEqual(["0.components.0.min_values"]);
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
    expect(
      issues({ type: "role_select", custom_id: "a", min_values: 0 })
    ).toEqual(["0.components.0.min_values"]);
  });

  it("limits the minimum length to the maximum", () => {
    expect(
      issues({
        type: "text_input",
        custom_id: "a",
        style: 1,
        min_length: 10,
        max_length: 5,
      })
    ).toEqual(["0.components.0.min_length"]);
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

  it("skips option values that are taken", () => {
    expect(newModalOption([{ label: "Option 2", value: "option_2" }])).toEqual({
      label: "Option 3",
      value: "option_3",
    });
    expect(newModalOption([{ label: "option_2" }])).toEqual({
      label: "Option 3",
      value: "option_3",
    });
  });
});

describe("modal editor helpers", () => {
  const label = (id: string) => ({
    type: "label",
    label: id,
    components: [{ type: "text_input", custom_id: id, style: 1 }],
  });

  it("keeps the options and placeholder when the type changes", () => {
    const select = {
      type: "string_select",
      custom_id: "color",
      required: false,
      placeholder: "Pick one",
      options: [
        { label: "Red", default: true, emoji: { name: "🔴" } },
        { label: "Blue", default: true },
      ],
    };

    expect(changeModalInputType(select, "radio_group")).toEqual({
      type: "radio_group",
      custom_id: "color",
      required: false,
      options: [{ label: "Red", default: true }, { label: "Blue" }],
    });
    expect(changeModalInputType(select, "user_select")).toEqual({
      type: "user_select",
      custom_id: "color",
      required: false,
      placeholder: "Pick one",
    });
  });

  it("adds options a type needs at least", () => {
    const res = changeModalInputType(
      { type: "checkbox_group", custom_id: "a", options: [{ label: "A" }] },
      "radio_group"
    );
    expect(res.options).toHaveLength(2);
  });

  it("makes identifiers from labels", () => {
    expect(modalIdentifierFromLabel("Your favorite color?", [])).toBe(
      "your_favorite_color"
    );
    expect(modalIdentifierFromLabel("{{user.name}}", [])).toBe("input");
    expect(modalIdentifierFromLabel("Color", [label("color")])).toBe("color_2");
    expect(modalIdentifierFromLabel("Color", [label("color")], 0)).toBe(
      "color"
    );
  });

  it("gives duplicated inputs their own identifier", () => {
    const components = [label("name"), label("name_copy")];
    expect(
      duplicateModalComponent(components, 0).components?.[0].custom_id
    ).toBe("name_copy_2");
    expect(components[0].components[0].custom_id).toBe("name");
  });
});

describe("importModalData", () => {
  const schema = getNodeValues("suspend_response_modal").dataSchema!;

  it("imports modals in the format of Discord's API", () => {
    const { modal, ignored } = importModalData({
      title: "Developer Onboarding",
      components: [
        { type: "text_display", content: "### Welcome" },
        {
          type: "label",
          label: "Focus",
          component: {
            type: "string_select",
            custom_id: "focus_select",
            required: true,
            options: [
              { label: "Frontend", value: "frontend", emoji: { name: "💻" } },
              { label: "Backend", value: "backend" },
            ],
          },
        },
        {
          type: 18,
          label: "Experience",
          component: {
            type: 4,
            custom_id: "experience_text",
            style: "paragraph",
            min_length: 10,
          },
          components: [{}],
        },
      ],
    });

    expect(modal.components?.[1].components?.[0].options?.[0]).toEqual({
      label: "Frontend",
      value: "frontend",
      emoji: { name: "💻" },
    });
    expect(modal.components?.[2]).toEqual({
      type: "label",
      label: "Experience",
      components: [
        {
          type: "text_input",
          custom_id: "experience_text",
          style: 2,
          min_length: 10,
        },
      ],
    });
    expect(ignored).toEqual([]);
    expect(schema.safeParse({ modal_data: modal }).success).toBe(true);
  });

  it("leaves out emojis of options that can't show them", () => {
    const { modal, ignored } = importModalData({
      components: [
        {
          type: "label",
          label: "Pick",
          component: {
            type: "radio_group",
            custom_id: "pick",
            options: [
              { label: "A", emoji: { name: "🔴" } },
              { label: "B", emoji: { id: "1", name: "b", animated: true } },
            ],
          },
        },
      ],
    });
    expect(modal.components?.[0].components?.[0].options).toEqual([
      { label: "A" },
      { label: "B" },
    ]);
    expect(ignored).toEqual(["option.emoji"]);
  });

  it("rejects emojis on options of groups", () => {
    const issues = (type: string) => {
      const res = schema.safeParse({
        modal_data: {
          title: "Form",
          components: [
            {
              type: "label",
              label: "Pick",
              components: [
                {
                  type,
                  custom_id: "pick",
                  options: [
                    { label: "A", emoji: { name: "🔴" } },
                    { label: "B" },
                  ],
                },
              ],
            },
          ],
        },
      });
      return res.success ? [] : res.error.issues.map((i) => i.path.join("."));
    };
    expect(issues("string_select")).toEqual([]);
    expect(issues("radio_group")).toEqual([
      "modal_data.components.0.components.0.options.0.emoji",
    ]);
  });

  it("keeps modals in Kite's format as they are", () => {
    const data = {
      title: "Form",
      components: [
        {
          type: "label",
          label: "Name",
          components: [{ type: "text_input", custom_id: "name", style: 1 }],
        },
      ],
    };
    expect(importModalData(data)).toEqual({ modal: data, ignored: [] });
  });

  it("leaves out fields of the wrong type", () => {
    const { modal, ignored } = importModalData({
      components: [
        { type: "text_display", content: { text: "Hi" } },
        { type: 1, components: "nope" },
        {
          type: "label",
          label: ["Pick"],
          component: {
            type: "string_select",
            custom_id: 5,
            options: "a,b",
          },
        },
        {
          type: "label",
          label: "Color",
          component: {
            type: "string_select",
            custom_id: "color",
            options: [{ label: "Red", emoji: { id: 123, name: "red" } }],
          },
        },
      ],
    });
    expect(modal.components).toEqual([
      { type: "text_display" },
      { type: "label", components: [{ type: "string_select" }] },
      {
        type: "label",
        label: "Color",
        components: [
          {
            type: "string_select",
            custom_id: "color",
            options: [{ label: "Red", emoji: { name: "red" } }],
          },
        ],
      },
    ]);
    expect(ignored).toEqual([
      "component.content",
      "component.label",
      "input.custom_id",
      "input.options",
      "option.emoji.id",
    ]);
  });

  it("rejects what isn't a modal", () => {
    expect(() => importModalData([])).toThrow();
    expect(() => importModalData({ components: {} })).toThrow();
  });
});
