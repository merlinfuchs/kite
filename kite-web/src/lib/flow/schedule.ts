import { CronExpressionParser } from "cron-parser";
import cronstrue from "cronstrue";
import { Node } from "@xyflow/react";
import { getNodeCreditsCost, getNodeValues } from "./nodes";
import { NodeData } from "./dataSchema";

// Mirrors the server, which checks the same number of upcoming runs.
const gapSamples = 100;

export interface SchedulePreview {
  nextRuns: Date[];
  minGapSeconds: number;
  runsPerMonth: number;
}

// getSchedulePreview returns undefined for expressions the preview can't
// parse. The server is the source of truth, it also accepts descriptors like
// @every that aren't supported here.
export function getSchedulePreview(
  cron: string,
  now = new Date()
): SchedulePreview | undefined {
  let runs: Date[];
  try {
    const interval = CronExpressionParser.parse(cron, {
      tz: "UTC",
      currentDate: now,
    });
    runs = interval.take(gapSamples + 1).map((d) => d.toDate());
  } catch {
    return undefined;
  }

  if (runs.length < 2) {
    return { nextRuns: runs, minGapSeconds: 0, runsPerMonth: runs.length };
  }

  let minGap = Infinity;
  for (let i = 1; i < runs.length; i++) {
    minGap = Math.min(minGap, runs[i].getTime() - runs[i - 1].getTime());
  }

  const spanMs = runs[runs.length - 1].getTime() - runs[0].getTime();
  const monthMs = 30 * 24 * 60 * 60 * 1000;

  return {
    nextRuns: runs.slice(0, 3),
    minGapSeconds: minGap / 1000,
    runsPerMonth: Math.round(((runs.length - 1) / spanMs) * monthMs),
  };
}

// describeSchedule returns undefined for expressions cronstrue can't describe,
// like @every.
export function describeSchedule(cron: string): string | undefined {
  try {
    return cronstrue.toString(cron, {
      use24HourTimeFormat: true,
      throwExceptionOnParseError: true,
    });
  } catch {
    return undefined;
  }
}

// getFlowCreditsCost is the cost of one run if every block runs once.
export function getFlowCreditsCost(nodes: Node[]) {
  return nodes.reduce(
    (sum, node) =>
      sum +
      (getNodeCreditsCost(
        getNodeValues(node.type || ""),
        node.data as NodeData
      ) ?? 0),
    0
  );
}
