import { describe, expect, it } from "vitest";
import { FlowAIChatRequest, FlowAIChatResponse } from "../types/wire.gen";
import { runFlowAIPrompt } from "./ai";
import { testEdge, testNode } from "./testUtils";

const entry = testNode("entry", "entry_command", {
  name: "test",
  description: "Test",
});

const usage = { prompts_used: 1, prompts_limit: 30 };

type Edits = FlowAIChatResponse["edits"];

// fakeAPI answers each request with the next of the given edits, or fails if
// there are none left, and records the requests.
function fakeAPI(...rounds: (Edits | ((req: FlowAIChatRequest) => Edits))[]) {
  const requests: FlowAIChatRequest[] = [];
  const send = async (req: FlowAIChatRequest) => {
    requests.push(req);
    const round = rounds[requests.length - 1];
    if (!round) {
      return {
        success: false as const,
        error: { code: "repair_limit", message: "Too many repairs", data: {} },
      };
    }
    return {
      success: true as const,
      data: {
        prompt_id: "p1",
        message: requests.length === 1 ? "Added a log." : "",
        edits: typeof round === "function" ? round(req) : round,
        issues: [],
        usage,
      },
    };
  };
  return { requests, send };
}

const run = (send: ReturnType<typeof fakeAPI>["send"], nodes = [entry]) =>
  runFlowAIPrompt({
    flow: { nodes, edges: [] },
    context: "command",
    selectedIds: [],
    messages: [{ role: "user", content: "Add a log" }],
    send,
  });

describe("runFlowAIPrompt", () => {
  it("sends the problems with the edits back to be fixed", async () => {
    const api = fakeAPI(
      [{ op: "add_node", ref: "$log", type: "action_log", after: "entry" }],
      // The repair refers to the block by the ID it was sent with.
      (req) => [
        {
          op: "update_node",
          id: req.flow.match(/- (\S+) action_log/)![1],
          data: { log_level: "info", log_message: "hi" },
        },
      ]
    );

    const res = await run(api.send);

    expect(api.requests).toHaveLength(2);
    expect(api.requests[1].repair_prompt_id).toBe("p1");
    expect(api.requests[1].issues[0]).toContain("log_level");
    expect(api.requests[1].messages.at(-1)).toEqual({
      role: "assistant",
      content: "Added a log.",
    });
    expect(res.message).toBe("Added a log.");
    expect(res.repairs).toBe(1);
    expect(res.issues).toEqual([]);
    const log = res.nodes.find((n) => n.type === "action_log")!;
    expect(log.data).toEqual({ log_level: "info", log_message: "hi" });
    expect(res.changedNodeIds).toEqual([log.id]);
  });

  it("doesn't send problems the flow had before", async () => {
    const broken = testNode("broken", "action_log", {});
    const api = fakeAPI([]);

    const res = await run(api.send, [entry, broken]);

    expect(api.requests).toHaveLength(1);
    expect(res.issues).toEqual([]);
    expect(res.changedNodeIds).toEqual([]);
  });

  it("keeps the edits if a repair fails", async () => {
    const api = fakeAPI([
      { op: "add_node", ref: "$log", type: "action_log", after: "entry" },
    ]);

    const res = await run(api.send);

    expect(api.requests).toHaveLength(2);
    expect(res.nodes).toHaveLength(2);
    expect(res.issues.at(-1)).toBe("Too many repairs");
  });

  it("throws if the prompt fails", async () => {
    const api = fakeAPI();
    await expect(run(api.send)).rejects.toThrow("Too many repairs");
  });

  it("serializes the current flow", async () => {
    const api = fakeAPI([]);
    await runFlowAIPrompt({
      flow: { nodes: [entry], edges: [testEdge("entry", "missing")] },
      context: "command",
      selectedIds: ["entry"],
      messages: [{ role: "user", content: "Hi" }],
      send: api.send,
    });
    expect(api.requests[0].flow).toContain("- entry entry_command (selected)");
    expect(api.requests[0].flow_type).toBe("command");
  });
});
