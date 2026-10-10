import {
  ModalComponentData,
  ModalComponentOptionData,
  ModalData,
} from "../types/flow.gen";

export const modalMaxComponents = 5;

export const modalInputTypes = [
  { value: "text_input", label: "Text Input" },
  { value: "string_select", label: "Select Menu" },
  { value: "user_select", label: "User Select" },
  { value: "role_select", label: "Role Select" },
  { value: "mentionable_select", label: "Mentionable Select" },
  { value: "channel_select", label: "Channel Select" },
  { value: "radio_group", label: "Radio Group" },
  { value: "checkbox_group", label: "Checkbox Group" },
  { value: "checkbox", label: "Checkbox" },
] as const;

export type ModalInputType = (typeof modalInputTypes)[number]["value"];

// Modals saved before labels existed are rows of text inputs that carry their
// own label. They become one label per input, as the service does in
// normalizeModalComponents.
export function normalizeModalComponents(
  components: ModalComponentData[] | undefined
): ModalComponentData[] {
  return (components ?? []).flatMap((c) => {
    if (!c || c.type) return c ? [c] : [];

    return (c.components ?? []).map(({ label, ...input }) => ({
      type: "label",
      label,
      components: [{ ...input, type: input.type || "text_input" }],
    }));
  });
}

export function normalizeModalData(data: ModalData): ModalData {
  return { ...data, components: normalizeModalComponents(data.components) };
}

function isModalText(c: ModalComponentData) {
  return c.type === "text_display";
}

// The number shown on a component in the editor. Inputs and texts are
// counted separately, so adding a text doesn't renumber the inputs.
export function modalComponentNumber(
  components: ModalComponentData[],
  index: number
) {
  const text = isModalText(components[index]);
  return components.slice(0, index + 1).filter((c) => isModalText(c) === text)
    .length;
}

// The number of a new input, which it uses in its label and identifier. It
// matches the number the editor shows for it, unless an input already uses
// that identifier, e.g. after one was removed.
export function nextModalInputNumber(components: ModalComponentData[]) {
  const ids = new Set(components.map((c) => c.components?.[0]?.custom_id));
  let n = components.filter((c) => !isModalText(c)).length + 1;
  while (ids.has(`input_${n}`)) n++;
  return n;
}

// The input a new label starts with, or a label's input after its type was
// changed. The identifier carries over, the rest depends on the type.
export function newModalInput(
  type: ModalInputType,
  customId?: string
): ModalComponentData {
  const base = { type, custom_id: customId };
  switch (type) {
    case "text_input":
      return { ...base, style: 1, required: true };
    case "string_select":
    case "radio_group":
      return {
        ...base,
        required: true,
        options: [
          { label: "Option 1", value: "option_1" },
          { label: "Option 2", value: "option_2" },
        ],
      };
    case "checkbox_group":
      return {
        ...base,
        required: true,
        options: [{ label: "Option 1", value: "option_1" }],
      };
    case "checkbox":
      return base;
    default:
      return { ...base, required: true };
  }
}

// The option added after the given ones. Its value must be unique, so it
// skips numbers that are taken, e.g. after an option was removed.
export function newModalOption(
  options: ModalComponentOptionData[]
): ModalComponentOptionData {
  const values = new Set(options.map((o) => o.value || o.label));
  let n = options.length + 1;
  while (values.has(`option_${n}`)) n++;
  return { label: `Option ${n}`, value: `option_${n}` };
}

export const modalOptionInputTypes = [
  "string_select",
  "radio_group",
  "checkbox_group",
] as const;

// How many options each input type needs at least and at most.
export const modalOptionCounts = {
  string_select: [1, 25],
  radio_group: [2, 10],
  checkbox_group: [1, 10],
} as const;

export const modalEntitySelectTypes = [
  "user_select",
  "role_select",
  "mentionable_select",
  "channel_select",
] as const;

function isOneOf(types: readonly string[], type?: string) {
  return types.includes(type ?? "");
}

export function modalInputHasOptions(type?: string) {
  return isOneOf(modalOptionInputTypes, type);
}

export function modalInputHasValueLimits(type?: string) {
  return (
    type === "string_select" ||
    type === "checkbox_group" ||
    isOneOf(modalEntitySelectTypes, type)
  );
}

// Whether more than one option of the input can be picked. Discord lets one
// option of a select be picked by default, and all of a checkbox group.
export function modalInputIsMultiValue(input: ModalComponentData) {
  if (!modalInputHasValueLimits(input.type)) return false;
  const defaultMax =
    input.type === "checkbox_group" ? input.options?.length ?? 0 : 1;
  return (input.max_values ?? defaultMax) > 1;
}

export function modalInputHasPlaceholder(type?: string) {
  return (
    !type ||
    type === "text_input" ||
    type === "string_select" ||
    isOneOf(modalEntitySelectTypes, type)
  );
}
