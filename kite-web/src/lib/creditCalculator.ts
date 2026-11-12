export interface CalculatorTemplate {
  id: string;
  name: string;
  unit: string;
  unitOne: string;
  perThousandMembers: number;
  // The most one run of any of the template's flows costs, per AI model
  // tier. It's the same for each tier when the template doesn't use AI.
  credits: Record<string, number>;
}

// suggestedActivity rounds the guess for a server's member count, so it
// doesn't look more precise than it is.
export function suggestedActivity(
  template: Pick<CalculatorTemplate, "perThousandMembers">,
  members: number
) {
  const runs = (template.perThousandMembers * members) / 1000;
  if (runs < 10) return Math.round(runs);
  const step = 10 ** (Math.floor(Math.log10(runs)) - 1);
  return Math.round(runs / step) * step;
}
