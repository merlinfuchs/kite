import { describe, expect, it } from "vitest";
import { messageSchema } from "../message/schema";
import { getTemplates, templateFiles } from "./templateFiles";
import { parameterKeys, referencedParameters } from "./parameters";
import { templateFileSchema, TemplateFile } from "./templates";

function templateFlows(file: TemplateFile) {
  return [
    ...file.commands.map((c) => c.flow_source),
    ...file.event_listeners.map((l) => l.flow_source),
    ...(file.messages ?? []).flatMap((m) => Object.values(m.flow_sources)),
  ];
}

// The strings in the blocks and messages of a template, with the path to
// each.
function blockStrings(file: TemplateFile) {
  const flows = templateFlows(file);
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
  for (const message of file.messages ?? []) {
    walk(message.data, message.id);
    if (message.send_to) walk(message.send_to, `${message.id}.send_to`);
  }
  return res;
}

// Files that don't match the format fail their own test below.
const files = templateFiles.flatMap((file) => {
  const res = templateFileSchema.safeParse(file);
  return res.success ? [res.data] : [];
});

describe("templates", () => {
  it("match the file embedded in the service", async () => {
    // Without usage, which is only for the credit calculator of the pricing
    // page.
    const installed = (templateFiles as TemplateFile[]).map((t) => ({
      id: t.id,
      version: t.version,
      name: t.name,
      description: t.description,
      icon: t.icon,
      category: t.category,
      about: t.about,
      faq: t.faq,
      setup: t.setup ?? [],
      parameters: t.parameters,
      commands: t.commands,
      event_listeners: t.event_listeners,
      messages: t.messages ?? [],
      variables: t.variables ?? [],
    }));
    await expect(JSON.stringify(installed, null, 2) + "\n").toMatchFileSnapshot(
      "../../../../kite-service/internal/core/bundle/templates.json"
    );
  });

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
      const keys = [
        ...template.commands,
        ...template.eventListeners,
        ...template.messages,
        ...template.variables,
      ].map((item) => item.key);
      expect(new Set(keys).size, template.id).toBe(keys.length);
    }
  });

  it("only use their own parameters, and all of them", () => {
    for (const file of files) {
      const declared = file.parameters.map((p) => p.key);
      const used = new Set([
        ...templateFlows(file).flatMap((flow) => referencedParameters(flow)),
        ...(file.messages ?? []).flatMap((m) => (m.send_to ? [m.send_to] : [])),
      ]);
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
  // has to be a parameter instead.
  it("have no Discord IDs in their blocks", () => {
    for (const file of files) {
      const ids = blockStrings(file).filter((s) =>
        /\b\d{17,20}\b/.test(s.value)
      );
      expect(ids, file.id).toEqual([]);
    }
  });

  it("have messages the message editor accepts", () => {
    for (const file of files) {
      for (const message of file.messages ?? []) {
        const res = messageSchema.safeParse(message.data);
        expect(res.error?.issues ?? [], message.id).toEqual([]);
      }
    }
  });

  // Blocks can only use the messages that are added along with them.
  it("only use their own messages", () => {
    for (const file of files) {
      const ids = (file.messages ?? []).map((m) => m.id);
      const used = templateFlows(file).flatMap((flow) =>
        flow.nodes.flatMap((n) =>
          typeof n.data.message_template_id === "string"
            ? [n.data.message_template_id]
            : []
        )
      );
      expect(
        used.filter((id) => !ids.includes(id)),
        file.id
      ).toEqual([]);
    }
  });

  // Blocks can only use the variables that are added along with them.
  it("only use their own variables, and all of them", () => {
    for (const file of files) {
      const ids = (file.variables ?? []).map((v) => v.id);
      const used = new Set(
        templateFlows(file).flatMap((flow) =>
          flow.nodes.flatMap((n) =>
            typeof n.data.variable_id === "string" ? [n.data.variable_id] : []
          )
        )
      );
      expect(
        [...used].filter((id) => !ids.includes(id)),
        file.id
      ).toEqual([]);
      expect(
        ids.filter((id) => !used.has(id)),
        file.id
      ).toEqual([]);
    }
  });

  it("post messages in channel parameters", () => {
    for (const file of files) {
      for (const message of file.messages ?? []) {
        if (!message.send_to) continue;
        const parameter = file.parameters.find(
          (p) => p.key === message.send_to
        );
        expect(parameter?.type, message.id).toBe("channel");
      }
    }
  });

  // Only flows have parameters, the content of a message can't use them yet.
  it("have no parameters in their messages", () => {
    for (const file of files) {
      for (const message of file.messages ?? []) {
        expect(parameterKeys(JSON.stringify(message.data)), message.id).toEqual(
          []
        );
      }
    }
  });
});
