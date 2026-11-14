import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import UserLookup from "@/tools/user-lookup/components/UserLookup";
import Link from "next/link";

const faqs = [
  {
    question: "Why do I need to log in?",
    answer:
      "Discord doesn't allow showing user profiles to anyone without an account. Logging in with Discord is free and takes a few seconds.",
  },
  {
    question: "Can I look up a user by their username?",
    answer:
      "No. Discord only lets bots look up users by their ID, so you need the ID first.",
  },
  {
    question: "Can I look up someone I don't share a server with?",
    answer:
      "Yes. The lookup works for any Discord account, as long as you have its ID.",
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
      title="Discord user lookup by ID"
      description="Look up any Discord user by their ID to see their avatar, banner, badges, server tag, and when they made their account."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>What the user lookup shows</h2>
          <p>
            You get the public profile of any Discord user, even if you
            don&apos;t share a server with them. That&apos;s their display name,
            username, avatar, banner, and the badges on their profile. It also
            shows their server tag and avatar decoration if they have one, with
            links to the full-size images.
          </p>
          <p>
            The account&apos;s creation date comes from the ID itself. If
            that&apos;s all you need, the{" "}
            <Link href="/tools/snowflake-decoder">snowflake decoder</Link> shows
            it without logging in.
          </p>
          <h2>How to find someone&apos;s Discord user ID</h2>
          <ol>
            <li>Open User Settings and go to Advanced.</li>
            <li>Turn on Developer Mode.</li>
            <li>
              Right-click the user, or tap their profile on your phone, and
              click on Copy User ID.
            </li>
          </ol>
          <p>
            You can also paste a mention like{" "}
            <code>&lt;@1083038298513887292&gt;</code> or a link to their
            profile. The lookup finds the ID in it.
          </p>
          <h2>Where the data comes from</h2>
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
