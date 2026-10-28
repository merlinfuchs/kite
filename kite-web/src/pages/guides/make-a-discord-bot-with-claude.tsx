import AIClientGuide, { GuideProse } from "@/components/landing/AIClientGuide";
import ConnectionMockup from "@/components/landing/ConnectionMockup";
import ConnectorDialogMockup from "@/components/landing/ConnectorDialogMockup";
import DiscordReplyMockup from "@/components/landing/DiscordReplyMockup";
import MCPSessionMockup from "@/components/landing/MCPSessionMockup";
import env from "@/lib/env/client";
import Link from "next/link";

const connectorLink =
  "https://claude.ai/customize/connectors?modal=add-custom-connector&connectorName=Kite&connectorUrl=https%3A%2F%2Fkite.onl%2Fapi%2Fmcp";

export default function MakeADiscordBotWithClaudePage() {
  return (
    <AIClientGuide
      content={{
        client: "Claude",
        title: "How to Make a Discord Bot With Claude (No Code)",
        description:
          "Make a Discord bot with Claude for free, without writing code. Connect Claude to Kite, describe your commands, and Claude builds them while Kite hosts the bot 24/7.",
        heading: "How to Make a Discord Bot With Claude",
        intro:
          "Claude can build your Discord bot for you. Connect it to Kite, describe what the bot should do, and Claude creates the commands and events. Kite runs the bot 24/7, so there's no code to host and nothing to deploy.",
        steps: [
          {
            title: "Add Kite to Claude",
            visual: (
              <ConnectorDialogMockup
                title="Add custom connector"
                button="Add"
              />
            ),
            body: (
              <>
                <p>
                  Kite is a connector for Claude. It gives Claude the tools to
                  read and build your bot.
                </p>
                <ol>
                  <li>
                    Open{" "}
                    <a href={connectorLink} target="_blank" rel="noreferrer">
                      Add Kite to Claude
                    </a>
                    . It opens Claude&apos;s{" "}
                    <strong>Add custom connector</strong> dialog with Kite
                    filled in.
                  </li>
                  <li>
                    Click on <strong>Add</strong>.
                  </li>
                </ol>
                <p>
                  You can also open <strong>Customize</strong>, then{" "}
                  <strong>Connectors</strong> in claude.ai or the Claude app,
                  click on <strong>+</strong>, choose{" "}
                  <strong>Add custom connector</strong> and paste{" "}
                  <code>https://kite.onl/api/mcp</code>.
                </p>
                <p>
                  The free plan of Claude can add one custom connector, so Kite
                  works without paying for Claude.
                </p>
              </>
            ),
          },
          {
            title: "Sign In and Pick Its Access",
            visual: <ConnectionMockup client="Claude" origin="claude.ai" />,
            body: (
              <>
                <p>
                  Claude opens Kite so you can sign in. Then you choose what
                  Claude can do:
                </p>
                <ul>
                  <li>
                    <strong>Full access</strong> to build and change your bot,
                    or <strong>Read only</strong> to only look at it.
                  </li>
                  <li>
                    Whether it can send requests to Discord as your bot, like
                    creating a channel.
                  </li>
                  <li>Which of your apps it can use.</li>
                </ul>
                <p>
                  Click on <strong>Connect</strong>, and Claude is ready. You
                  can disconnect it anytime under API tokens in Kite.
                </p>
              </>
            ),
          },
          {
            title: "Describe Your Bot",
            visual: (
              <MCPSessionMockup
                session={{
                  prompt:
                    "Add a /ban command with a user and a reason, and reply with who was banned.",
                  calls: [
                    { tool: "list_apps", result: "Found My Bot" },
                    { tool: "get_blocks", result: "Looked up 3 blocks" },
                    { tool: "create_command", result: "Created /ban" },
                  ],
                  answer:
                    "Added /ban to My Bot. It bans the user, keeps the reason in the audit log and replies in the channel.",
                }}
              />
            ),
            body: (
              <>
                <p>
                  Start a new chat and tell Claude what your bot should do. Name
                  the app if you have more than one.
                </p>
                <p>
                  Claude looks up how each block works, builds the flow and
                  saves it. Kite checks the flow first, so Claude can&apos;t
                  save one that&apos;s broken.
                </p>
                <p>
                  Claude asks you before it uses a tool of Kite. Allow it once,
                  or for good so it stops asking.
                </p>
              </>
            ),
          },
          {
            title: "Try It in Discord",
            visual: (
              <DiscordReplyMockup
                command="ban"
                content="Banned @Sam: Spamming invites"
              />
            ),
            body: (
              <>
                <p>
                  Type <code>/ban</code> in your server. A new command can take
                  up to a minute to show up.
                </p>
                <p>
                  The command is now part of your bot, like one you built
                  yourself. Open it in Kite&apos;s editor to see the blocks
                  Claude used, or ask Claude to change it.
                </p>
              </>
            ),
          },
        ],
        prompts: [
          "When someone joins, give them the Member role and welcome them in #welcome.",
          "Add a /poll command with a question and up to four options, and buttons to vote.",
          "Why does my /ticket command fail? Check the logs and fix it.",
          "Add the verification template and post its message in #verify.",
        ],
        faqs: [
          {
            question: "Can I make a Discord bot with Claude for free?",
            answer:
              "Yes. The free plan of Claude can add one custom connector, which is enough for Kite. Kite hosts the bot for free, and the free plan has 10,000 credits a month for running its commands.",
          },
          {
            question: "Does Claude write code for my bot?",
            answer:
              "No. Claude builds flows in Kite, which are made of blocks like Ban member or Send message. You can open them in Kite's editor and change them yourself, without reading code.",
          },
          {
            question: "Does my bot stay online when I close Claude?",
            answer:
              "Yes. Kite runs the bot 24/7. Claude only builds and changes it.",
          },
          {
            question: "Is it safe to let Claude change my bot?",
            answer:
              "You choose what Claude can do when you connect it, and Claude asks before it uses a tool. Every change is kept in the history of the command, so you can restore an earlier version.",
          },
          {
            question: "Does it use my Claude usage limits?",
            answer:
              "Yes. Claude uses its own model to build your bot, so the chat counts towards your Claude limits like any other. It doesn't use Kite's AI prompts or credits.",
          },
        ],
        children: (
          <GuideProse>
            <h2>Using Claude Code Instead</h2>
            <p>
              Claude Code can build your bot from the terminal too. Add Kite
              with this command, then run <code>/mcp</code> in Claude Code and
              pick <code>kite</code> to sign in.
            </p>
            <pre>
              <code>
                claude mcp add --transport http kite https://kite.onl/api/mcp
              </code>
            </pre>
            <h2>Why Not Have Claude Write a discord.js Bot?</h2>
            <p>
              Claude can write a bot with discord.js or discord.py. Then you
              need a server that keeps it running, a place to store the token,
              and every change goes through the code. If the bot crashes at
              night, it stays offline until you notice.
            </p>
            <p>
              With Kite, Claude builds flows that Kite runs, and anyone on your
              team can open them in the editor. See{" "}
              <Link href="/discord-mcp-server">Kite&apos;s MCP server</Link> for
              everything Claude can do, or the{" "}
              <a href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/mcp`}>
                MCP documentation
              </a>{" "}
              for other clients.
            </p>
          </GuideProse>
        ),
      }}
    />
  );
}
