import { GaugeIcon, ServerIcon, UserIcon } from "lucide-react";
import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useAppStateGuilds, useCreditLimits } from "@/lib/hooks/api";
import CreditLimitDialog from "./CreditLimitDialog";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import ConfirmDialog from "../common/ConfirmDialog";
import { useCreditLimitDeleteMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { CreditLimit } from "@/lib/types/wire.gen";
import { Progress } from "../ui/progress";
import { Badge } from "../ui/badge";

export function periodLabel(period: string) {
  return period === "day" ? "per day" : "per month";
}

export function useTargetName() {
  const guilds = useAppStateGuilds();

  return (scope: string, targetId: string | null) => {
    if (!targetId) {
      return scope === "guild" ? "Every server" : "Every user";
    }
    if (scope === "guild") {
      const guild = guilds?.find((g) => g!.id === targetId);
      if (guild) return guild.name;
      return `Server ${targetId}`;
    }
    return `User ${targetId}`;
  };
}

export default function CreditLimitList() {
  const limits = useCreditLimits();

  const createButton = (
    <CreditLimitDialog>
      <Button>Create limit</Button>
    </CreditLimitDialog>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!limits ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : limits.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no credit limits"
          description="Stop a single server or user from using up all of your app's credits."
          action={createButton}
        />
      ) : (
        <>
          {limits.map((limit) => (
            <CreditLimitListEntry limit={limit!} key={limit!.id} />
          ))}
          <div className="flex">{createButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}

function CreditLimitListEntry({ limit }: { limit: CreditLimit }) {
  const deleteMutation = useCreditLimitDeleteMutation(useAppId(), limit.id);
  const targetName = useTargetName();

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Limit deleted!");
        } else {
          toast.error(
            `Failed to delete limit: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  const Icon = limit.scope === "guild" ? ServerIcon : UserIcon;
  const used = limit.credits_used;
  const reached =
    used !== null && limit.credits !== null && used >= limit.credits;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <Icon className="h-5 w-5 text-muted-foreground" />
          <div>{targetName(limit.scope, limit.target_id)}</div>
          {!limit.target_id && <Badge variant="secondary">Default</Badge>}
          {reached && <Badge variant="destructive">Limit reached</Badge>}
        </CardTitle>
        <CardDescription className="text-sm flex items-center space-x-1">
          <GaugeIcon className="h-4 w-4" />
          <span>
            {limit.credits === null
              ? "No limit, exempt from the default"
              : `${limit.credits.toLocaleString()} credits ${periodLabel(
                  limit.period
                )}`}
          </span>
        </CardDescription>
      </CardHeader>
      {used !== null && limit.credits !== null && (
        <CardContent className="space-y-2">
          <Progress
            className="h-2"
            value={
              limit.credits === 0
                ? 100
                : Math.min(100, (used / limit.credits) * 100)
            }
          />
          <div className="text-xs text-muted-foreground">
            {used.toLocaleString()} of {limit.credits.toLocaleString()} credits
            used {limit.period === "day" ? "today" : "this month"}
          </div>
        </CardContent>
      )}
      <CardFooter className="flex space-x-3">
        <CreditLimitDialog limit={limit}>
          <Button size="sm" variant="outline">
            Edit
          </Button>
        </CreditLimitDialog>
        <ConfirmDialog
          title="Are you sure that you want to delete this limit?"
          description={
            limit.target_id
              ? "The default limit will apply again, if there is one."
              : "Servers or users without a limit of their own won't be limited anymore."
          }
          onConfirm={remove}
        >
          <Button size="sm" variant="ghost">
            Delete
          </Button>
        </ConfirmDialog>
      </CardFooter>
    </Card>
  );
}
