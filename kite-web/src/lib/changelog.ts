// The changes users notice, newest first, for the changelog page and
// llms.txt. Add an entry when a change ships to kite.onl.
export interface ChangelogEntry {
  // YYYY-MM-DD, the day the change went live.
  date: string;
  title: string;
  description: string;
  // A page with more about the change, like a landing page or template.
  href?: string;
}

export const changelog: ChangelogEntry[] = [
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
    title: "File uploads in modals and copy and paste for blocks",
    description:
      "Modals can ask members for a file, and you can copy and paste blocks in the editor.",
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
];
