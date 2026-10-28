import AIClientGuide, { GuideProse } from "@/components/landing/AIClientGuide";
import ConnectionMockup from "@/components/landing/ConnectionMockup";
import ConnectorDialogMockup from "@/components/landing/ConnectorDialogMockup";
import DiscordReplyMockup from "@/components/landing/DiscordReplyMockup";
import MCPSessionMockup from "@/components/landing/MCPSessionMockup";
import env from "@/lib/env/client";
import Link from "next/link";

export default function MakeADiscordBotWithChatGPTPage() {
  return (
    <AIClientGuide
      content={{
        client: "ChatGPT",
        title: "How to Make a Discord Bot With ChatGPT (No Code)",
        description:
          "Use ChatGPT to make a Discord bot without coding. Add Kite as an MCP server, ask ChatGPT for the commands you want, and Kite keeps the bot online for free.",
        heading: "How to Make a Discord Bot With ChatGPT",
        intro:
          "You can ask ChatGPT for a Discord bot and get one that actually runs. Add Kite to ChatGPT, say which commands and events you want, and ChatGPT builds them in Kite. Kite hosts the bot for free, so there's no code to paste into a file and no server to keep running.",
        steps: [
          {
            title: "Add Kite to ChatGPT",
            visual: (
              <ConnectorDialogMockup
                title="Add custom MCP server"
                button="Create"
              />
            ),
            body: (
              <>
                <p>
                  Kite is an MCP server, which gives ChatGPT the tools to read
                  and build your bot.
                </p>
                <ol>
                  <li>
                    Open{" "}
                    <a
                      href="https://chatgpt.com/plugins"
                      target="_blank"
                      rel="noreferrer"
                    >
                      chatgpt.com/plugins
                    </a>
                    , click on <strong>+</strong> and choose{" "}
                    <strong>Add custom MCP server</strong>.
                  </li>
                  <li>
                    Name it <code>Kite</code> and paste{" "}
                    <code>https://kite.onl/api/mcp</code> as the URL.
                  </li>
                  <li>
                    Choose OAuth to sign in, confirm the warning about custom
                    servers and create it.
                  </li>
                </ol>
                <p>
                  OpenAI decides which plans can add their own MCP servers, and
                  workspace admins can turn it off. If you can&apos;t find the
                  option, your plan or workspace might not have it.
                </p>
              </>
            ),
          },
          {
            title: "Sign In and Choose What It Can Do",
            visual: <ConnectionMockup client="ChatGPT" origin="chatgpt.com" />,
            body: (
              <>
                <p>
                  After you create it, ChatGPT sends you to Kite to sign in. On
                  that page you decide how much ChatGPT gets:
                </p>
                <ul>
                  <li>
                    <strong>Full access</strong> lets it build and change your
                    bot. <strong>Read only</strong> lets it look but not touch.
                  </li>
                  <li>
                    Discord API access, for requests like creating a channel or
                    reading the members of a server.
                  </li>
                  <li>
                    The apps it can see. If you have several, pick only the bot
                    you&apos;re working on.
                  </li>
                </ul>
                <p>
                  Click on <strong>Connect</strong> to finish. To cut ChatGPT
                  off later, remove it under API tokens in Kite.
                </p>
              </>
            ),
          },
          {
            title: "Describe Your Bot",
            visual: (
              <MCPSessionMockup
                title="chatgpt · my-bot"
                session={{
                  prompt:
                    "@Kite Add a /suggest command that posts the suggestion in #suggestions with 👍 and 👎 reactions.",
                  calls: [
                    { tool: "list_apps", result: "Found My Bot" },
                    { tool: "list_channels", result: "Found #suggestions" },
                    { tool: "get_blocks", result: "Looked up 4 blocks" },
                    { tool: "create_command", result: "Created /suggest" },
                  ],
                  answer:
                    "Added /suggest to My Bot. It posts each suggestion in #suggestions and adds 👍 and 👎 for voting.",
                }}
              />
            ),
            body: (
              <>
                <p>
                  Start a new chat, type <code>@</code> and pick{" "}
                  <strong>Kite</strong>. Then tell ChatGPT what your bot should
                  do. Name the app if you have more than one.
                </p>
                <p>
                  Before it saves anything, ChatGPT reads how the blocks it
                  needs work. Kite then checks the finished flow and turns down
                  broken ones, so a half-built command never reaches your
                  server.
                </p>
              </>
            ),
          },
          {
            title: "Try It in Discord",
            visual: (
              <DiscordReplyMockup
                command="suggest"
                content="Add a channel for memes 👍 👎"
              />
            ),
            body: (
              <>
                <p>
                  Run <code>/suggest</code> in your server. Discord can take up
                  to a minute to show new commands.
                </p>
                <p>
                  From here on it&apos;s a normal Kite command. Change the text
                  yourself in the editor, or go back to the chat and ask ChatGPT
                  for a cooldown or a log channel.
                </p>
              </>
            ),
          },
        ],
        prompts: [
          "Add a /ban command with a user and a reason, and reply with who was banned.",
          "When someone joins, give them the Member role and welcome them in #welcome.",
          "Why does my /ticket command fail? Check the logs and fix it.",
          "Add the verification template and post its message in #verify.",
        ],
        faqs: [
          {
            question: "Can I make a Discord bot with ChatGPT for free?",
            answer:
              "Yes. Kite hosts the bot for free, and the free plan has 10,000 credits a month for running its commands. Which ChatGPT plans can add their own MCP servers is up to OpenAI, so check that you see Add custom MCP server under Plugins.",
          },
          {
            question: "Does ChatGPT write code for my bot?",
            answer:
              "No. It builds flows out of Kite's blocks, like Ban member or Send message, and Kite runs them. To tweak something, open the command in the editor. There's no code to read.",
          },
          {
            question: "Can I make a bot that answers with ChatGPT?",
            answer:
              "Yes, but that's a different thing: a bot that uses AI to answer your members. Kite's Ask AI block does that without an OpenAI API key. Start with the Ask AI template from the AI Discord bot page.",
          },
          {
            question: "Does my bot stay online when I close ChatGPT?",
            answer:
              "Yes. You only need ChatGPT while you build. The bot itself runs on Kite's servers around the clock.",
          },
          {
            question: "Is it safe to let ChatGPT change my bot?",
            answer:
              "You decide what ChatGPT gets when you sign in, and you can make it read only. If it changes something you didn't want, restore the earlier version from the command's history.",
          },
        ],
        children: (
          <GuideProse>
            <h2>Want a Bot That Talks Like ChatGPT?</h2>
            <p>
              Maybe you&apos;re looking for something else: a bot that answers
              questions in your server like ChatGPT does. You don&apos;t need
              ChatGPT or an OpenAI API key for that. The{" "}
              <Link href="/ai-discord-bot">AI Discord bot</Link> page shows how
              to add an <code>/ask</code> command and a bot that replies when
              it&apos;s mentioned, with its own personality.
            </p>
            <h2>Using Codex Instead</h2>
            <p>
              Codex can build your bot from the terminal too. Add Kite with this
              command. If Codex doesn&apos;t open Kite to sign in, run{" "}
              <code>codex mcp login kite</code>.
            </p>
            <pre>
              <code>codex mcp add kite --url https://kite.onl/api/mcp</code>
            </pre>
            <h2>Why Not Copy Code From ChatGPT?</h2>
            <p>
              ChatGPT can write a bot with discord.js or discord.py. Then you
              need a server that keeps it running, and every change means asking
              for new code and deploying it again. Code from tutorials also
              breaks when the libraries change.
            </p>
            <p>
              With Kite, ChatGPT builds flows that Kite runs and keeps up to
              date. See{" "}
              <Link href="/discord-mcp-server">Kite&apos;s MCP server</Link> for
              everything ChatGPT can do, or the{" "}
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
