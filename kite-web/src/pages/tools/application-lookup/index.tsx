import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import ApplicationLookup from "@/tools/application-lookup/components/ApplicationLookup";

const faqs = [
  {
    question: "How do I get the ID of a bot?",
    answer:
      "Turn on Developer Mode in User Settings under Advanced. Then right-click the bot and click on Copy User ID. For most bots, the user ID and the application ID are the same.",
  },
  {
    question: "Why does it say there is no app with this ID?",
    answer:
      "Some older bots have a user ID that's different from their application ID. Copy the application ID from the bot's profile or its invite link instead.",
  },
  {
    question: "What are privileged intents?",
    answer:
      "Access to server members, their online status, and the content of messages. Bots in 100 or more servers need Discord's approval for each one.",
  },
];

export default function ApplicationLookupPage() {
  return (
    <ToolLayout
      toolId="application-lookup"
      title="Discord application lookup"
      description="Paste the ID of a Discord bot or app to see its description, invite link, privileged intents, and the permissions it asks for."
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
          <p>
            Check the privileged intents before you add a bot. They show whether
            it can read messages and see the members of your server.
          </p>
        </ToolArticle>
      }
    >
      <ApplicationLookup />
    </ToolLayout>
  );
}
