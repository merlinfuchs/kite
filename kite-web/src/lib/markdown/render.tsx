import { ReactNode } from "react";
import { SingleASTNode } from "simple-markdown";

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
    case "paragraph":
      return (
        <div key={key} className="paragraph">
          {children()}
        </div>
      );
    case "strong":
      return <strong key={key}>{children()}</strong>;
    case "inlineCode":
      return <code key={key}>{node.content}</code>;
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
    default:
      return Array.isArray(node.content) ? (
        <span key={key}>{children()}</span>
      ) : typeof node.content === "string" ? (
        node.content
      ) : null;
  }
}
