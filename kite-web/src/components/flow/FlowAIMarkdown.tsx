import SimpleMarkdown from "simple-markdown";
import { memo, useMemo } from "react";

// The AI's answers only use paragraphs, lists, bold and inline code.
const { newline, paragraph, list, strong, inlineCode, escape, text } =
  SimpleMarkdown.defaultRules;
const rules = { newline, paragraph, list, strong, inlineCode, escape, text };
const parse = SimpleMarkdown.parserFor(rules);
const output = SimpleMarkdown.outputFor(rules, "react");

export default memo(function FlowAIMarkdown({ text }: { text: string }) {
  // Block elements are only parsed at the end of a block.
  const content = useMemo(
    () => output(parse(text.trim() + "\n\n", { inline: false })),
    [text]
  );

  return (
    <div className="space-y-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:list-disc [&_ul]:pl-5 [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs">
      {content}
    </div>
  );
});
