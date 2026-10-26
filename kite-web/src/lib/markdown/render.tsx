import { ReactNode } from "react";
import SimpleMarkdown, { SingleASTNode } from "simple-markdown";

// Lists only start after a blank line, which models and people writing like
// on Discord often leave out.
export function addListBreaks(text: string) {
  return text.replace(
    /^(?!\s*(?:[-*]|\d+\.)\s)(.+)\n(?=\s*(?:[-*]|\d+\.)\s)/gm,
    "$1\n\n"
  );
}

// simple-markdown's own React output builds elements by hand with the element
// type of React 18, which React 19 refuses to render. This renders its syntax
// tree with JSX instead.
export function renderMarkdown(
  nodes: SingleASTNode[] | SingleASTNode
): ReactNode {
  const list = Array.isArray(nodes) ? nodes : [nodes];
  return list.map((node, i) => renderNode(node, String(i)));
}

function renderNode(node: SingleASTNode, key: string): ReactNode {
  const children = () =>
    Array.isArray(node.content)
      ? node.content.map((child: SingleASTNode, i: number) =>
          renderNode(child, `${key}.${i}`)
        )
      : node.content;

  switch (node.type) {
    case "text":
      return node.content;
    case "newline":
      return null;
    case "br":
      return <br key={key} />;
    case "paragraph":
      return (
        <div key={key} className="paragraph">
          {children()}
        </div>
      );
    case "heading": {
      const Tag = `h${node.level}` as "h1";
      return <Tag key={key}>{children()}</Tag>;
    }
    case "strong":
      return <strong key={key}>{children()}</strong>;
    case "em":
      return <em key={key}>{children()}</em>;
    case "u":
      return <u key={key}>{children()}</u>;
    case "del":
      return <del key={key}>{children()}</del>;
    case "inlineCode":
      return <code key={key}>{node.content}</code>;
    case "codeBlock":
      return (
        <pre key={key}>
          <code>{node.content}</code>
        </pre>
      );
    case "blockQuote":
      return <blockquote key={key}>{children()}</blockquote>;
    case "list": {
      const items = (node.items as SingleASTNode[][]).map((item, i) => (
        <li key={i}>
          {item.map((child, j) => renderNode(child, `${key}.${i}.${j}`))}
        </li>
      ));
      return node.ordered ? (
        <ol key={key} start={node.start ?? undefined}>
          {items}
        </ol>
      ) : (
        <ul key={key}>{items}</ul>
      );
    }
    case "link":
    case "url":
    case "autolink": {
      const href = SimpleMarkdown.sanitizeUrl(node.target);
      if (!href) return <span key={key}>{children()}</span>;
      // nodrag and stopPropagation keep clicks in the flow editor from
      // selecting or dragging the node the link is in.
      return (
        <a
          key={key}
          href={href}
          target="_blank"
          rel="noopener noreferrer nofollow"
          className="nodrag underline underline-offset-2"
          onClick={(e) => e.stopPropagation()}
        >
          {children()}
        </a>
      );
    }
    default:
      return Array.isArray(node.content) ? (
        <span key={key}>{children()}</span>
      ) : typeof node.content === "string" ? (
        node.content
      ) : null;
  }
}
