import { describe, expect, it } from "vitest";
import { convertInventorItems } from "./convert";
import { parseInventorExport } from "./parse";

// A command as the inventor.bot dashboard returned it, with the text typed in.
function exportOf(text: string, extra: unknown[] = []) {
  return JSON.stringify({
    kite_inventor_export: 1,
    items: [
      {
        name: "ping",
        type: "flow",
        blob: {
          block_list: {
            id: "a",
            blocks: [
              {
                id: "b",
                tag: "",
                internal_id: "send_advanced_message",
                deactivated: false,
                fields: [
                  {
                    id: "b-message_components",
                    internal_id: "message_components",
                    type: "block_list",
                    value: "",
                    block_list: {
                      id: "c",
                      blocks: [
                        {
                          id: "d",
                          tag: "hi",
                          internal_id: "advmsgcmp_content_text",
                          deactivated: false,
                          fields: text
                            ? [
                                {
                                  id: "d-content",
                                  internal_id: "content",
                                  type: "text",
                                  value: text,
                                  block_list: { id: "", blocks: null },
                                },
                              ]
                            : [],
                        },
                        ...extra,
                      ],
                    },
                  },
                ],
              },
            ],
          },
          trigger: {
            id: "t",
            internal_id: "slash_command",
            fields: [
              {
                id: "t-1",
                internal_id: "command_integration_types",
                type: "multi_select",
                value: "Server Installable",
                block_list: { id: "", blocks: null },
              },
            ],
          },
          local_variables: [],
        },
      },
    ],
  });
}

describe("inventor.bot import", () => {
  it("turns a slash command that sends text into a command", () => {
    const [item] = convertInventorItems(parseInventorExport(exportOf("Pong!")));
    expect(item.status).toBe("ready");
    expect(item.name).toBe("ping");
    const types = item.flow!.nodes.map((n) => n.type);
    expect(types).toContain("entry_command");
    expect(types).toContain("action_response_create");
    expect(JSON.stringify(item.flow)).toContain("Pong!");
  });

  it("can't import a message without text", () => {
    const [item] = convertInventorItems(parseInventorExport(exportOf("")));
    expect(item.status).toBe("unsupported");
    expect(item.flow).toBeNull();
  });

  it("lists message parts it doesn't know instead of guessing", () => {
    const [item] = convertInventorItems(
      parseInventorExport(
        exportOf("Pong!", [
          {
            id: "e",
            tag: "",
            internal_id: "advmsgcmp_embed",
            deactivated: false,
            fields: [],
          },
        ])
      )
    );
    expect(item.status).toBe("review");
    expect(item.notes.join(" ")).toContain("embed");
  });

  it("can't import triggers other than slash commands", () => {
    const text = exportOf("Pong!").replace('"slash_command"', '"member_join"');
    const [item] = convertInventorItems(parseInventorExport(text));
    expect(item.status).toBe("unsupported");
    expect(item.notes.join(" ")).toContain("member join");
  });

  it("explains text that isn't from the helper", () => {
    expect(() => parseInventorExport("not json")).toThrow(/helper/);
  });
});
