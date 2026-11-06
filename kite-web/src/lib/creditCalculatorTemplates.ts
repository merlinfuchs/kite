import { CalculatorTemplate } from "./creditCalculator";
import { aiModelTiers } from "./flow/aiModels";
import { getFlowCredits } from "./flow/credits";
import { Template, templateFlows } from "./flow/templates";

// getCalculatorTemplates runs on the server, so the templates and blocks
// aren't part of the pricing page's bundle. That's why it isn't in
// creditCalculator.ts, which the calculator imports.
export function getCalculatorTemplates(
  templates: Template[]
): CalculatorTemplate[] {
  return templates.flatMap((template) => {
    const usage = template.usage;
    if (!usage) return [];
    return [
      {
        id: template.id,
        name: template.name,
        unit: usage.unit,
        unitOne: usage.unit_one,
        perThousandMembers: usage.per_thousand_members,
        credits: Object.fromEntries(
          aiModelTiers.map((t) => [t.value, templateCredits(template, t.value)])
        ),
      },
    ];
  });
}

function templateCredits(template: Template, aiModel: string) {
  return Math.max(
    0,
    ...templateFlows(template).map(({ flow }) => {
      const nodes = flow.nodes.map((n) =>
        n.data.ai_chat_completion_data
          ? {
              ...n,
              data: {
                ...n.data,
                ai_chat_completion_data: {
                  ...n.data.ai_chat_completion_data,
                  model: aiModel,
                },
              },
            }
          : n
      );
      return getFlowCredits(nodes, flow.edges).max;
    })
  );
}
