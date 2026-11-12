import { track } from "@/lib/analytics";
import { useMoreCreditsPlan } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { formatNumber } from "@/lib/utils";
import Link from "next/link";
import { Button } from "../ui/button";
import { UpgradeButton } from "./PlanLimit";

export default function GetCreditsButton({ from }: { from: string }) {
  const appId = useAppId();
  const plan = useMoreCreditsPlan();

  if (plan) {
    return (
      <UpgradeButton plan={plan} from={from} size="sm">
        Get {formatNumber(plan.feature_usage_credits_per_month)} credits with
        Kite {plan.title}
      </UpgradeButton>
    );
  }

  return (
    <Button size="sm" asChild>
      <Link
        href={`/apps/${appId}/premium`}
        onClick={() => track("upsell_clicked", { from })}
      >
        Get more credits
      </Link>
    </Button>
  );
}
