import { Edge } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { FlowAIChatRequest, FlowAIChatResponse } from "../types/wire.gen";
import { runFlowAIPrompt } from "./ai";
import { testNode } from "./testUtils";

const entry = testNode("entry", "entry_command", {
  name: "test",
  description: "Test",
});

const usage = { prompts_used: 1, prompts_limit: 30 };
const logData = { log_level: "info", log_message: "hi" };

type Edits = FlowAIChatResponse["edits"];

// fakeAPI answers each request with the next of the given edits, or fails
// with the given error if there are none left, and records the requests.
function fakeAPI(
  rounds: (Edits | ((req: FlowAIChatRequest) => Edits))[],
  error = { code: "repair_limit", message: "Too many repairs", data: {} }
) {
  const requests: FlowAIChatRequest[] = [];
  const send = async (req: FlowAIChatRequest) => {
    requests.push(req);
    const round = rounds[requests.length - 1];
    if (!round) return { success: false as const, error };
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

// run runs a prompt against an editor holding the given blocks.
async function run(api: ReturnType<typeof fakeAPI>, nodes = [entry]) {
  const editor = {
    flow: { nodes, edges: [] as Edge[] },
    changedNodeIds: [] as string[],
  };
  const res = await runFlowAIPrompt({
    context: "command",
    messages: [{ role: "user", content: "Add a log" }],
    getFlow: () => editor.flow,
    applyFlow: (flow, changedNodeIds) => {
      editor.flow = flow;
      editor.changedNodeIds = changedNodeIds;
    },
    send: api.send,
  });
  return { ...res, editor };
}

const addLog: Edits = [
  { op: "add_node", ref: "$log", type: "action_log", after: "entry" },
];

describe("runFlowAIPrompt", () => {
  it("sends the problems with the edits back to be fixed", async () => {
    const api = fakeAPI([
      addLog,
      // The repair refers to the block by the ID it was sent with.
      (req) => [
        {
          op: "update_node",
          id: req.flow.match(/- (\S+) action_log/)![1],
          data: { log_level: "info", log_message: "hi" },
        },
      ],
    ]);

    const res = await run(api);

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
    const log = res.editor.flow.nodes.find((n) => n.type === "action_log")!;
    expect(log.data).toEqual({ log_level: "info", log_message: "hi" });
    expect(res.changedNodeIds).toEqual([log.id]);
    expect(res.editor.changedNodeIds).toEqual([log.id]);
  });

  it("doesn't send problems the flow had before", async () => {
    const api = fakeAPI([addLog]);

    const res = await run(api, [entry, testNode("broken", "action_log")]);

    expect(api.requests).toHaveLength(2);
    expect(api.requests[1].issues.length).toBeGreaterThan(0);
    expect(api.requests[1].issues.join()).not.toContain("broken");
  });

  it("stops quietly when no more repairs are allowed", async () => {
    const api = fakeAPI([addLog]);

    const res = await run(api);

    expect(api.requests).toHaveLength(2);
    expect(res.editor.flow.nodes).toHaveLength(2);
    expect(res.issues).not.toContain("Too many repairs");
    expect(res.issues.length).toBeGreaterThan(0);
  });

  it("reports why a repair failed", async () => {
    const api = fakeAPI([addLog], {
      code: "flow_ai_unavailable",
      message: "Unavailable",
      data: {},
    });

    const res = await run(api);

    expect(res.issues.at(-1)).toBe("Unavailable");
  });

  it("keeps problems a repair without edits didn't fix", async () => {
    const api = fakeAPI([addLog, []]);

    const res = await run(api);

    expect(api.requests).toHaveLength(3);
    expect(res.repairs).toBe(1);
    expect(res.issues.length).toBeGreaterThan(0);
  });

  it("doesn't send problems the user caused while waiting", async () => {
    const api = fakeAPI([[]]);
    const editor = { flow: { nodes: [entry], edges: [] as Edge[] } };
    const send = api.send;
    api.send = async (req: FlowAIChatRequest) => {
      editor.flow = {
        nodes: [entry, testNode("new", "action_log")],
        edges: [],
      };
      return send(req);
    };

    const res = await runFlowAIPrompt({
      context: "command",
      messages: [{ role: "user", content: "Hi" }],
      getFlow: () => editor.flow,
      applyFlow: (flow) => (editor.flow = flow),
      send: api.send,
    });

    expect(api.requests).toHaveLength(1);
    expect(res.issues).toEqual([]);
  });

  it("stops repairing edits the user undid", async () => {
    const api = fakeAPI([addLog, []]);
    const editor = { flow: { nodes: [entry], edges: [] as Edge[] } };
    const send = api.send;
    api.send = async (req: FlowAIChatRequest) => {
      // Undo of the first round.
      if (req.repair_prompt_id) editor.flow = { nodes: [entry], edges: [] };
      return send(req);
    };

    const res = await runFlowAIPrompt({
      context: "command",
      messages: [{ role: "user", content: "Add a log" }],
      getFlow: () => editor.flow,
      applyFlow: (flow) => (editor.flow = flow),
      send: api.send,
    });

    expect(api.requests).toHaveLength(2);
    expect(res.issues).toEqual([]);
    expect(editor.flow.nodes).toHaveLength(1);
  });

  it("stops when aborted", async () => {
    const controller = new AbortController();
    const api = fakeAPI([addLog]);
    const send = api.send;
    api.send = async (req: FlowAIChatRequest) => {
      controller.abort();
      return send(req);
    };

    await expect(
      runFlowAIPrompt({
        context: "command",
        messages: [{ role: "user", content: "Add a log" }],
        getFlow: () => ({ nodes: [entry], edges: [] }),
        applyFlow: () => {},
        send: api.send,
        signal: controller.signal,
      })
    ).rejects.toThrow();
    expect(api.requests).toHaveLength(1);
  });

  it("throws if the prompt fails", async () => {
    await expect(run(fakeAPI([]))).rejects.toThrow("Too many repairs");
  });

  it("applies the edits to the flow as it is when they arrive", async () => {
    const other = testNode("other", "action_log", logData);
    const api = fakeAPI([]);
    const editor = { flow: { nodes: [entry], edges: [] as Edge[] } };
    // The user adds a block while the AI works on the prompt.
    const send = fakeAPI([[...addLog]]).send;
    api.send = async (req: FlowAIChatRequest) => {
      if (!req.repair_prompt_id)
        editor.flow = { nodes: [entry, other], edges: [] };
      return send(req);
    };

    await runFlowAIPrompt({
      context: "command",
      messages: [{ role: "user", content: "Add a log" }],
      getFlow: () => editor.flow,
      applyFlow: (flow) => (editor.flow = flow),
      send: api.send,
    });

    expect(editor.flow.nodes.map((n) => n.id)).toContain("other");
    expect(editor.flow.nodes).toHaveLength(3);
  });

  it("sends only the blocks the user selected as selected", async () => {
    const api = fakeAPI([addLog]);
    await run(api, [{ ...entry, selected: true }]);
    expect(api.requests[0].flow).toContain("- entry entry_command (selected)");
    // The added block is selected to highlight it, but not sent as selected.
    expect(api.requests[1].flow.match(/\(selected\)/g)).toHaveLength(1);
  });
});
