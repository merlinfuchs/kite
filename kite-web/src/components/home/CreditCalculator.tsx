import { aiModelTiers } from "@/lib/flow/aiModels";
import { CalculatorTemplate, suggestedActivity } from "@/lib/creditCalculator";
import { cn, formatNumber } from "@/lib/utils";
import { useState } from "react";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

const freeCredits = 10000;
// Above this share of the free credits a busy month could use them up.
const closeShare = 0.6;

const defaultSelected = ["moderation", "welcomer"];

// CreditCalculator estimates the monthly credits of a bot made of templates,
// from guesses based on the member count that people can change.
export default function CreditCalculator({
  templates,
}: {
  templates: CalculatorTemplate[];
}) {
  const [members, setMembers] = useState(1000);
  const [selected, setSelected] = useState<string[]>(defaultSelected);
  // Activity people typed in, which the member count no longer changes.
  const [activity, setActivity] = useState<Record<string, number>>({});
  const [aiModels, setAiModels] = useState<Record<string, string>>({});
  const [otherActions, setOtherActions] = useState(0);

  const runs = (t: CalculatorTemplate) =>
    activity[t.id] ?? suggestedActivity(t, members);
  const model = (t: CalculatorTemplate) =>
    aiModels[t.id] ?? aiModelTiers[0].value;
  const unitCredits = (t: CalculatorTemplate) => t.credits[model(t)];

  const total =
    templates
      .filter((t) => selected.includes(t.id))
      .reduce((sum, t) => sum + runs(t) * unitCredits(t), 0) + otherActions;
  const share = total / freeCredits;
  const level = share > 1 ? "over" : share > closeShare ? "close" : "within";

  return (
    <Card className="mt-10 text-left overflow-hidden">
      <div className="p-5 md:p-6 space-y-6">
        <div>
          <label
            htmlFor="calculator-members"
            className="font-semibold block mb-1"
          >
            How many members does your server have?
          </label>
          <p className="text-sm text-muted-foreground mb-3">
            Used to guess how often things happen. You can change every guess
            below.
          </p>
          <Input
            id="calculator-members"
            type="number"
            min={0}
            className="max-w-48"
            value={members}
            onChange={(e) => setMembers(nonNegative(e.target.value))}
          />
        </div>

        <div>
          <div className="font-semibold mb-1">What will your bot do?</div>
          <p className="text-sm text-muted-foreground mb-3">
            Pick the templates you&apos;d use. Each shows the most one run
            costs.
          </p>
          <div className="divide-y border rounded-md">
            {templates.map((t) => {
              const checked = selected.includes(t.id);
              return (
                <div key={t.id} className="p-3">
                  <label className="flex items-center gap-3 cursor-pointer">
                    <Checkbox
                      checked={checked}
                      onCheckedChange={(c) =>
                        setSelected((s) =>
                          c ? [...s, t.id] : s.filter((id) => id !== t.id)
                        )
                      }
                    />
                    <span className="font-medium flex-1">{t.name}</span>
                    <span className="text-sm text-muted-foreground text-right">
                      {unitCredits(t)} credit{unitCredits(t) === 1 ? "" : "s"}{" "}
                      per {t.unitOne}
                    </span>
                  </label>
                  {checked && (
                    <div className="flex flex-wrap items-center gap-x-3 gap-y-2 mt-3 pl-7">
                      <Input
                        type="number"
                        min={0}
                        aria-label={`${t.unit} a month`}
                        className="w-28 h-9"
                        value={runs(t)}
                        onChange={(e) =>
                          setActivity((a) => ({
                            ...a,
                            [t.id]: nonNegative(e.target.value),
                          }))
                        }
                      />
                      <span className="text-sm text-muted-foreground">
                        {t.unit} a month
                      </span>
                      {usesAi(t) && (
                        <Select
                          value={model(t)}
                          onValueChange={(v) =>
                            setAiModels((m) => ({ ...m, [t.id]: v }))
                          }
                        >
                          <SelectTrigger
                            className="w-36 h-9"
                            aria-label="AI model"
                          >
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {aiModelTiers.map((tier) => (
                              <SelectItem key={tier.value} value={tier.value}>
                                {tier.label} model
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
            <div className="p-3">
              <div className="font-medium">Your own flows</div>
              <p className="text-sm text-muted-foreground mt-0.5">
                Most actions cost 1 credit. The editor shows what one run of
                each flow costs.
              </p>
              <div className="flex items-center gap-3 mt-3">
                <Input
                  type="number"
                  min={0}
                  aria-label="Other actions a month"
                  className="w-28 h-9"
                  value={otherActions}
                  onChange={(e) => setOtherActions(nonNegative(e.target.value))}
                />
                <span className="text-sm text-muted-foreground">
                  other actions a month
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-muted/50 border-t p-5 md:p-6">
        <div className="text-sm text-muted-foreground">
          Estimated credits a month
        </div>
        <div className="flex items-baseline gap-2 mt-1">
          <span className="text-4xl font-bold">{formatNumber(total)}</span>
          <span className="text-muted-foreground">
            of {formatNumber(freeCredits)} free
          </span>
        </div>
        <div className="h-2.5 rounded-full bg-secondary overflow-hidden mt-4">
          <div
            className={cn(
              "h-full transition-all",
              {
                over: "bg-destructive",
                close: "bg-yellow-500",
                within: "bg-primary",
              }[level]
            )}
            style={{ width: `${Math.min(share, 1) * 100}%` }}
          />
        </div>
        <p className="text-sm mt-3">
          {level === "over" ? (
            <>
              <span className="font-medium">More than the free plan.</span>{" "}
              Premium plans include more credits, see the{" "}
              <a href="#pricing" className="underline">
                plans above
              </a>
              .
            </>
          ) : level === "close" ? (
            <>
              <span className="font-medium">Close to the free limit.</span> A
              busier month than this could use up the credits and stop your bot
              until the next one.
            </>
          ) : (
            <>
              <span className="font-medium">
                Comfortably within the free plan.
              </span>{" "}
              {formatNumber(freeCredits - total)} credits are left for busier
              months and your own flows.
            </>
          )}
        </p>
        <p className="text-xs text-muted-foreground mt-4">
          Uses the most each run can cost. A flow for every message in a busy
          server can cost far more than this, unless an event filter or a
          condition skips the messages it doesn&apos;t need.
        </p>
      </div>
    </Card>
  );
}

// Templates without AI cost the same for every model.
function usesAi(t: CalculatorTemplate) {
  return new Set(Object.values(t.credits)).size > 1;
}

function nonNegative(value: string) {
  return Math.max(0, Number(value));
}
