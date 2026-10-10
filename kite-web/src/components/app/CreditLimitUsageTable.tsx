import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import { useCreditLimitUsage } from "@/lib/hooks/api";
import CreditLimitDialog from "./CreditLimitDialog";
import { useTargetName } from "./CreditLimitList";

// Shows who used the most credits, so owners can see who to limit.
export default function CreditLimitUsageTable() {
  const [scope, setScope] = useState("guild");
  const [period, setPeriod] = useState("month");

  const entries = useCreditLimitUsage(scope, period);
  const targetName = useTargetName();

  return (
    <Card>
      <CardHeader className="flex flex-col md:flex-row md:items-start md:justify-between gap-4 space-y-0">
        <div className="space-y-1.5">
          <CardTitle className="text-base">Top usage</CardTitle>
          <CardDescription>
            The {scope === "guild" ? "servers" : "users"} that used the most
            credits {period === "day" ? "today" : "this month"}.
          </CardDescription>
        </div>
        <div className="flex gap-2 flex-wrap">
          <Tabs value={scope} onValueChange={setScope}>
            <TabsList>
              <TabsTrigger value="guild">Servers</TabsTrigger>
              <TabsTrigger value="user">Users</TabsTrigger>
            </TabsList>
          </Tabs>
          <Tabs value={period} onValueChange={setPeriod}>
            <TabsList>
              <TabsTrigger value="day">Today</TabsTrigger>
              <TabsTrigger value="month">This month</TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      </CardHeader>
      <CardContent>
        {!entries ? (
          <Skeleton className="h-32" />
        ) : entries.length === 0 ? (
          <div className="text-sm text-muted-foreground py-6 text-center">
            No usage yet.
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{scope === "guild" ? "Server" : "User"}</TableHead>
                <TableHead className="text-right">Credits used</TableHead>
                <TableHead className="text-right">Limit</TableHead>
                <TableHead></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {entries.map((entry) => (
                <TableRow key={entry!.target_id}>
                  <TableCell>
                    <div>{targetName(scope, entry!.target_id)}</div>
                    <div className="text-xs text-muted-foreground">
                      {entry!.target_id}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    {entry!.credits_used.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right">
                    {entry!.credits === null
                      ? "None"
                      : entry!.credits.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right">
                    <CreditLimitDialog
                      defaults={{
                        scope,
                        target_id: entry!.target_id,
                        period,
                      }}
                    >
                      <Button size="sm" variant="outline">
                        Set limit
                      </Button>
                    </CreditLimitDialog>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
