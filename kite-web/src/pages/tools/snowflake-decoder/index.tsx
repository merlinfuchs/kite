import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import SnowflakeDecoder from "@/tools/snowflake-decoder/components/SnowflakeDecoder";

const faqs = [
  {
    question: "How do I copy a Discord ID?",
    answer:
      "Turn on Developer Mode in User Settings under Advanced. Then right-click a user, server, channel or message and click on Copy ID.",
  },
  {
    question: "Can I find out who an ID belongs to?",
    answer:
      "Not from the ID alone. It only contains when it was created. To look up the user or server, you need the Discord API or a bot.",
  },
  {
    question: "What are the worker ID and process ID?",
    answer:
      "They show which of Discord's servers created the ID. They are only useful to Discord and don't tell you anything about the user or message.",
  },
];

export default function SnowflakeDecoderPage() {
  return (
    <ToolLayout
      toolId="snowflake-decoder"
      title="Discord Snowflake Decoder"
      description="Paste the ID of a Discord user, server, channel or message to find out when it was created."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>What Is a Discord Snowflake?</h2>
          <p>
            Every ID on Discord is a snowflake: a number that contains the
            moment it was created. Users, servers, channels, messages and roles
            all have one. That&apos;s why you can tell how old an account or a
            server is from its ID alone.
          </p>
          <h2>How Is the Date Calculated?</h2>
          <p>
            The first 42 bits of the ID count the milliseconds since the start
            of 2015, when Discord started using snowflakes. The decoder adds
            that to January 1, 2015 to get the date. The remaining bits show
            which server created the ID and keep IDs from the same millisecond
            apart.
          </p>
        </ToolArticle>
      }
    >
      <SnowflakeDecoder />
    </ToolLayout>
  );
}
