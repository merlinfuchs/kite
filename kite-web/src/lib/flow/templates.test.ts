import { describe, expect, it } from "vitest";
import {
  getTemplates,
  templateFiles,
  templateFileSchema,
  templateSettingKeys,
  TemplateFile,
} from "./templates";

// The strings in the blocks of a template, with the path to each.
function blockStrings(file: TemplateFile) {
  const flows = [
    ...file.commands.map((c) => c.flow_source),
    ...file.event_listeners.map((l) => l.flow_source),
  ];
  const res: { path: string; value: string }[] = [];
  const walk = (value: unknown, path: string) => {
    if (typeof value === "string") res.push({ path, value });
    else if (Array.isArray(value))
      value.forEach((v, i) => walk(v, `${path}.${i}`));
    else if (value && typeof value === "object") {
      for (const [key, v] of Object.entries(value)) walk(v, `${path}.${key}`);
    }
  };
  for (const flow of flows) {
    for (const node of flow.nodes) walk(node.data, node.id);
  }
  return res;
}

// Files that don't match the format fail their own test below.
const files = templateFiles.flatMap((file) => {
  const res = templateFileSchema.safeParse(file);
  return res.success ? [res.data] : [];
});

describe("templates", () => {
  it("match the template file format", () => {
    for (const file of templateFiles) {
      const res = templateFileSchema.safeParse(file);
      expect(res.error?.issues, (file as TemplateFile).id).toBeUndefined();
    }
  });

  it("have unique ids", () => {
    const ids = getTemplates().map((t) => t.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  // The keys match what an app installed to a later version of the template.
  it("have unique item keys", () => {
    for (const template of getTemplates()) {
      const keys = [...template.commands, ...template.eventListeners].map(
        (item) => item.key
      );
      expect(new Set(keys).size, template.id).toBe(keys.length);
    }
  });

  it("only use their own settings, and all of them", () => {
    for (const file of files) {
      const declared = file.settings.map((s) => s.key);
      const used = new Set(
        blockStrings(file).flatMap((s) => templateSettingKeys(s.value))
      );
      expect(
        [...used].filter((key) => !declared.includes(key)),
        file.id
      ).toEqual([]);
      expect(
        declared.filter((key) => !used.has(key)),
        file.id
      ).toEqual([]);
    }
  });

  // An ID typed into a block only exists in the server it was copied from, it
  // has to be a setting instead.
  it("have no Discord IDs in their blocks", () => {
    for (const file of files) {
      const ids = blockStrings(file).filter((s) =>
        /\b\d{17,20}\b/.test(s.value)
      );
      expect(ids, file.id).toEqual([]);
    }
  });
});
