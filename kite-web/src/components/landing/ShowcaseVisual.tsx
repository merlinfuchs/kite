import { useIsMobile } from "@/lib/hooks/use-mobile";
import { cn } from "@/lib/utils";
import { Edge, Node } from "@xyflow/react";
import dynamic from "next/dynamic";
import { DiscordReplyNodeData } from "./DiscordReplyNode";

// The flow needs the browser to measure itself. Loading it lazily also keeps
// every block out of the page bundle, so the page hydrates without waiting
// for them.
const ShowcaseFlow = dynamic(() => import("./ShowcaseFlow"), {
  ssr: false,
});

export interface Showcase {
  flow: { nodes: Node[]; edges: Edge[] };
  // The slash command or the event that the bot replies to.
  command?: string;
  event?: string;
  reply: string;
}

// A flow that ends in the bot's reply in Discord.
export default function ShowcaseVisual({
  showcase,
  hero,
}: {
  showcase: Showcase;
  // For the homepage hero: bigger, and without a frame, so the flow floats on
  // the glow behind it.
  hero?: boolean;
}) {
  // On phones, argument blocks next to each other make the flow too wide to
  // read, so it only shows the steps.
  const isMobile = useIsMobile();
  const flowNodes = isMobile
    ? showcase.flow.nodes.filter((n) => n.type !== "option_command_argument")
    : showcase.flow.nodes;

  // The reply goes below the last block, which is the one nothing follows.
  const last = flowNodes
    .filter((n) => !showcase.flow.edges.some((e) => e.source === n.id))
    .reduce((a, b) => (b.position.y > a.position.y ? b : a));
  const replyData: DiscordReplyNodeData = {
    command: showcase.command,
    event: showcase.event,
    content: showcase.reply,
  };
  const nodes: Node[] = [
    ...flowNodes,
    {
      // Not "reply", which example flows use for their response block.
      id: "discord-reply",
      type: "discord_reply",
      position: { x: last.position.x - 30, y: last.position.y + 130 },
      data: replyData,
    },
  ];
  const edges: Edge[] = [
    ...showcase.flow.edges.filter(
      (e) =>
        nodes.some((n) => n.id === e.source) &&
        nodes.some((n) => n.id === e.target)
    ),
    {
      id: "discord-reply",
      source: last.id,
      target: "discord-reply",
      // The animation repaints the flow on every frame, which makes taps on
      // the page slow on phones.
      animated: !isMobile,
      style: isMobile ? { strokeDasharray: 5 } : undefined,
    },
  ];

  return (
    <div
      className={cn(
        "p-4",
        hero
          ? "h-[420px] lg:h-[700px]"
          : "h-[440px] rounded-lg border bg-muted/50"
      )}
    >
      {/* Remounting fits the view again when the nodes change. */}
      <ShowcaseFlow
        key={isMobile ? "mobile" : "desktop"}
        nodes={nodes}
        edges={edges}
      />
    </div>
  );
}
