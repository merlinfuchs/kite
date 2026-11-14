import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import UserLookup from "@/tools/user-lookup/components/UserLookup";

const faqs = [
  {
    question: "Why do I need to log in?",
    answer:
      "Discord doesn't allow showing user profiles to anyone without an account. Logging in with Discord is free and takes a few seconds.",
  },
  {
    question: "How do I get someone's user ID?",
    answer:
      "Turn on Developer Mode in User Settings under Advanced. Then right-click the user and click on Copy User ID.",
  },
  {
    question: "Can I look up a user by their username?",
    answer:
      "No. Discord only lets bots look up users by their ID, so you need the ID first.",
  },
  {
    question: "Does the user find out that I looked them up?",
    answer:
      "No. The lookup only reads their public profile, the same one you see when you click on their name.",
  },
  {
    question: "Why don't I see their About Me or connections?",
    answer:
      "Discord doesn't share them with bots. Only the avatar, banner, names, badges, and server tag are public.",
  },
];

export default function UserLookupPage() {
  return (
    <ToolLayout
      toolId="user-lookup"
      title="Discord user lookup"
      description="Paste the ID of a Discord user to see their avatar, banner, badges, and when they made their account."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>What the user lookup shows</h2>
          <p>
            You get the public profile of any Discord user, even if you
            don&apos;t share a server with them. That&apos;s their display name,
            username, avatar, banner, and the badges on their profile. It also
            shows their server tag and avatar decoration if they have one. The
            account&apos;s creation date comes from the ID itself.
          </p>
          <p>
            Kite asks Discord for the profile with its own bot, because Discord
            only shows profiles to bots. You need to log in with Discord to use
            it. Results are kept for 10 minutes, so a new avatar can take that
            long to show up.
          </p>
        </ToolArticle>
      }
    >
      <UserLookup />
    </ToolLayout>
  );
}
