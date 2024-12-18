import { useLogSummary } from "@/lib/hooks/api";
import Link from "next/link";
import { Button } from "../ui/button";

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAppId } from "@/lib/hooks/params";
import { useMemo } from "react";
import { formatNumber } from "@/lib/utils";

export default function LogSummaryCard() {
  const logSummary = useLogSummary();

  const percentages = useMemo(() => {
    if (!logSummary?.total_entries)
      return {
        error: 0,
        warn: 0,
        info: 0,
        debug: 0,
      };

    return {
      error: (logSummary.total_errors / logSummary.total_entries) * 100,
      warn: (logSummary.total_warnings / logSummary.total_entries) * 100,
      info: (logSummary.total_infos / logSummary.total_entries) * 100,
      debug: (logSummary.total_debugs / logSummary.total_entries) * 100,
    };
  }, [logSummary]);

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardDescription>Logs in the last 24 hours</CardDescription>
        <CardTitle className="text-4xl">
          {formatNumber(logSummary?.total_entries)}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-auto">
        <div className="flex flex-col gap-2">
          <div className="bg-red-500/40 rounded-md overflow-hidden flex justify-between relative h-6">
            <div
              className="bg-red-500/80 text-xs font-bold text-white absolute inset-0 flex items-center px-3"
              style={{
                width: `max(65px, ${percentages.error}%)`,
              }}
            >
              ERROR
            </div>
            <div className="flex items-center px-3 text-sm font-bold absolute top-0 bottom-0 right-0">
              <div>{logSummary?.total_errors}</div>
            </div>
          </div>
          <div className="bg-orange-500/40 rounded-md overflow-hidden flex justify-between relative h-6">
            <div
              className="bg-orange-500/80 text-xs font-bold text-white absolute inset-0 flex items-center px-3"
              style={{
                width: `max(65px, ${percentages.warn}%)`,
              }}
            >
              WARN
            </div>
            <div className="flex items-center px-3 text-sm font-bold absolute top-0 bottom-0 right-0">
              <div>{logSummary?.total_warnings}</div>
            </div>
          </div>
          <div className="bg-blue-500/40 rounded-md overflow-hidden flex justify-between relative h-6">
            <div
              className="bg-blue-500/80 text-xs font-bold text-white absolute inset-0 flex items-center px-3"
              style={{
                width: `max(65px, ${percentages.info}%)`,
              }}
            >
              INFO
            </div>
            <div className="flex items-center px-3 text-sm font-bold absolute top-0 bottom-0 right-0">
              <div>{logSummary?.total_infos}</div>
            </div>
          </div>
          <div className="bg-gray-500/40 rounded-md overflow-hidden flex justify-between relative h-6">
            <div
              className="bg-gray-500/80 text-xs font-bold text-white absolute inset-0 flex items-center px-3"
              style={{
                width: `max(65px, ${percentages.debug}%)`,
              }}
            >
              DEBUG
            </div>
            <div className="flex items-center px-3 text-sm font-bold absolute top-0 bottom-0 right-0">
              <div>{logSummary?.total_debugs}</div>
            </div>
          </div>
        </div>
      </CardContent>
      <CardFooter>
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]/logs",
              query: {
                appId: useAppId(),
              },
            }}
          >
            View Logs
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
