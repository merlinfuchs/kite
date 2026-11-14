import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import InviteResolver from "@/tools/invite-resolver/components/InviteResolver";
import Link from "next/link";

const faqs = [
  {
    question: "Can I look up a server by its ID instead?",
    answer:
      "Only through an invite. Discord doesn't show a server to people outside it unless they have an invite.",
  },
  {
    question: "Does looking up an invite make me join the server?",
    answer:
      "No. The invite resolver only reads the invite. It doesn't use it, and the server doesn't see it.",
  },
  {
    question: "Do I need to log in?",
    answer:
      "No. Discord shows invites to anyone, so the lookup runs in your browser without an account.",
  },
  {
    question: "Why is the member count different from the one in Discord?",
    answer:
      "Discord only gives an estimate for invites, so it can be slightly off.",
  },
  {
    question: "Why does it say the invite doesn't exist?",
    answer:
      "The invite expired, reached its maximum uses, or was deleted. Ask the server for a new one.",
  },
];

export default function InviteResolverPage() {
  return (
    <ToolLayout
      toolId="invite-resolver"
      title="Discord invite resolver and server lookup"
      description="Paste a Discord invite to see the server behind it before you join, with its ID, member count, boosts, and who made the invite."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>What the invite resolver shows</h2>
          <p>
            You see what Discord shows before you join: the server&apos;s name,
            icon, banner, and description, how many members it has, and how many
            are online. It also shows the server ID and when the server was
            made, its boosts and verification level, the channel the invite
            opens, and who made the invite.
          </p>
          <p>
            It works with links like discord.gg/abc123 and
            discord.com/invite/abc123, and with just the code. Invites to group
            DMs and friend invites work too.
          </p>
          <h2>Check an invite before you join</h2>
          <p>
            Invites in DMs and on other sites don&apos;t always lead where they
            say. The resolver shows the real server name, its size, and whether
            Discord verified it, without joining. A server that claims to be
            official but isn&apos;t verified and has a few dozen members is
            worth a second look.
          </p>
          <h2>Look up a server by its ID</h2>
          <p>
            Discord only shows a server to people outside it through an invite,
            so you need one to look it up. If you only have the ID, the{" "}
            <Link href="/tools/snowflake-decoder">snowflake decoder</Link> tells
            you when the server was made.
          </p>
        </ToolArticle>
      }
    >
      <InviteResolver />
    </ToolLayout>
  );
}
