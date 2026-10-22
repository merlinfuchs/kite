import {
  BlockDefinition,
  BlockField,
  BlockOperation,
  BlockOperationField,
} from "./types";

// Settings the Popcat blocks share. Which of them an endpoint sends, and
// what they mean for it, is in its operation, see design/integrations.md.
const settings = {
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
} satisfies Record<string, [string, string]>;

// A Popcat block, with the settings its operations send.
export function popcatBlock({
  setting_label,
  operations,
  response,
  ...block
}: Omit<
  BlockDefinition,
  "category" | "credits" | "strict_settings" | "run" | "fields"
> & {
  setting_label: string;
  operations: BlockOperation[];
  response?: "url";
}): BlockDefinition {
  const sent = new Set(operations.flatMap((o) => o.fields.map((f) => f.name)));
  return {
    ...block,
    category: "Popcat",
    credits: 1,
    strict_settings: true,
    run: {
      kind: "request_choice",
      integration: "popcat",
      setting: "popcat_endpoint",
      setting_label,
      operations,
      // Popcat wraps results in {"error": false, "message": ...}.
      ...(response ? { response } : { result_key: "message" }),
    },
    fields: Object.entries(settings)
      .filter(([name]) => sent.has(name))
      .map(
        ([name, [label, description]]): BlockField => ({
          name,
          type: "string",
          label,
          description,
        })
      ),
  };
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
