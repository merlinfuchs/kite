import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import ApplicationLookup from "@/tools/application-lookup/components/ApplicationLookup";
import Link from "next/link";

const faqs = [
  {
    question: "Why does it say there is no app with this ID?",
    answer:
      "Some older bots have a user ID that's different from their application ID. Copy the application ID from the bot's invite link instead.",
  },
  {
    question: "Do I need to log in?",
    answer:
      "No. Discord shows these details to anyone, so the lookup runs in your browser without an account.",
  },
  {
    question: "What are privileged intents?",
    answer:
      "Access to server members, their online status, and the content of messages. Bots in 100 or more servers need Discord's approval for each one.",
  },
  {
    question: "Can I look up my own bot made with Kite?",
    answer:
      "Yes. Copy its application ID from the General Information page of your app in the Discord Developer Portal.",
  },
];

export default function ApplicationLookupPage() {
  return (
    <ToolLayout
      toolId="application-lookup"
      title="Discord bot and application lookup"
      description="Look up a Discord bot by its ID to see its description, invite link, privileged intents, and the permissions it asks for before you add it."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>What the application lookup shows</h2>
          <p>
            Every Discord bot belongs to an application. The lookup shows its
            public details: the name, icon, and description, whether it&apos;s
            verified, and whether anyone can add it. You also see the
            permissions and scopes its invite link asks for, and links to its
            terms and privacy policy.
          </p>
          <h2>Check a bot before you add it</h2>
          <p>
            The privileged intents show whether a bot can read messages and see
            the members of your server. The install permissions are the number
            in its invite link. Paste it into the{" "}
            <Link href="/tools/permissions-calculator">
              permissions calculator
            </Link>{" "}
            to see which permissions it asks for, like Administrator or Ban
            Members.
          </p>
          <h2>How to find a bot&apos;s application ID</h2>
          <ol>
            <li>Open User Settings and go to Advanced.</li>
            <li>Turn on Developer Mode.</li>
            <li>Right-click the bot and click on Copy User ID.</li>
          </ol>
          <p>
            For almost every bot, the user ID and the application ID are the
            same. The ID is also the <code>client_id</code> in the bot&apos;s
            invite link.
          </p>
        </ToolArticle>
      }
    >
      <ApplicationLookup />
    </ToolLayout>
  );
}
