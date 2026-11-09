import env from "@/lib/env/client";

// The changes users notice, newest first, for the changelog page and
// llms.txt. Add an entry when a change ships to kite.onl.
export interface ChangelogEntry {
  // YYYY-MM-DD, the day the change went live.
  date: string;
  title: string;
  description: string;
  // A page with more about the change, like a landing page, a template, or
  // the docs.
  href?: string;
}

export const changelog: ChangelogEntry[] = [
  {
    date: "2026-11-09",
    title:
      "Run JavaScript can read Discord, use stored variables, and send requests",
    description:
      "The code of a Run JavaScript block can read from Discord, ER:LC, and Cookie API, read and change stored variables, and send requests to other websites with fetch. Put {{secrets.NAME}} in a request to send a secret without the code seeing it. The editor completes these calls as you type.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/blocks/actions/action_javascript_run`,
  },
  {
    date: "2026-11-09",
    title: "Append and Prepend add to lists",
    description:
      "Set stored variable in Append or Prepend mode adds the value as one item when the variable holds a list, instead of turning the list into text. Changes from runs at the same time no longer overwrite each other, also for the first value of a variable.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/blocks/actions/action_variable_set`,
  },
  {
    date: "2026-11-08",
    title: "Run JavaScript block",
    description:
      "Run a short piece of JavaScript with values from your flow, for calculations and text that would take many blocks. It costs 10 credits a run.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/blocks/actions/action_javascript_run`,
  },
  {
    date: "2026-11-08",
    title: "Voice event and voice blocks",
    description:
      "The Voice state update event runs when a member joins, leaves, or moves to another voice channel. Edit voice state mutes, deafens, or moves a member. Play soundboard sound plays a soundboard sound in a voice channel, and unlike Play audio it works on the free plan.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/event`,
  },
  {
    date: "2026-11-08",
    title: "Server boost remove event",
    description:
      "Run a flow when a member stops boosting your server, like to take back a booster reward role.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/event`,
  },
  {
    date: "2026-11-08",
    title: "Tags in Edit channel",
    description:
      "The Edit channel block changes the tags of a forum post, like marking a ticket as Resolved.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/blocks/actions/action_channel_edit`,
  },
  {
    date: "2026-11-08",
    title: "Install Kite as an app",
    description:
      "Install the dashboard on your computer or phone from Chrome, Edge, or Safari on iPhone, so it opens in its own window.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/managing-your-app#installing-kite`,
  },
  {
    date: "2026-11-08",
    title: "Duplicate blocks with Ctrl+D",
    description:
      "Press Ctrl+D or Cmd+D to duplicate the selected blocks. When a block needs an integration your app hasn't enabled, you can enable it from the block.",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/flow-editor#keyboard-shortcuts`,
  },
  {
    date: "2026-11-05",
    title: "AI apps can test your flows",
    description:
      "Claude, ChatGPT, and other apps connected to Kite's MCP server can run a flow as a test before you use it. A test shows the messages and changes the flow would make without making them, and doesn't use credits.",
    href: "/discord-mcp-server",
  },
  {
    date: "2026-11-05",
    title: "Leveling template",
    description:
      "Members earn XP for chatting and level up, with level-up messages, a reward role, /rank, and /leaderboard.",
    href: "/marketplace/leveling",
  },
  {
    date: "2026-11-05",
    title: "Economy template",
    description:
      "Members collect coins with /daily and /work, send them with /pay, and compete on /richest.",
    href: "/marketplace/economy",
  },
  {
    date: "2026-11-05",
    title: "List stored variable values block",
    description:
      "Read many values of a stored variable at once, sorted by value or by when they changed. Leaderboards use it.",
  },
  {
    date: "2026-11-05",
    title: "Flow AI fixes more of its own mistakes",
    description:
      "Flow AI repairs common mistakes in the block settings it writes, so fewer prompts end in an error.",
    href: "/ai-discord-bot-maker",
  },
  {
    date: "2026-10-30",
    title: "Kite Plus for $2.99 a month",
    description:
      "Kite Plus sits between the free plan and Kite Premium, with 50 commands, 30,000 credits, and 250 servers. The free plan now has 15 commands.",
    href: "/pricing",
  },
  {
    date: "2026-10-30",
    title: "Play audio in voice channels",
    description:
      "Upload a sound and your bot plays it in a voice channel with the Play audio block. Kite Plus and up.",
  },
  {
    date: "2026-10-30",
    title: "Profile per server and Send message as",
    description:
      "Your bot can have another nickname, avatar, banner, and About Me in each server. The Send message as block sends one message with another name and avatar. Kite Premium and up.",
    href: "/pricing",
  },
  {
    date: "2026-10-30",
    title: "Animated avatar and banner",
    description:
      "Set your bot's avatar and banner in its settings, and use GIFs or animated WebPs on Kite Plus and up.",
  },
  {
    date: "2026-10-30",
    title: "Bots stay online without credits",
    description:
      "A bot that runs out of credits stays online and tells members why its commands paused. You get a direct message at 80% and 100%.",
  },
  {
    date: "2026-10-30",
    title: "Integration blocks get their own tab",
    description:
      "ER:LC, Roblox, and Cookie API blocks have their own tab in the block list, and search finds blocks in every tab.",
  },
  {
    date: "2026-10-29",
    title: "AI Translate and the Flag Translator template",
    description:
      "AI Translate translates text to another language. The Flag Translator template translates a message when someone reacts with a country's flag.",
    href: "/marketplace/flag-translator",
  },
  {
    date: "2026-10-29",
    title: "If block",
    description:
      "The If block runs some blocks when something is true and others when it isn't.",
  },
  {
    date: "2026-10-29",
    title: "Templates show the permissions they need",
    description:
      "Before you add a template, it lists the Discord permissions your bot needs for it and warns about the ones it doesn't have.",
    href: "/marketplace",
  },
  {
    date: "2026-10-28",
    title: "New Marketplace",
    description:
      "The Marketplace has categories, featured templates, and previews that look like Discord. You can star templates and add them to your bot from their page.",
    href: "/marketplace",
  },
  {
    date: "2026-10-28",
    title: "Guides for Claude and ChatGPT",
    description:
      "Step by step guides to make a Discord bot from Claude or ChatGPT with Kite's MCP server.",
    href: "/guides",
  },
  {
    date: "2026-10-27",
    title: "MCP server",
    description:
      "Build and change your bot from Claude, ChatGPT, Codex, Cursor, or VS Code. You sign in with your Kite account, so there's no token to copy.",
    href: "/discord-mcp-server",
  },
  {
    date: "2026-10-27",
    title: "Cookie API integration",
    description:
      "Blocks for transcripts, welcome cards, QR codes, captchas, and Minecraft player checks from Cookie API.",
    href: "/cookie-api",
  },
  {
    date: "2026-10-27",
    title: "Templates from the community",
    description:
      "Share your commands and event listeners as a template and list it in the Marketplace for others to add.",
    href: "/marketplace",
  },
  {
    date: "2026-10-26",
    title: "Save history",
    description:
      "The flow and message editors keep your last 25 saves, so you can go back to an earlier version.",
  },
  {
    date: "2026-10-26",
    title: "Notes and faster editing in the flow editor",
    description:
      "Add notes to a flow, insert a block into a connection, and open a menu with a right click or a long press.",
  },
  {
    date: "2026-10-26",
    title: "Giveaways, Sticky Message, Reminders, Honeypot, and AFK templates",
    description:
      "Five more templates by Kite. Giveaways picks random winners from everyone who clicks Enter.",
    href: "/marketplace/giveaway",
  },
  {
    date: "2026-10-26",
    title: "10 event listeners on the free plan",
    description: "The free plan has 10 event listeners instead of 5.",
    href: "/pricing",
  },
  {
    date: "2026-10-25",
    title: "AI Automod template",
    description:
      "AI checks every message for spam, scams, ads, and insults, deletes them, and logs them for your moderators.",
    href: "/marketplace/ai-automod",
  },
  {
    date: "2026-10-24",
    title: "See your credits before they run out",
    description:
      "The dashboard forecasts your bot's credits for the month, the editor shows what a run costs, and the pricing page has a calculator.",
    href: "/pricing",
  },
  {
    date: "2026-10-24",
    title: "Reading is free",
    description:
      "Getting servers, channels, members, messages, and stored variables no longer costs credits.",
    href: "/pricing",
  },
  {
    date: "2026-10-24",
    title:
      "Counting, Starboard, Suggestions, Server Info, and Verification templates",
    description:
      "Five more templates by Kite. When you add a template, you pick its channels and roles from your server instead of typing IDs.",
    href: "/marketplace",
  },
  {
    date: "2026-10-24",
    title: "Lookup Value, Forward message, and End flow blocks",
    description:
      "Lookup Value finds a value in a table, Forward message forwards a message to another channel, and End flow stops a flow early. Generate Random Value can also pick random items from a list.",
  },
  {
    date: "2026-10-24",
    title: "Keep or undo Flow AI changes",
    description:
      "Review what Flow AI changed in your flow, then keep it or undo it.",
  },
  {
    date: "2026-10-23",
    title: "Audit log event, Edit role, and Delete role",
    description:
      "Run a flow when something shows up in the server's audit log, and edit or delete roles from a flow.",
  },
  {
    date: "2026-10-22",
    title: "AI Check, AI Choice, and AI Score",
    description:
      "Let AI answer a yes or no question, pick one of several options, or rate something, and run other blocks depending on the answer.",
    href: "/ai-discord-bot",
  },
  {
    date: "2026-10-22",
    title: "Leave server and Publish message blocks",
    description:
      "Your bot can leave a server and publish messages in announcement channels.",
  },
  {
    date: "2026-10-21",
    title: "Reply to a message",
    description:
      "Create channel message can reply to another message, and you choose who gets pinged.",
  },
  {
    date: "2026-10-20",
    title: "ER:LC integration and template",
    description:
      "Get your ER:LC private server's players and run commands on it from Discord.",
    href: "/erlc-discord-bot",
  },
  {
    date: "2026-10-20",
    title: "Respond to webhooks",
    description:
      "Answer the website that called your webhook, and check that a request really comes from it.",
  },
  {
    date: "2026-10-18",
    title: "File uploads in modals",
    description:
      "Modals can ask members for a file, and the flow gets it with their other answers.",
  },
  {
    date: "2026-10-16",
    title: "Direct messages, typing, and forum posts",
    description:
      "Run a flow when someone DMs your bot, show the bot as typing, create forum posts, and set slowmode on channels.",
  },
  {
    date: "2026-10-13",
    title: "Commands update on their own",
    description:
      "Changes to your commands reach Discord automatically after you save. New bots get a setup checklist on their dashboard.",
  },
  {
    date: "2026-10-10",
    title: "Server boost event and select menus",
    description:
      "Run a flow when someone boosts the server, and add user, role, and channel select menus to messages.",
  },
  {
    date: "2026-10-09",
    title: "Webhooks, reaction events, and cooldowns",
    description:
      "Start a flow from another website, run one when someone reacts to a message, and limit how often a command can be used.",
  },
  {
    date: "2026-09-29",
    title: "Flow AI in the editor",
    description:
      "Describe a command and Flow AI adds and connects the blocks for you.",
    href: "/ai-discord-bot-maker",
  },
  {
    date: "2026-09-29",
    title: "Discord API Request and Create poll blocks",
    description:
      "Discord API Request does anything Discord allows that has no block yet. Create poll sends a poll to a channel.",
  },
  {
    date: "2026-09-28",
    title: "Events for joining and leaving servers",
    description: "Run a flow when your bot joins or leaves a server.",
  },
  {
    date: "2026-09-25",
    title: "Kite is out of beta",
    description:
      "The open beta that started in September 2024 is over. Your bot shows its plan where it said beta.",
  },
  {
    date: "2026-09-25",
    title: "Scheduled tasks",
    description:
      "Run a flow on a schedule, like every morning at 9. Kite shows when it runs next.",
  },
  {
    date: "2026-09-25",
    title: "Rotating status",
    description:
      "Your bot can rotate through several statuses, and the Set status block changes it from a flow. Paid plans only.",
  },
  {
    date: "2026-09-25",
    title: "Undo, redo, and copy and paste in the flow editor",
    description:
      "Undo and redo your changes, and copy and paste blocks within a flow or into another one.",
  },
  {
    date: "2026-09-24",
    title: "Join and leave voice channels",
    description: "Your bot can join a voice channel and leave it again.",
  },
  {
    date: "2026-09-23",
    title: "Discord's new message layouts",
    description:
      "Message templates can use Discord's newer components, like containers, sections, and image galleries.",
  },
  {
    date: "2026-09-23",
    title: "Share commands and event listeners",
    description:
      "Export a command or event listener and import it into another bot.",
  },
  {
    date: "2026-01-08",
    title: "Kite moves to the AGPL-3.0 license",
    description:
      "Kite's code is open source under the AGPL-3.0. Anyone can read it or host Kite themselves.",
  },
  {
    date: "2025-09-06",
    title: "Event filters and error handling",
    description:
      "Event Filter runs an event listener only for the events you want. Handle Errors runs other blocks when a block fails.",
  },
  {
    date: "2025-09-04",
    title: "Channel and thread blocks",
    description:
      "Create, edit, and delete channels and threads from a flow. Command arguments can have choices, a minimum, and a maximum.",
  },
  {
    date: "2025-07-28",
    title: "A docs page for every block",
    description:
      "The documentation explains each block, its settings, and what it returns.",
  },
  {
    date: "2025-07-22",
    title: "Get members, roles, channels, and Roblox users",
    description:
      "Blocks that get Discord members, roles, and channels by ID, and Get Roblox User for Roblox players.",
  },
  {
    date: "2025-07-20",
    title: "Search the Web",
    description: "Your bot can look things up on the web with AI.",
  },
  {
    date: "2025-02-26",
    title: "Paid plans",
    description:
      "The first paid plan raises the limits of a bot that needs more. Every bot still starts on the free plan.",
    href: "/pricing",
  },
  {
    date: "2025-01-19",
    title: "Templates",
    description:
      "Add ready-made commands to your bot, and answer a few questions while adding them.",
    href: "/marketplace",
  },
  {
    date: "2025-01-14",
    title: "Modals",
    description:
      "Ask members to fill in a form, and continue the flow with their answers.",
  },
  {
    date: "2025-01-08",
    title: "Calculate Value and Generate Random Value",
    description: "Do math in a flow and pick random numbers.",
  },
  {
    date: "2024-12-09",
    title: "Event listeners",
    description:
      "Run a flow when something happens in a server, like a new message or a new member, without a command.",
  },
  {
    date: "2024-12-01",
    title: "Commands in user apps",
    description:
      "Members can add your bot to their own Discord account and use its commands in any server or DM.",
  },
  {
    date: "2024-11-29",
    title: "Ask AI and direct messages",
    description:
      "Your bot can answer with AI and send direct messages to members.",
  },
  {
    date: "2024-10-03",
    title: "Stored variables",
    description: "Keep values like points or counts between runs of a command.",
  },
  {
    date: "2024-09-24",
    title: "Kite opens to everyone",
    description:
      "Kite starts as a free open beta. You build slash commands from blocks, design messages with buttons, and set your bot's own status.",
  },
];
