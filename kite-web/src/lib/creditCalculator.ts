// What the credit calculator on the pricing page asks for each template: what
// one run is, and a guess of how many there are a month per 1,000 members,
// which people can change.
export const calculatorActivities: Record<
  string,
  { unit: string; unitOne: string; perThousandMembers: number }
> = {
  moderation: {
    unit: "moderation commands",
    unitOne: "command",
    perThousandMembers: 20,
  },
  "ask-ai": { unit: "AI answers", unitOne: "answer", perThousandMembers: 150 },
  welcomer: {
    unit: "members joining or leaving",
    unitOne: "member",
    perThousandMembers: 120,
  },
  verification: {
    unit: "members verifying",
    unitOne: "member",
    perThousandMembers: 60,
  },
  suggestions: {
    unit: "suggestions",
    unitOne: "suggestion",
    perThousandMembers: 20,
  },
  "server-info": {
    unit: "info commands",
    unitOne: "command",
    perThousandMembers: 50,
  },
  modmail: {
    unit: "modmail messages and replies",
    unitOne: "message",
    perThousandMembers: 30,
  },
  erlc: { unit: "ER:LC commands", unitOne: "command", perThousandMembers: 100 },
};

export interface CalculatorTemplate {
  id: string;
  name: string;
  unit: string;
  unitOne: string;
  perThousandMembers: number;
  // The most one run of any of the template's flows costs. Templates with AI
  // have it per model tier instead.
  credits: number;
  aiCredits: Record<string, number> | null;
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
