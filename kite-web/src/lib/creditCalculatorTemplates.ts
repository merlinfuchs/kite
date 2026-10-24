import { calculatorActivities, CalculatorTemplate } from "./creditCalculator";
import { aiModelTiers } from "./flow/aiModels";
import { getFlowCredits } from "./flow/credits";
import { FlowData } from "./flow/dataSchema";
import { getTemplates, Template } from "./flow/templates";

// getCalculatorTemplates runs at build time, so the templates and blocks
// aren't part of the pricing page's bundle. That's why it isn't in
// creditCalculator.ts, which the calculator imports.
export function getCalculatorTemplates(): CalculatorTemplate[] {
  return getTemplates().map((template) => {
    const activity = calculatorActivities[template.id];
    if (!activity) {
      throw new Error(`template ${template.id} has no calculator activity`);
    }

    const usesAi = templateFlows(template).some((flow) =>
      flow.nodes.some((n) => n.data.ai_chat_completion_data)
    );

    return {
      id: template.id,
      name: template.name,
      ...activity,
      credits: templateCredits(template),
      aiCredits: usesAi
        ? Object.fromEntries(
            aiModelTiers.map((t) => [
              t.value,
              templateCredits(template, t.value),
            ])
          )
        : null,
    };
  });
}

function templateFlows(template: Template): FlowData[] {
  return [
    ...template.commands.map((c) => c.flowSource({})),
    ...template.eventListeners.map((l) => l.flowSource({})),
    ...template.messages.flatMap((m) => Object.values(m.flowSources({}))),
  ];
}

function templateCredits(template: Template, aiModel?: string) {
  return Math.max(
    0,
    ...templateFlows(template).map((flow) => {
      const nodes = flow.nodes.map((n) =>
        aiModel && n.data.ai_chat_completion_data
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
