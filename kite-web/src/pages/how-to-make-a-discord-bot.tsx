import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import env from "@/lib/env/client";
import { faqJsonLd } from "@/lib/seo";
import Link from "next/link";

const faqs = [
  {
    question: "Can I make a Discord bot for free?",
    answer:
      "Yes. Creating a Discord application is free, and Kite builds and hosts your bot for free. The free plan includes 10,000 credits per month and up to 100 servers.",
  },
  {
    question: "Do I need to know how to code to make a Discord bot?",
    answer:
      "No. With a bot maker like Kite you build commands by connecting blocks. You only need code if you want to write and host the bot yourself.",
  },
  {
    question: "How long does it take to make a Discord bot?",
    answer:
      "Creating the application and your first command takes about 10 minutes. New commands can take up to a minute to show up in Discord.",
  },
  {
    question: "Can I make a Discord bot on my phone?",
    answer:
      "You can create the application and add it to Kite on your phone. The flow editor works best on a computer, because you drag and connect blocks.",
  },
];

export default function HowToMakeADiscordBotPage() {
  return (
    <HomeLayout
      title="How to Make a Discord Bot Without Coding"
      description="Learn how to make your own Discord bot for free without coding. Create a Discord application, add it to Kite and build your first slash command in about 10 minutes."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <article className="prose dark:prose-invert text-foreground mx-auto my-16 max-w-3xl px-5">
        <h1>How to Make a Discord Bot Without Coding</h1>
        <p>
          You can make your own Discord bot for free without writing any code.
          This guide shows you how to create the bot in Discord, add it to Kite
          and build your first slash command. It takes about 10 minutes.
        </p>

        <h2>What You Need</h2>
        <ul>
          <li>A Discord account</li>
          <li>
            A Discord server where you can add bots. Create a new server to test
            your bot if you don&apos;t have one.
          </li>
        </ul>

        <h2>Step 1: Create a Discord Application</h2>
        <p>
          Every Discord bot belongs to an application. The application holds the
          name, avatar and token of your bot.
        </p>
        <ol>
          <li>
            Open the{" "}
            <a
              href="https://discord.com/developers/applications"
              target="_blank"
              rel="noreferrer noopener"
            >
              Discord Developer Portal
            </a>{" "}
            and sign in.
          </li>
          <li>
            Click on <strong>New Application</strong> and give your bot a name.
          </li>
          <li>Upload an avatar for your bot.</li>
          <li>
            Open the <strong>Bot</strong> section on the left.
          </li>
          <li>
            Turn on the <strong>Presence Intent</strong>,{" "}
            <strong>Server Members Intent</strong> and{" "}
            <strong>Message Content Intent</strong>. Without them your bot
            can&apos;t see new members or read messages.
          </li>
          <li>
            Click on <strong>Reset Token</strong> and copy the token.
          </li>
        </ol>
        <p>
          Keep the token secret. Anyone who has it can control your bot. If you
          share it by accident, reset it and use the new one.
        </p>

        <h2>Step 2: Add Your Bot to Kite</h2>
        <p>
          Kite hosts your bot and gives you an editor to build its commands.
        </p>
        <ol>
          <li>
            Open <Link href="/apps">Kite</Link> and sign in with Discord.
          </li>
          <li>
            Click on <strong>Create app</strong>, paste your token and confirm.
          </li>
          <li>
            Click on <strong>Open app</strong>.
          </li>
          <li>
            Click on <strong>Invite app</strong> at the top right and choose
            your server.
          </li>
        </ol>
        <p>
          Your bot now shows up in your server. It&apos;s online, but it
          doesn&apos;t do anything yet.
        </p>

        <h2>Step 3: Build Your First Command</h2>
        <p>
          A slash command is what members type to use your bot, like{" "}
          <code>/hello</code>. In Kite you build each command by connecting
          blocks.
        </p>
        <ol>
          <li>
            Click on the <strong>/</strong> icon in the sidebar and then on{" "}
            <strong>Create command</strong>.
          </li>
          <li>
            Name it <code>hello</code> and give it a short description.
          </li>
          <li>
            Drag a <strong>Create response message</strong> block from the left
            into the editor.
          </li>
          <li>
            Connect the command block to the new block by dragging from the blue
            dot of one block to the other.
          </li>
          <li>
            Click on the response block and enter a message, like{" "}
            <code>Hello {"{{user.mention}}"}!</code>
          </li>
          <li>
            Click on <strong>Save Changes</strong> and then on{" "}
            <strong>Deploy</strong>.
          </li>
        </ol>
        <p>
          Type <code>/hello</code> in your server to try it. A new command can
          take up to a minute to appear. Restart Discord if you don&apos;t see
          it.
        </p>

        <h2>Step 4: Add More Features</h2>
        <p>
          Now you can add more commands, or let your bot react to events like
          new members and messages. Some ideas:
        </p>
        <ul>
          <li>
            Moderation commands like <code>/ban</code> and <code>/kick</code>.
            See <Link href="/discord-moderation-bot">moderation bot</Link>.
          </li>
          <li>
            A welcome message for new members. See{" "}
            <Link href="/discord-welcome-bot">welcome bot</Link>.
          </li>
          <li>
            An <code>/ask</code> command that answers with AI. See{" "}
            <Link href="/ai-discord-bot">AI bot</Link>.
          </li>
          <li>Buttons and select menus that give members roles.</li>
        </ul>
        <p>
          Templates add finished commands for moderation, welcome messages and
          AI to your app in a few clicks. If you get stuck, click on{" "}
          <strong>Ask AI</strong> in the editor and describe what the command
          should do.
        </p>

        <h2>Making a Discord Bot With Code</h2>
        <p>
          You can also write a bot yourself with a library like discord.js for
          JavaScript or discord.py for Python. This gives you full control, but
          you need to know the language, handle errors and keep a server running
          so the bot stays online. For most community servers, a bot maker is
          faster to set up and easier to change.
        </p>

        <h2>Frequently Asked Questions</h2>
        {faqs.map((faq) => (
          <div key={faq.question}>
            <h3>{faq.question}</h3>
            <p>{faq.answer}</p>
          </div>
        ))}

        <p>
          For more details, read the{" "}
          <a href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/getting-started`}>
            getting started guide
          </a>{" "}
          in the documentation.
        </p>
      </article>
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
