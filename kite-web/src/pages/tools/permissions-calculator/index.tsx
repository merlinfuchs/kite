import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import PermissionsCalculator from "@/tools/permissions-calculator/components/PermissionsCalculator";

const faqs = [
  {
    question: "Should my bot ask for Administrator?",
    answer:
      "Usually not. Administrator gives the bot every permission, so a bug or a leaked token can do a lot of damage. Many server admins won't add a bot that asks for it.",
  },
  {
    question: "Can I change the permissions after adding the bot?",
    answer:
      "Yes. Server admins can edit the role of the bot in the server settings at any time, or give it permissions in single channels.",
  },
  {
    question: "Why is the number so long?",
    answer:
      "Discord has more than 50 permissions, and each one is a bit of the number. Permissions with a high bit, like Use External Apps, make the number very large.",
  },
];

export default function PermissionsCalculatorPage() {
  return (
    <ToolLayout
      toolId="permissions-calculator"
      title="Discord Permissions Calculator"
      description="Select the permissions your bot needs to get the permissions integer and an invite link for your bot."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>What Is the Permissions Integer?</h2>
          <p>
            Discord stores permissions as a single number. Each permission is
            one bit of that number, and the integer above adds up the
            permissions you select. Invite links pass it in the{" "}
            <code>permissions</code> parameter, so the role of your bot gets
            exactly these permissions when someone adds it to a server.
          </p>
          <h2>Which Permissions Does My Bot Need?</h2>
          <p>
            Only ask for the permissions your bot uses. A bot that only answers
            slash commands needs no permissions at all, because the replies go
            through the command. A moderation bot needs Kick Members, Ban
            Members and Moderate Members. Server admins are more likely to add a
            bot that asks for little.
          </p>
          <h2>How Do I Use the Invite Link?</h2>
          <p>
            Open the link, choose a server where you have the Manage Server
            permission and click on Authorize. The bot joins with a role that
            has the permissions you selected.
          </p>
        </ToolArticle>
      }
    >
      <PermissionsCalculator />
    </ToolLayout>
  );
}
