import SimpleMarkdown from "simple-markdown";
import { memo, useMemo } from "react";

// The AI's answers only use paragraphs, lists, bold and inline code.
const { newline, paragraph, list, strong, inlineCode, escape, text } =
  SimpleMarkdown.defaultRules;
const rules = { newline, paragraph, list, strong, inlineCode, escape, text };
const parse = SimpleMarkdown.parserFor(rules);
const output = SimpleMarkdown.outputFor(rules, "react");

export default memo(function FlowAIMarkdown({ text }: { text: string }) {
  const content = useMemo(() => {
    // Lists only start after a blank line, which models often leave out, and
    // blocks only end with one.
    const blocks = text
      .trim()
      .replace(
        /^(?!\s*(?:[-*]|\d+\.)\s)(.+)\n(?=\s*(?:[-*]|\d+\.)\s)/gm,
        "$1\n\n"
      );
    return output(parse(blocks + "\n\n", { inline: false }));
  }, [text]);

  return (
    <div className="space-y-2 [&_.paragraph]:whitespace-pre-line [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:list-disc [&_ul]:pl-5 [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs">
      {content}
    </div>
  );
});
