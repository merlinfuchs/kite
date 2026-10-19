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
  { value: "file_upload", label: "File Upload" },
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
    case "file_upload":
      return { ...base, required: true };
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

// Discord lets up to 10 files be uploaded to a file upload.
export const modalMaxFiles = 10;

export function modalInputHasValueLimits(type?: string) {
  return (
    type === "string_select" ||
    type === "checkbox_group" ||
    type === "file_upload" ||
    isOneOf(modalEntitySelectTypes, type)
  );
}

// Whether more than one option of the input can be picked, or more than one
// file uploaded. Discord lets one option of a select be picked by default, one
// file be uploaded, and all options of a checkbox group.
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

// Discord's length limits of the texts in a modal, which the block's schema
// enforces too. The editor shows them as counters.
export const modalTextLimits = {
  title: 45,
  label: 45,
  description: 100,
  placeholder: 100,
  value: 4000,
  content: 4000,
  custom_id: 100,
  option: 100,
} as const;

// The input after its type was changed in the editor. Unlike newModalInput,
// it keeps the settings the new type has too, so switching between similar
// types, e.g. a select menu and a radio group, doesn't lose the options.
export function changeModalInputType(
  input: ModalComponentData,
  type: ModalInputType
): ModalComponentData {
  const res = newModalInput(type, input.custom_id);
  if (type !== "checkbox" && input.type !== "checkbox") {
    res.required = input.required;
  }
  if (modalInputHasPlaceholder(type) && input.placeholder) {
    res.placeholder = input.placeholder;
  }
  if (modalInputHasOptions(type) && input.options?.length) {
    const [min, max] =
      modalOptionCounts[type as keyof typeof modalOptionCounts];
    const options = input.options.slice(0, max).map((o) => {
      // Only the options of select menus show emojis.
      const { emoji, ...rest } = o;
      return type === "string_select" && emoji ? { ...rest, emoji } : rest;
    });
    for (let i = options.length; i < min; i++) {
      options.push(newModalOption(options));
    }
    // A radio group picks one option, so only the first default stays.
    if (type === "radio_group") {
      let picked = false;
      for (const o of options) {
        if (o.default && picked) o.default = undefined;
        if (o.default) picked = true;
      }
    }
    res.options = options;
  }
  return res;
}

// An identifier made from a label, e.g. "Your favorite color?" becomes
// "your_favorite_color". It gets a number if another input uses it already.
export function modalIdentifierFromLabel(
  label: string,
  components: ModalComponentData[],
  ownIndex?: number
) {
  const base =
    label
      .toLowerCase()
      .replace(/{{.*?}}/g, "")
      .replace(/[^a-z0-9]+/g, "_")
      .replace(/^_+|_+$/g, "")
      .slice(0, 90) || "input";

  const ids = new Set(
    components
      .filter((_, i) => i !== ownIndex)
      .map((c) => c.components?.[0]?.custom_id)
  );
  let id = base;
  for (let n = 2; ids.has(id); n++) id = `${base}_${n}`;
  return id;
}

// A copy of the component at the given index, as inserted after it. A copied
// input gets an identifier of its own, as two inputs can't share one.
export function duplicateModalComponent(
  components: ModalComponentData[],
  index: number
): ModalComponentData {
  const component: ModalComponentData = structuredClone(components[index]);
  const input = component.components?.[0];
  if (input) {
    const ids = new Set(components.map((c) => c.components?.[0]?.custom_id));
    const base = `${input.custom_id || "input"}_copy`;
    let id = base;
    for (let n = 2; ids.has(id); n++) id = `${base}_${n}`;
    input.custom_id = id;
  }
  return component;
}

// Discord's numbers of the component types a modal can have, so modals in
// the format of Discord's API can be imported too.
const discordModalComponentTypes: Record<number, string> = {
  3: "string_select",
  4: "text_input",
  5: "user_select",
  6: "role_select",
  7: "mentionable_select",
  8: "channel_select",
  10: "text_display",
  18: "label",
  19: "file_upload",
  21: "radio_group",
  22: "checkbox_group",
  23: "checkbox",
};

const modalComponentKeys = [
  "type",
  "label",
  "description",
  "content",
  "components",
];
const modalInputKeys = [
  "type",
  "custom_id",
  "style",
  "label",
  "min_length",
  "max_length",
  "min_values",
  "max_values",
  "required",
  "value",
  "placeholder",
  "options",
  "channel_types",
  "default",
];
const modalOptionKeys = ["label", "value", "description", "default", "emoji"];

function isObject(v: unknown): v is Record<string, any> {
  return !!v && typeof v === "object" && !Array.isArray(v);
}

// Imports a modal from JSON, e.g. pasted in the editor. Besides Kite's own
// format, it accepts the format of Discord's API, which has numbered types,
// one "component" per label and named text input styles, and returns the
// fields Kite doesn't support, which are left out.
export function importModalData(data: unknown): {
  modal: ModalData;
  ignored: string[];
} {
  if (!isObject(data)) throw new Error("The modal must be an object");
  if (data.title !== undefined && typeof data.title !== "string") {
    throw new Error("title must be a text");
  }
  if (data.components !== undefined && !Array.isArray(data.components)) {
    throw new Error("components must be a list");
  }

  const ignored = new Set<string>();

  const pick = (obj: Record<string, any>, keys: string[], path: string) => {
    const res: Record<string, any> = {};
    for (const [key, value] of Object.entries(obj)) {
      if (keys.includes(key)) res[key] = value;
      else ignored.add(`${path}.${key}`);
    }
    return res;
  };

  const type = (t: unknown) =>
    typeof t === "number" ? discordModalComponentTypes[t] ?? t : t;

  const importInput = (input: Record<string, any>): ModalComponentData => {
    const res = pick(
      { ...input, type: type(input.type) },
      modalInputKeys,
      "input"
    );
    if (res.style === "short") res.style = 1;
    if (res.style === "paragraph") res.style = 2;
    if (Array.isArray(res.options)) {
      res.options = res.options.filter(isObject).map((o) => {
        const option = pick(o, modalOptionKeys, "option");
        // Only select menus show emojis, and only their ID and name are
        // kept, e.g. not Discord's "animated".
        if (option.emoji !== undefined) {
          const { id, name } = isObject(option.emoji)
            ? option.emoji
            : ({} as Record<string, any>);
          if (res.type === "string_select" && (id || name)) {
            option.emoji = {
              id: id ? String(id) : undefined,
              name: name ? String(name) : undefined,
            };
            if (!option.emoji.id) delete option.emoji.id;
            if (!option.emoji.name) delete option.emoji.name;
          } else {
            delete option.emoji;
            ignored.add("option.emoji");
          }
        }
        return option;
      });
    }
    return res;
  };

  const components = (data.components ?? [])
    .filter((c: unknown) => isObject(c) && Object.keys(c).length > 0)
    .map((c: Record<string, any>): ModalComponentData => {
      // Rows of the legacy format, Discord's action rows, are converted by
      // normalizeModalComponents.
      if (c.type === 1 || c.type === "action_row") {
        return {
          components: (c.components ?? []).filter(isObject).map(importInput),
        };
      }

      const { component, ...rest } = c;
      const res = pick(
        { ...rest, type: type(rest.type) },
        modalComponentKeys,
        "component"
      );

      // Discord's API gives a label its input as "component". Empty inputs
      // next to it are left over and dropped.
      const inputs = (
        isObject(component) ? [component] : res.components ?? []
      ).filter((i: unknown) => isObject(i) && Object.keys(i).length > 0);
      if (res.type === "text_display") {
        delete res.components;
      } else {
        res.components = inputs.map(importInput);
      }
      return res;
    });

  return {
    modal: normalizeModalData({ title: data.title, components }),
    ignored: Array.from(ignored),
  };
}
