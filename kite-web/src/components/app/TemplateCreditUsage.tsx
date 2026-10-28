import env from "@/lib/env/client";
import { getFlowCredits } from "@/lib/flow/credits";
import { Template, templateFlows } from "@/lib/flow/templates";
import { formatNumber } from "@/lib/utils";
import { FlowData } from "@/lib/types/flow.gen";
import Link from "next/link";
import { useMemo } from "react";

// The credits of the free plan, which the runs are counted against.
const freeCredits = 10000;

// TemplateCreditUsage sums up what the flows of a template cost per run, so
// people know before adding it. Each item of the template shows its own.
export default function TemplateCreditUsage({
  template,
}: {
  template: Template;
}) {
  const credits = useMemo(
    () =>
      templateFlows(template).map(
        (f) => getFlowCredits(f.flow.nodes, f.flow.edges).max
      ),
    [template]
  );
  if (!credits.length) return null;
  const min = Math.min(...credits);
  const max = Math.max(...credits);

  return (
    <p className="text-sm text-muted-foreground mt-4 max-w-2xl">
      {max === 0
        ? "Its parts are free to run."
        : `Its parts cost ${
            min === max ? "up to" : `${min} to`
          } ${max} credits each time they run.`}{" "}
      Reading data, conditions and loops are free, and the free plan includes{" "}
      {formatNumber(freeCredits)} credits a month.{" "}
      <Link
        href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
        target="_blank"
        className="underline hover:text-foreground"
      >
        About credits
      </Link>
    </p>
  );
}

// FlowCredits is what one flow costs at most each time it runs.
export function FlowCredits({ flow }: { flow: FlowData }) {
  const credits = useMemo(
    () => getFlowCredits(flow.nodes, flow.edges).max,
    [flow]
  );

  return (
    <div
      className="rounded-md border bg-background px-2.5 py-1.5 text-xs font-medium"
      title={
        credits
          ? `About ${formatNumber(
              Math.floor(freeCredits / credits)
            )} runs a month on the free plan`
          : undefined
      }
    >
      {credits === 0
        ? "Free to run"
        : `Up to ${credits} ${credits === 1 ? "credit" : "credits"} per run`}
    </div>
  );
}
