import { blockDefinitions } from "./blocks";
import { discordEventTypes } from "./flow/eventTypes";
import { templatePath } from "./flow/templates";
import { BillingPlan, Bundle, StatsGetResponse } from "./types/wire.gen";

export interface LlmsTxtData {
  stats: StatsGetResponse | null;
  plans: BillingPlan[] | null;
  bundles: Bundle[] | null;
  docsUrl: string;
  githubUrl: string;
  // YYYY-MM-DD, the day the numbers are from.
  today: string;
}

const site = "https://kite.onl";

const number = (n: number) => n.toLocaleString("en-US");

const price = (n: number) => `$${n.toFixed(2)}`;

// What each plan feature is called on the pricing page.
const planFeatures: [keyof BillingPlan, string][] = [
  ["feature_rotating_status", "rotating status"],
  ["feature_animated_profile", "animated avatar and banner"],
  ["feature_voice_audio", "audio in voice channels"],
  ["feature_server_profiles", "a profile per server"],
  ["feature_personas", "the Send message as block"],
  ["feature_priority_support", "priority support"],
];

function planLimits(plan: BillingPlan) {
  return [
    `${number(plan.feature_max_commands)} commands`,
    `${number(plan.feature_max_event_listeners)} event listeners`,
    `${number(plan.feature_max_variables)} stored variables`,
    `${number(plan.feature_usage_credits_per_month)} credits a month`,
    `up to ${number(plan.feature_max_guilds)} servers per bot`,
    `${number(plan.feature_max_collaborators)} collaborators`,
    `${number(plan.feature_max_ai_prompts_per_month)} Flow AI prompts a month`,
  ].join(", ");
}

// llmsTxt is /llms.txt, a summary of Kite for AI assistants with the numbers of
// the API, like the homepage shows them.
export function llmsTxt(data: LlmsTxtData) {
  const { stats, docsUrl, githubUrl, today } = data;
  const plans = (data.plans ?? [])
    .filter((p) => !p.hidden)
    .sort((a, b) => a.price - b.price);
  const free = plans.find((p) => p.default);
  const paid = plans.filter((p) => !p.default);
  const bundles = data.bundles ?? [];
  const blocks = blockDefinitions.filter((b) => b.category);

  const lines: string[] = [
    "# Kite",
    "",
    "> Kite is a free Discord bot maker. You build slash commands and event listeners by connecting blocks in a visual editor, and Kite hosts the bot 24/7. Kite is open source under AGPL-3.0.",
    "",
    "Each bot is your own Discord application, with your own name, avatar, and status. Kite started in 2024.",
  ];

  if (stats) {
    lines.push(
      "",
      `As of ${today}, ${number(stats.users)} users have made ${number(
        stats.apps
      )} bots with ${number(
        stats.commands + stats.event_listeners
      )} commands and event listeners. They ran ${number(
        stats.executions_last_month
      )} flows in the last month.`
    );
  }

  lines.push(
    "",
    "## Free plan",
    "",
    "- No credit card and no time limit.",
    "- Online 24/7, whether the bot is used or not. Kite never turns a bot off for inactivity.",
    "- Up to 10 bots per account, each with its own plan and limits.",
    '- Shows "Powered by Kite.onl" as its status until you set your own, which is free.'
  );
  if (free) lines.push(`- ${capitalize(planLimits(free))}.`);

  if (paid.length > 0) {
    lines.push(
      "",
      "## Paid plans",
      "",
      "Prices are per bot, in USD, and include tax. Yearly costs 10 times the monthly price. You can cancel any time.",
      ""
    );
    for (const plan of paid) {
      const features = planFeatures
        .filter(([key]) => plan[key] === true)
        .map(([, name]) => name);
      lines.push(
        `- Kite ${plan.title}: ${price(plan.price)} a month or ${price(
          plan.yearly_price
        )} a year. ${capitalize(planLimits(plan))}.${
          features.length > 0 ? ` Includes ${list(features)}.` : ""
        }`
      );
    }
  }

  lines.push(
    "",
    "## What you can build",
    "",
    `- ${blocks.length} blocks for messages, members, roles, channels, threads, voice, stored variables, AI, and requests to other websites. Conditions and loops are blocks too, and cost no credits.`,
    `- Slash commands with arguments, permissions, and cooldowns, plus buttons, select menus, and modals.`,
    `- Event listeners for ${discordEventTypes.length} Discord events: ${list(
      discordEventTypes.map((t) => t.label.toLowerCase())
    )}. Also scheduled tasks and webhooks from other websites.`,
    "- Stored variables keep points, counts, and settings for each member or server between runs.",
    "- Leveling and economy systems are built with stored variables. The Marketplace has no template for them right now.",
    "- Flow AI, the AI assistant in the editor, builds and changes commands from a description.",
    "- AI blocks: Ask AI, Search the Web, AI Check, AI Choice, AI Score, and AI Translate.",
    "- Integrations with ER:LC, Roblox, and Cookie API.",
    `- An MCP server at ${site}/api/mcp, to build and change bots from Claude, ChatGPT, Codex, Cursor, or VS Code.`
  );

  if (bundles.length > 0) {
    const official = bundles.filter((b) => b.official).length;
    lines.push(
      "",
      "## Templates",
      "",
      `The Marketplace has ${
        bundles.length
      } templates, ${official} made by Kite and ${
        bundles.length - official
      } from the community. These are added the most:`,
      ""
    );
    for (const b of [...bundles]
      .sort((a, b) => b.install_count - a.install_count)
      .slice(0, 10)) {
      const description = b.listing?.description ?? b.description;
      lines.push(
        `- [${b.listing?.name ?? b.name}](${site}${templatePath(b.id)}): ${
          /[.!?]$/.test(description) ? description : `${description}.`
        } ${b.official ? "Made by Kite" : "From the community"}, added ${number(
          b.install_count
        )} times.`
      );
    }
  }

  lines.push(
    "",
    "## Pages",
    "",
    `- [Pricing](${site}/pricing): Every plan and what each block costs in credits.`,
    `- [Marketplace](${site}/marketplace): Templates to add to a bot.`,
    `- [Free Discord bot hosting](${site}/discord-bot-hosting): How Kite hosts bots 24/7.`,
    `- [AI Discord bot maker](${site}/ai-discord-bot-maker): Building a bot with Flow AI.`,
    `- [Discord MCP server](${site}/discord-mcp-server): Building a bot from Claude, ChatGPT, and other AI apps.`,
    `- [How to make a Discord bot](${site}/guides/how-to-make-a-discord-bot): A step by step guide.`,
    `- [Kite compared to other bot makers](${site}/alternatives): BotGhost, inventor.bot, BotWiz, Bot Designer for Discord, SCNX, FromFlow, and VibeBot, with their free limits and the date they were checked.`,
    `- [Documentation](${docsUrl}): Every block, event, and placeholder.`,
    `- [Source code](${githubUrl}): Kite on GitHub.`,
    ""
  );

  return lines.join("\n");
}

function capitalize(text: string) {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

function list(items: string[]) {
  if (items.length < 3) return items.join(" and ");
  return `${items.slice(0, -1).join(", ")}, and ${items[items.length - 1]}`;
}
