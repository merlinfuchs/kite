import { ReactNode } from "react";
import SimpleMarkdown, { SingleASTNode } from "simple-markdown";

// simple-markdown's own React output builds elements by hand with the element
// type of React 18, which React 19 refuses to render. This renders its syntax
// tree with JSX instead.

type Link = (props: { href: string; children: ReactNode }) => ReactNode;

export function renderMarkdown(
  nodes: SingleASTNode[] | SingleASTNode,
  options: { link?: Link } = {}
): ReactNode {
  const list = Array.isArray(nodes) ? nodes : [nodes];
  return list.map((node, i) => renderNode(node, String(i), options));
}

function renderNode(
  node: SingleASTNode,
  key: string,
  options: { link?: Link }
): ReactNode {
  const children = () =>
    Array.isArray(node.content)
      ? node.content.map((child: SingleASTNode, i: number) =>
          renderNode(child, `${key}.${i}`, options)
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
      const Tag = `h${Math.min(Math.max(node.level, 1), 6)}` as "h1";
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
          {item.map((child, j) =>
            renderNode(child, `${key}.${i}.${j}`, options)
          )}
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
      const content = children();
      if (!href) return <span key={key}>{content}</span>;
      const Link = options.link;
      return Link ? (
        <Link key={key} href={href}>
          {content}
        </Link>
      ) : (
        <a key={key} href={href} target="_blank" rel="noopener noreferrer">
          {content}
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
