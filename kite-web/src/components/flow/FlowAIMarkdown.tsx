import { Fragment } from "react";

const bullet = /^\s*[-*]\s+/;
const numbered = /^\s*\d+[.)]\s+/;

// Renders the AI's answers, which only use paragraphs, lists, bold and inline
// code.
export default function FlowAIMarkdown({ text }: { text: string }) {
  const groups: { kind: "text" | "ul" | "ol"; lines: string[] }[] = [];
  for (const line of text.trim().split("\n")) {
    const kind = bullet.test(line) ? "ul" : numbered.test(line) ? "ol" : "text";
    const last = groups.at(-1);
    // Blank lines end a paragraph.
    if (!line.trim()) {
      if (last) groups.push({ kind: "text", lines: [] });
    } else if (last?.kind === kind) {
      last.lines.push(line);
    } else {
      groups.push({ kind, lines: [line] });
    }
  }

  return (
    <div className="space-y-2">
      {groups
        .filter((g) => g.lines.length > 0)
        .map((group, i) => {
          if (group.kind === "text") {
            return (
              <p key={i} className="whitespace-pre-wrap">
                <Inline text={group.lines.join("\n")} />
              </p>
            );
          }
          const List = group.kind;
          return (
            <List
              key={i}
              className={
                List === "ul"
                  ? "list-disc pl-5 space-y-1"
                  : "list-decimal pl-5 space-y-1"
              }
            >
              {group.lines.map((line, j) => (
                <li key={j}>
                  <Inline
                    text={line.replace(bullet, "").replace(numbered, "")}
                  />
                </li>
              ))}
            </List>
          );
        })}
    </div>
  );
}

function Inline({ text }: { text: string }) {
  return (
    <>
      {text.split(/(`[^`]+`|\*\*[^*]+\*\*)/).map((part, i) =>
        part.startsWith("`") && part.endsWith("`") && part.length > 1 ? (
          <code key={i} className="rounded bg-muted px-1 py-0.5 text-xs">
            {part.slice(1, -1)}
          </code>
        ) : part.startsWith("**") && part.endsWith("**") && part.length > 3 ? (
          <strong key={i}>{part.slice(2, -2)}</strong>
        ) : (
          <Fragment key={i}>{part}</Fragment>
        )
      )}
    </>
  );
}
