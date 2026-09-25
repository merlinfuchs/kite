import { mkdirSync, writeFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import {
  FlowAIChatMessage,
  FlowAIChatRequest,
  FlowAIChatResponse,
  FlowAICheckResponse,
} from "../types/wire.gen";
import { checkFlowAIPrompt, runFlowAIPrompt } from "./ai";
import { EvalCase, EvalRoute, evalCases } from "./ai.eval.cases";
import { serializeFlow } from "./serialize";

// Runs the prompts of ai.eval.cases.ts against the flow AI and writes a
// report. Start the eval server first, then run with its URL:
//
//   OPENROUTER_API_KEY=... go run ./internal/core/flowai/evalserver
//   FLOW_AI_EVAL_URL=http://localhost:4455 npx vitest run ai.eval
//
// FLOW_AI_EVAL_FILTER only runs the cases whose name contains one of its
// comma separated parts, and
// FLOW_AI_EVAL_CHECK_ONLY=1 only checks the prompts.
const url = process.env.FLOW_AI_EVAL_URL;
const filter = process.env.FLOW_AI_EVAL_FILTER ?? "";
const outDir = process.env.FLOW_AI_EVAL_OUT ?? "eval-results";
const checkOnly = process.env.FLOW_AI_EVAL_CHECK_ONLY === "1";
// Like the service's default max_repairs, which the eval server doesn't know.
const maxRepairs = 2;

// USD per million input, cached input and output tokens.
const prices: Record<string, [number, number, number]> = {
  "gpt-5-mini": [0.25, 0.025, 2],
  "gpt-5-nano": [0.05, 0.005, 0.4],
};

interface EvalInfo {
  model: string;
  tokens: {
    InputTokens: number;
    CachedInputTokens: number;
    OutputTokens: number;
  };
  ms: number;
}

interface CaseResult {
  name: string;
  route: EvalRoute[];
  check?: string;
  checkOk?: boolean;
  checkMessage?: string;
  // For questions that are then built: the answer to the question.
  answer?: string;
  buildPrompt?: string;
  message?: string;
  edited: boolean;
  repairs: number;
  issues: string[];
  missingTypes: string[];
  // IDs in the flow that weren't in the prompt or the flow before.
  inventedIds: string[];
  // The edits of each round, for debugging.
  edits: unknown[][];
  buildOk: boolean;
  cost: number;
  ms: number;
  error?: string;
  flow?: string;
}

function cost(info: EvalInfo) {
  const [input, cached, output] = prices[info.model] ?? [0, 0, 0];
  const t = info.tokens;
  return (
    ((t.InputTokens - t.CachedInputTokens) * input +
      t.CachedInputTokens * cached +
      t.OutputTokens * output) /
    1_000_000
  );
}

async function post<T>(path: string, body: unknown) {
  const res = await fetch(url + path, {
    method: "POST",
    body: JSON.stringify(body),
  });
  return (await res.json()) as
    | { success: true; data: T & { eval: EvalInfo } }
    | { success: false; error: { code: string; message: string; data: {} } };
}

async function runCase(c: EvalCase): Promise<CaseResult> {
  const result: CaseResult = {
    name: c.name,
    route: c.route,
    edited: false,
    repairs: 0,
    issues: [],
    missingTypes: [],
    inventedIds: [],
    edits: [],
    buildOk: false,
    cost: 0,
    ms: 0,
  };
  let flow = c.flow();
  const variables = c.variables ?? [];
  const known = c.prompt + JSON.stringify(flow) + JSON.stringify(variables);

  const check = await checkFlowAIPrompt({
    context: c.context,
    prompt: c.prompt,
    flow,
    variables,
    send: async (req) => {
      const res = await post<FlowAICheckResponse>("/check", req);
      if (res.success) result.cost += cost(res.data.eval);
      return res;
    },
  });
  if (check) {
    result.check = check.verdict;
    result.checkMessage = [
      check.message,
      check.suggested_prompt,
      ...check.fields.map((f) => `[${f.type}] ${f.label}`),
    ]
      .filter(Boolean)
      .join(" | ");
    // The check can only tell prompts to send from ones to clarify.
    const expected: string[] = c.route.map((r) =>
      r === "clarify" ? "clarify" : "send"
    );
    result.checkOk = expected.includes(check.verdict);
  }

  if (checkOnly) return result;

  const start = Date.now();
  const prompt = async (messages: FlowAIChatMessage[]) => {
    let rounds = 0;
    const res = await runFlowAIPrompt({
      context: c.context,
      messages,
      variables,
      getFlow: () => flow,
      applyFlow: (applied) => {
        flow = { nodes: applied.nodes, edges: applied.edges };
        result.edited = true;
      },
      send: async (req: FlowAIChatRequest) => {
        if (rounds++ > maxRepairs) {
          return {
            success: false,
            error: { code: "repair_limit", message: "", data: {} },
          };
        }
        const res = await post<FlowAIChatResponse>("/chat", req);
        if (res.success) {
          result.cost += cost(res.data.eval);
          result.edits.push(res.data.edits);
        }
        return res;
      },
    });
    result.message = res.message;
    result.buildPrompt = res.buildPrompt;
    result.repairs += res.repairs;
    result.issues = res.issues;
    return res;
  };

  let answeredFirst = false;
  try {
    const messages: FlowAIChatMessage[] = [{ role: "user", content: c.prompt }];
    const res = await prompt(messages);
    answeredFirst = !result.edited;

    // Like clicking "Build this".
    if (c.thenBuild && answeredFirst && res.buildPrompt) {
      result.answer = res.message;
      await prompt([
        ...messages,
        { role: "assistant", content: res.message },
        { role: "user", content: res.buildPrompt },
      ]);
    }
  } catch (err) {
    result.error = (err as Error).message;
  }
  result.ms = Date.now() - start;

  const types = new Set(flow.nodes.map((n) => n.type));
  result.missingTypes = (c.types ?? []).filter(
    (t) => !t.split("|").some((alt) => types.has(alt))
  );
  // Discord IDs, and stored variables, as the eval's app has none.
  for (const node of flow.nodes) {
    const data = JSON.stringify(node.data);
    const ids = [
      ...(data.match(/\b\d{15,20}\b/g) ?? []),
      ...(typeof node.data.variable_id === "string"
        ? [node.data.variable_id]
        : []),
    ];
    result.inventedIds.push(...ids.filter((id) => !known.includes(id)));
  }
  const usesComponent = flow.edges.some((e) =>
    e.sourceHandle?.startsWith("component_")
  );

  const shouldEdit = c.route.includes("build") || !!c.thenBuild;
  const mayEdit = shouldEdit || c.route.includes("clarify");
  result.buildOk =
    !result.error &&
    result.issues.length === 0 &&
    result.missingTypes.length === 0 &&
    result.inventedIds.length === 0 &&
    (!c.componentBranch || usesComponent) &&
    (result.edited ? mayEdit : !shouldEdit || c.route.length > 1) &&
    // Questions are answered before anything is built.
    (!c.thenBuild || (answeredFirst && !!result.answer));
  result.flow = serializeFlow(flow.nodes, flow.edges, c.context);
  return result;
}

function report(results: CaseResult[]) {
  const sum = (f: (r: CaseResult) => number) =>
    results.reduce((a, r) => a + f(r), 0);
  const checked = results.filter((r) => r.checkOk !== undefined);
  const lines = [
    `# Flow AI eval, ${new Date().toISOString()}`,
    "",
    `Builds ok: ${sum((r) => +r.buildOk)}/${results.length}. ` +
      `Checks ok: ${sum((r) => +!!r.checkOk)}/${checked.length}. ` +
      `Repairs: ${sum((r) => r.repairs)}. ` +
      `Cost: $${sum((r) => r.cost).toFixed(4)}, ` +
      `avg ${(sum((r) => r.ms) / results.length / 1000).toFixed(
        1
      )}s per build.`,
    "",
    "| Case | Expected | Check | Build | Repairs | Cost | Time |",
    "| --- | --- | --- | --- | --- | --- | --- |",
    ...results.map(
      (r) =>
        `| ${r.name} | ${r.route.join("/")} | ${r.check ?? "-"}${
          r.checkOk === false ? " ✗" : ""
        } | ${r.buildOk ? "ok" : "✗"}${r.edited ? " (edited)" : ""} | ${
          r.repairs
        } | $${r.cost.toFixed(4)} | ${(r.ms / 1000).toFixed(1)}s |`
    ),
    "",
    ...results.flatMap((r) => [
      `## ${r.name}`,
      "",
      r.checkMessage ? `Check: ${r.check}: ${r.checkMessage}` : "",
      r.answer ? `> ${r.answer.replace(/\n/g, "\n> ")}` : "",
      r.buildPrompt ? `Build this: ${r.buildPrompt}` : "",
      r.message ? `> ${r.message.replace(/\n/g, "\n> ")}` : "",
      r.error ? `Error: ${r.error}` : "",
      r.missingTypes.length ? `Missing: ${r.missingTypes.join(", ")}` : "",
      r.inventedIds.length ? `Invented: ${r.inventedIds.join(", ")}` : "",
      ...r.issues.map((i) => `- ${i}`),
      "",
      "```",
      r.flow ?? "",
      "```",
      ...r.edits.map(
        (edits, i) =>
          `<details><summary>Edits of round ${
            i + 1
          }</summary>\n\n\`\`\`json\n${JSON.stringify(
            edits,
            null,
            1
          )}\n\`\`\`\n</details>`
      ),
      "",
    ]),
  ];
  return lines.filter((l, i, a) => l !== "" || a[i - 1] !== "").join("\n");
}

describe.skipIf(!url)("flow AI eval", () => {
  it(
    "runs the cases",
    async () => {
      const cases = evalCases.filter((c) =>
        filter.split(",").some((f) => c.name.includes(f))
      );
      const results: CaseResult[] = [];
      // A few at a time, to stay within rate limits.
      for (let i = 0; i < cases.length; i += 6) {
        results.push(
          ...(await Promise.all(cases.slice(i, i + 6).map(runCase)))
        );
      }

      mkdirSync(outDir, { recursive: true });
      const file = `${outDir}/flow-ai-${Date.now()}.md`;
      writeFileSync(file, report(results));
      console.log(report(results).split("\n").slice(0, 3).join("\n"));
      console.log(`Report: ${file}`);
      expect(results).toHaveLength(cases.length);
    },
    60 * 60 * 1000
  );
});
