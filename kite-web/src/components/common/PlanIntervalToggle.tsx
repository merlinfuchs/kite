import { BillingInterval } from "@/lib/billing";
import { BillingPlan } from "@/lib/types/wire.gen";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";

export default function PlanIntervalToggle({
  plans,
  value,
  onChange,
}: {
  plans: BillingPlan[];
  value: BillingInterval;
  onChange: (value: BillingInterval) => void;
}) {
  // The smallest saving of any plan, so the badge holds for every plan.
  const saving = Math.min(
    ...plans
      .filter((plan) => plan.price > 0)
      // Rounded first, so float error can't turn a 20% saving into 19%.
      .map((plan) =>
        Math.floor(
          Math.round((1 - plan.yearly_price / (plan.price * 12)) * 1e4) / 100
        )
      )
  );

  return (
    <Tabs
      value={value}
      onValueChange={(v) => onChange(v as BillingInterval)}
      className="flex justify-center"
    >
      <TabsList>
        <TabsTrigger value="monthly" className="px-4">
          Monthly
        </TabsTrigger>
        <TabsTrigger value="yearly" className="px-4 gap-2">
          Yearly
          {saving > 0 && (
            <span className="text-xs font-semibold text-primary">
              Save {saving}%
            </span>
          )}
        </TabsTrigger>
      </TabsList>
    </Tabs>
  );
}
