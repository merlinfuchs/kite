import {
  DiscordBold,
  DiscordItalic,
  DiscordLink,
  DiscordUnderlined,
} from "@skyra/discord-components-react";
import Markdown, { Components } from "react-markdown";

const components: Partial<Components> = {
  strong({ children }) {
    return <DiscordBold>{children}</DiscordBold>;
  },
  em({ children }) {
    return <DiscordItalic>{children}</DiscordItalic>;
  },
  u({ children }) {
    return <DiscordUnderlined>{children}</DiscordUnderlined>;
  },
  a({ children, href }) {
    return <DiscordLink href={href}></DiscordLink>;
  },
};

export default function MessageMarkdown({ children }: { children: string }) {
  return <Markdown components={components}>{children}</Markdown>;
}
