import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import InviteResolver from "@/tools/invite-resolver/components/InviteResolver";

const faqs = [
  {
    question: "Can I look up a server by its ID instead?",
    answer:
      "Only through an invite. Discord doesn't show a server to people outside it unless they have an invite.",
  },
  {
    question: "Does looking up an invite make me join the server?",
    answer:
      "No. The invite resolver only reads the invite. It doesn't use it and the server doesn't see it.",
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
      title="Discord invite resolver"
      description="Paste a Discord invite to see the server behind it, with its ID, member count, and boosts, without joining."
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
        </ToolArticle>
      }
    >
      <InviteResolver />
    </ToolLayout>
  );
}
