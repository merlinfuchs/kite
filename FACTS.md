# Kite facts sheet

```
Product: Kite
One-line description: A free Discord bot maker. You build commands and event listeners by connecting blocks in a visual editor, and Kite hosts the bot 24/7.
Who it's for: Discord server owners and admins who want their own bot without writing or hosting code. Also ER:LC and Roblox communities, and people who build bots from Claude or ChatGPT.
Reader's technical level: Mostly non-technical. They know Discord's own words, like server, channel, role, slash command, permission, embed, button, and Developer Portal. Explain or avoid API, HTTP request, webhook, JSON, expression, and bot token.
Situations that bring people here, from Google Search Console queries and OpenPanel landing pages and referrers for 2026-10-03 to 2026-11-01, and the commands and event listeners of apps created in the 30 days to 2026-11-02. The counts include what templates add.
- Wanting a custom Discord bot of your own for free, without coding. Most non-brand searches are "discord bot maker", "free discord bot maker", "discord bot creator", "custom discord bot", and "discord bot maker free no coding".
- Looking for free 24/7 hosting for a bot. /discord-bot-hosting is the marketing page people land on most after the homepage, from searches like "free discord bot hosting 24/7".
- Moderation commands. /ban, /kick, /unban, and /mute are the most created commands, in about 1,200 apps each.
- An AI chatbot for the server. /ask is in 920 apps, and /ai-discord-bot gets the most visits of the landing pages after /discord-bot-hosting.
- Welcoming new members. 989 apps have a listener for members joining, and 286 for members leaving.
- Server utilities like /userinfo, /serverinfo, /avatar, and /say, in about 250 apps each.
- Giveaways, modmail, and tickets. /giveaway is in 240 apps, /reply in 252, and ticket commands like /close and /claim in about 100.
- Connecting an ER:LC private server. /status, /players, and /announce are in about 150 apps.
- Fun commands like /hug, /kiss, and /coinflip, in about 50 to 65 apps each.
- Leaving another bot maker, mostly BotGhost. /alternatives/botghost is the most visited comparison page.
- Asking ChatGPT how to make a bot. ChatGPT is the third biggest referrer after Google and Discord, with 1,916 sessions.
Main alternatives:
- Other no-code bot makers. BotGhost by far the most, from searches like "botghost alternative", then BotWiz and Bot Designer for Discord. Also inventor.bot (formerly inventor.gg), SCNX, FromFlow, and VibeBot. Each has a comparison page under /alternatives, with its limits and the date they were checked.
- Writing a bot in code and hosting it yourself. Some searches for free hosting ask for Python, like "free discord bot hosting python".
Why pick this one: The free plan keeps your own bot online 24/7 on up to 100 servers, and you can replace Kite's status with your own for free. Kite is open source under AGPL-3.0. Don't claim it's the only or first bot maker to do anything.

Features (feature | what it means for the user | free or paid):
- Visual flow editor | build a command by connecting blocks, and see every step | free
- Slash commands, buttons, select menus, and modals | members use the bot like any other Discord bot | free
- Event listeners for messages, new members, reactions, boosts, and more | the bot reacts without a command | free
- Scheduled tasks | run a flow on a schedule | free, every 5 minutes at most, every minute on Kite Premium
- Webhooks | start a flow from another website or service | free
- Message templates | design embeds, buttons, and select menus and see them as Discord shows them | free
- Stored variables | keep points, counts, and settings per member or server, up to 64 KB per value | free
- Marketplace templates | add a ready-made bot, like moderation or giveaways, then change it | free
- Share and list your own templates | anyone with the link can add it, and listed ones show in the Marketplace | free
- Flow AI, opened with the Ask AI button in the editor | describe a command and it adds and connects the blocks | free, 2 prompts a month, more on paid plans
- AI blocks (Ask AI, Search the Web, AI Check, AI Choice, AI Score, AI Translate) | the bot answers questions and sorts messages | free, paid in credits
- HTTP Request and Discord API Request blocks | call any API | free
- Run JavaScript block | a short piece of JavaScript for logic that would take many blocks, up to 4,000 characters and 25 ms, can read from Discord, ER:LC, and Cookie API, use stored variables, and send requests to other websites | free, 10 credits a run, plus 1 for each read from ER:LC or Cookie API, 1 for each change or list of a stored variable, and 3 for each request to another website
- Integrations: ER:LC, Roblox, and Cookie API | ready blocks for ER:LC private servers, Roblox users and groups, and Cookie API | free, ER:LC needs your server key
- Collaborators | others edit the same bot with you | free, 3 including you, more on paid plans
- MCP server at https://kite.onl/api/mcp | build and change the bot from Claude, ChatGPT, Codex, Cursor, or VS Code | free
- Custom status | your own status instead of "Powered by Kite.onl" | free
- Rotating status and the Set status block | up to 10 statuses that change every minute | Kite Plus and up
- Animated avatar and banner | GIF or animated WebP profile | Kite Plus and up
- Play audio in voice channels | upload a sound and the bot plays it | Kite Plus and up
- Profile per server | another nickname, avatar, banner, About Me, and name style in each server. The name style is a font, an effect, and colors. Servers show the font. The profile also shows the effect and colors | Kite Premium and up
- Send message as | one message with another name and avatar, like the user who ran the command | Kite Premium and up
- Priority support | Kite Premium and up

Free plan (called Basic in the API and on the plan cards):
- 15 commands, 10 event listeners, 25 stored variables, and 25 message templates.
- 10,000 credits per month.
- Up to 100 servers per bot.
- 3 collaborators including you, 10 secrets, 5 scheduled tasks every 5 minutes at most, and 5 webhooks.
- 2 Flow AI prompts per month.
- Online 24/7, whether the bot is used or not. Kite never turns a bot off for inactivity.
- Shows "🪁 Powered by Kite.onl" as status until you set your own. Never say "no branding". Say the status can be replaced for free.
- No time limit and no credit card.
- Up to 10 apps per account, each with its own plan, credits, and limits. Set by max_apps_per_user in default.toml, which prod doesn't override.

Paid plans (monthly, yearly), read from https://api.kite.onl/v1/billing/plans on 2026-11-02. That endpoint is the source of truth. Plans in kite-service/internal/config/default.toml don't reach production, which uses /opt/kite/kite.toml:
- Plus: $2.99, $29.99. 50 commands, 15 event listeners, 50 variables, 50 message templates. 30,000 credits. 250 servers. 5 collaborators. 10 Flow AI prompts. Rotating status, animated profile, and voice audio.
- Premium: $5.99, $59.99. 100 commands, 25 event listeners, 100 variables, 100 message templates. 100,000 credits. 1,000 servers. 10 collaborators. 25 Flow AI prompts. 10 scheduled tasks every minute at most, and 10 webhooks. Adds profile per server, Send message as, and priority support. Marked popular.
- Ultimate: $14.99, $149.99. 100 commands, 50 event listeners, 200 variables, 200 message templates. 1,000,000 credits. 2,500 servers. 25 collaborators. 100 Flow AI prompts. 15 scheduled tasks and 15 webhooks.
- Yearly costs 10 times the monthly price, so 2 months are free.
- A plan belongs to one app, and prices are per app and month. Your other apps stay on the free plan.
- Prices are in USD and include tax. Payment goes through Lemon Squeezy as merchant of record.
- A donation link sits under the plans on the app's Plans page.
Trial, refunds, cancellation:
- No trial. The free plan has no time limit.
- Full refund of the first payment of a subscription within 14 days, once per app, by email to contact@kite.onl. No refunds of renewals or plan changes. Payments charged by mistake, like twice, are always refunded.
- Cancel any time. The app keeps its plan until the end of the paid period.
- Source: kite-web/src/pages/refund.tsx and section 7 of kite-web/src/pages/terms.tsx.

Credits:
- Most actions, like sending a message or banning a member, cost 1 credit.
- Conditions, loops, and getting servers, channels, members, messages, and stored variables are free.
- AI blocks, HTTP requests, and the Run JavaScript block cost more. The pricing page lists every cost, generated from the block definitions.
- Being online costs nothing. Flow AI prompts don't use credits.
- Out of credits: the bot stays online, its commands and events pause, and members who use it are told why. It warns at 80%, DMs the owner at 80% and 100%, and runs again when the month ends or a few minutes after an upgrade.

Stats (number | what it counts | source | date checked):
- 72,685 | bots made | GET https://api.kite.onl/v1/stats apps, shown floored on the homepage | 2026-11-02
- 79,314 | users | same endpoint | 2026-11-02
- 198,504 | commands and event listeners | same endpoint | 2026-11-02
- 23,098,351 | flow runs last month | same endpoint | 2026-11-02
- 108 | Marketplace templates, 19 by Kite and 89 from the community | GET https://api.kite.onl/v1/bundles | 2026-11-05
- 334 | GitHub stars | github.com/merlinfuchs/kite | 2026-11-02
- 2024 | year Kite started | first commit 2024-01-10 | 2026-11-02
- Most added templates: Moderation 482, Ask AI 432, Server Info 260, Welcomer 257, Giveaways 235 | install_count from /v1/bundles | 2026-11-02

Main page action: "Get started for free", to /apps. Marketplace pages lead with adding a template. /pricing leads with the plan cards.

Exact names of things users see:
- Dashboard sidebar: Dashboard, Upgrade (Plan on a paid plan), Settings, Logs, Servers, Emojis, Commands, Event listeners, Scheduled tasks, Webhooks, Message templates, Marketplace, Your templates, Stored variables, Secrets, Integrations.
- Editor: flow, block, and the Ask AI button, which opens Flow AI. Ask AI is also the name of a block and a template.
- Block names as in kite-web/src/lib/blocks titles, like Send message, Ban member, Ask AI, Search the Web, AI Check, HTTP Request, Edit server profile, Send message as, Set status.
- Integrations: Discord, ER:LC, Roblox, Cookie API.
- Badges: Template creator (10, 100 or 1,000 users), Early creator, Early supporter, Kite Plus/Premium/Ultimate subscriber, Contributor and Staff on creator pages and in the account menu. Staff pick on templates.
- Plans: Basic, Plus, Premium, and Ultimate. Basic is the free plan.
- The plan page in the dashboard is titled Plans, and its sidebar link says Upgrade, or Plan on a paid plan. There's no Premium page.
- Pages: /pricing, /marketplace, /alternatives, /guides, /tools, /refund. Docs at docs.kite.onl. The dashboard is kite.onl/apps.
- Discord permissions by Discord's names: Administrator, Manage Server, Manage Roles, Manage Webhooks, Ban Members.

Words users use, from Google Search Console queries for 2026-10-03 to 2026-11-01:
- "discord bot maker" by far the most, then "bot creator", "bot builder", and "bot generator".
- "custom discord bot", "make your own discord bot", and "create a discord bot".
- "free", "no coding", "without coding", and "24/7", often together, like "free discord bot hosting 24/7 no coding".
- "ai discord bot maker" for the AI.
- "kite discord", "kite bot", and "kite.onl" for Kite itself.
Words to never use:
- "guild" for a Discord server. The dashboard page is Servers.
- "bundle" or "bundle install". Say template.
- "plugin" for anything new. Plugins are hidden for new bots.
- "no branding".
- "Premium" for every paid plan. Say "paid plans".
- "Kite's AI" or "AI builder" as a name. The name is Flow AI.
- Plan IDs or setting keys, like basic, feature_max_guilds, or temporary_name.
- "entitlement", "execution", "operation", "node", or "handle". Say plan, run, step, block, or output.

Project-specific rules:
- "app" in the dashboard, the docs, and anything about building or managing it, because that's what Discord calls it. "bot" or "Discord bot" on the marketing site, in page titles, meta descriptions, headings, and FAQ answers, because that's what people search for and ask AI about. Where a page needs both, connect them once: "Each app is your own Discord bot."
- The AI in the editor is Flow AI. On the marketing site, in titles and meta descriptions, and on its first mention on a page, describe it with the words people search for: "Flow AI, the AI assistant that builds your bot" or "AI Discord bot builder". Say "the Ask AI button" only for the button itself.
- Flow.ai is an unrelated chatbot company. Never use "Flow AI" alone in page titles, meta descriptions, or headings. Pair it with Kite, like "Kite's Flow AI", or lead with the search words, like "AI Discord bot maker".
- The free plan is Basic. Say "the free plan" in sentences and FAQ answers, and "Basic" only where plans are listed by name, like plan cards, tables, and badges. Mention both once where it isn't obvious: "Basic, the free plan".
- Say "Kite Plus", "Kite Premium", and "Kite Ultimate" in sentences. Plain "Plus" only in badges and card headings.
- "Plus and up" or "Premium and up" for features of more than one plan.
- Name a locked feature with the cheapest plan that has it.
- "Online 24/7" and "always online" are fine. Don't add a credits caveat to them.
- Feature words are common nouns, lowercase in sentences: stored variables, message templates, event listeners, scheduled tasks, webhooks, secrets, credits, flows, blocks, and templates. Capitalize only names: Kite and its plans, Flow AI, the Marketplace, integrations like Cookie API and ER:LC, and Discord's own names like Server Members Intent.
- A UI label is written as it appears, in sentence case, and marked as something to click: backticks in the docs, bold on the site. "Open **Stored variables** in the sidebar." Block names work the same way: "add a Send message block".
- "template" for official and community templates alike. "Made by Kite" and "From the community" tell them apart.
- Kite pages say "Discord bot templates", not "server templates", which are copies of a server's channels and roles.
- Every claim about another bot maker goes on its /alternatives page with lastChecked. Say when the other product is the better fit.
- The free limits are typed by hand in pricing.tsx's meta description and on the /alternatives pages. Change them there when the free plan changes.
- Legal pages are written separately and promise only what the law requires. Don't edit them in a copy pass.

Language notes:
- English only. American English.
- Sentence case for headings, page titles, buttons, menu items, and labels. Product, plan, block, and integration names keep their capitals, like Flow AI, Kite Premium, Send message, and ER:LC.
- Use the Oxford comma: "commands, events, and messages". Older copy without it gets fixed when the sentence is edited anyway.
```
