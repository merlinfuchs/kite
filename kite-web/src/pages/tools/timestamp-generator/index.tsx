import { faqJsonLd } from "@/lib/seo";
import ToolArticle from "@/tools/common/components/ToolArticle";
import ToolLayout from "@/tools/common/components/ToolLayout";
import TimestampGenerator from "@/tools/timestamp-generator/components/TimestampGenerator";

const faqs = [
  {
    question: "Why does my friend see a different time?",
    answer:
      "That's how timestamps work. Discord shows each reader the time in their own time zone, so everyone sees the right local time for the same moment.",
  },
  {
    question: "Can bots use timestamps?",
    answer:
      "Yes. A bot can put the same codes in its messages and in the description and fields of embeds.",
  },
  {
    question: "Does the relative timestamp update?",
    answer:
      'Yes. A relative timestamp like "in 2 hours" counts down while the message is open and changes to "2 hours ago" afterwards.',
  },
];

export default function TimestampGeneratorPage() {
  return (
    <ToolLayout
      toolId="timestamp-generator"
      title="Discord Timestamp Generator"
      description="Pick a date and time and copy the code into your message. Everyone sees the time in their own time zone."
      jsonLd={[faqJsonLd(faqs)]}
      article={
        <ToolArticle faqs={faqs}>
          <h2>How Do Discord Timestamps Work?</h2>
          <p>
            A timestamp code like <code>&lt;t:1791763200:F&gt;</code> contains a
            moment in time as a Unix timestamp and a letter for the format. When
            you send it in a message, Discord shows the date and time in the
            time zone and language of each reader. This is useful for events
            with people from different countries.
          </p>
          <h2>Which Format Should I Use?</h2>
          <p>
            Use the long date and time for events, and the relative format for
            deadlines, like &quot;in 3 days&quot;. You can put several
            timestamps in one message, for example the date and the relative
            time next to each other.
          </p>
        </ToolArticle>
      }
    >
      <TimestampGenerator />
    </ToolLayout>
  );
}
