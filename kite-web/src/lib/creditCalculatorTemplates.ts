import { CalculatorTemplate } from "./creditCalculator";
import { aiModelTiers } from "./flow/aiModels";
import { getFlowCredits } from "./flow/credits";
import { getTemplates } from "./flow/templateFiles";
import { Template, templateFlows } from "./flow/templates";

// getCalculatorTemplates runs at build time, so the templates and blocks
// aren't part of the pricing page's bundle. That's why it isn't in
// creditCalculator.ts, which the calculator imports.
export function getCalculatorTemplates(): CalculatorTemplate[] {
  // The templates of the site all have a usage.
  return getTemplates().flatMap(({ usage, ...template }) =>
    usage
      ? [
          {
            id: template.id,
            name: template.name,
            unit: usage.unit,
            unitOne: usage.unit_one,
            perThousandMembers: usage.per_thousand_members,
            credits: Object.fromEntries(
              aiModelTiers.map((t) => [
                t.value,
                templateCredits({ ...template, usage }, t.value),
              ])
            ),
          },
        ]
      : []
  );
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
