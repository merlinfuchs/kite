import env from "@/lib/env/client";
import { getFlowCredits } from "@/lib/flow/credits";
import { discordEventType } from "@/lib/flow/eventTypes";
import {
  componentLabel,
  Template,
  TemplateFlow,
  templateFlows,
} from "@/lib/flow/templates";
import { cn, formatNumber } from "@/lib/utils";
import Link from "next/link";
import { useMemo } from "react";
import { Card } from "../ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";

// Runs a month the credits of each flow are shown for. Small screens only
// have room for the middle one.
const volumes = [100, 500, 2000];
const volumeClassName = (i: number) => (i === 1 ? "" : "hidden sm:table-cell");

// TemplateCreditUsage shows what each flow of a template costs per run and
// at a few levels of activity, so people know before adding it.
export default function TemplateCreditUsage({
  template,
}: {
  template: Template;
}) {
  const rows = useMemo(
    () =>
      templateFlows(template).map((f) => ({
        ...describeFlow(f),
        key: f.key,
        credits: getFlowCredits(f.flow.nodes, f.flow.edges).max,
      })),
    [template]
  );

  return (
    <section>
      <h2 className="font-semibold mb-1">Credit usage</h2>
      <p className="text-sm text-muted-foreground mb-4 max-w-2xl">
        The most credits each part costs when it runs. Reading data, conditions
        and loops are free, and the free plan includes 10,000 credits a month.{" "}
        <Link
          href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
          target="_blank"
          className="underline hover:text-foreground"
        >
          About credits
        </Link>
      </p>
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Part</TableHead>
              <TableHead className="text-right whitespace-nowrap">
                Per run
              </TableHead>
              {volumes.map((v, i) => (
                <TableHead
                  key={v}
                  className={cn(
                    "text-right whitespace-nowrap",
                    volumeClassName(i)
                  )}
                >
                  {formatNumber(v)} runs
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.key}>
                <TableCell>
                  <div className="font-medium">{row.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {row.trigger}
                  </div>
                </TableCell>
                <TableCell className="text-right">{row.credits}</TableCell>
                {volumes.map((v, i) => (
                  <TableCell
                    key={v}
                    className={cn(
                      "text-right text-muted-foreground",
                      volumeClassName(i)
                    )}
                  >
                    {formatNumber(row.credits * v)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </section>
  );
}

function describeFlow(f: TemplateFlow) {
  switch (f.kind) {
    case "command":
      return { name: `/${f.name}`, trigger: "Someone uses the command" };
    case "event": {
      const type = discordEventType(f.eventType);
      return {
        name: type?.label ?? f.eventType,
        trigger: type?.description ?? "The event happens",
      };
    }
    case "button":
      return {
        name: `${componentLabel(f.message.data, f.flowSourceId) ?? "A"} button`,
        trigger: "Someone clicks the button",
      };
  }
}
