import { BlockField, BlockOperation, BlockOperationField } from "./types";

// Settings the Popcat blocks share. Which of them an endpoint sends, and
// what they mean for it, is in its operation, see design/integrations.md.
const settings: Record<string, [string, string]> = {
  popcat_text: [
    "Text",
    "The main text the endpoint takes, like the text to put on a meme, the text to change or what to search for.",
  ],
  popcat_text_2: ["Second Text", "The second text of endpoints that take two."],
  popcat_text_3: ["Third Text", "The third text of endpoints that take three."],
  popcat_image: [
    "Image URL",
    "URL of an image, like {{user.avatar_url}}. Popcat reads PNG and JPEG images best.",
  ],
  popcat_image_2: [
    "Second Image URL",
    "URL of the second image of endpoints that take two.",
  ],
  popcat_color: ["Color", "A hex color without #, like ff0000."],
};

// The settings of a Popcat block, for the given names.
export function popcatFields(...names: (keyof typeof settings)[]) {
  return names.map(
    (name): BlockField => ({
      name,
      type: "string",
      label: settings[name][0],
      description: settings[name][1],
    })
  );
}

// A query parameter of an operation, sent from the given setting.
export function popcatQuery(
  name: keyof typeof settings,
  target: string,
  label?: string,
  description?: string,
  required = true
): BlockOperationField {
  return { name, in: "query", target, required, label, description };
}

// A path parameter of an operation, sent from the given setting.
export function popcatPath(
  name: keyof typeof settings,
  target: string,
  label?: string,
  description?: string
): BlockOperationField {
  return { name, in: "path", target, required: true, label, description };
}

// A Popcat endpoint, as an operation whose ID is the endpoint's path.
export function popcatOperation(
  path: string,
  label: string,
  description: string,
  ...fields: BlockOperationField[]
): BlockOperation {
  const id = path
    .replace(/^\/v2\//, "")
    .replace(/\/\{[^}]+\}$/, "")
    .replace(/-/g, "_");
  return { id, label, description, method: "GET", path, fields };
}
