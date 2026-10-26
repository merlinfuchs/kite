import SimpleMarkdown from "simple-markdown";
import { memo, useMemo } from "react";
import { addListBreaks, renderMarkdown } from "@/lib/markdown/render";

// The AI's answers only use paragraphs, lists, bold and inline code.
const { newline, paragraph, list, strong, inlineCode, escape, text } =
  SimpleMarkdown.defaultRules;
const rules = { newline, paragraph, list, strong, inlineCode, escape, text };
const parse = SimpleMarkdown.parserFor(rules);

export default memo(function FlowAIMarkdown({ text }: { text: string }) {
  const content = useMemo(() => {
    // Blocks only end with a blank line.
    const blocks = addListBreaks(text.trim());
    return renderMarkdown(parse(blocks + "\n\n", { inline: false }));
  }, [text]);

  return (
    <div className="space-y-2 [&_.paragraph]:whitespace-pre-line [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:list-disc [&_ul]:pl-5 [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs">
      {content}
    </div>
  );
});
