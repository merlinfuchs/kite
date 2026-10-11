import { Edge, Node } from "@xyflow/react";
import dynamic from "next/dynamic";
import DiscordReplyMockup from "./DiscordReplyMockup";

// The flow needs the browser to measure itself.
const FlowExample = dynamic(() => import("../flow/FlowExample"), {
  ssr: false,
});

export interface Showcase {
  flow: { nodes: Node[]; edges: Edge[] };
  // The slash command or the event that the bot replies to.
  command?: string;
  event?: string;
  reply: string;
}

// A flow with the bot's reply in Discord overlapping its corner.
export default function ShowcaseVisual({ showcase }: { showcase: Showcase }) {
  return (
    <div>
      {/* On large screens the reply overlaps the bottom of this box by 3.5rem, so the flow keeps out of it. */}
      <div className="h-[380px] rounded-lg border bg-muted/50 p-4 lg:pb-[4.5rem]">
        <FlowExample nodes={showcase.flow.nodes} edges={showcase.flow.edges} />
      </div>
      <div className="relative mt-4 lg:-mt-14 lg:ml-auto lg:w-[80%] lg:shadow-xl rounded-lg">
        <DiscordReplyMockup
          command={showcase.command}
          event={showcase.event}
          content={showcase.reply}
        />
      </div>
    </div>
  );
}
