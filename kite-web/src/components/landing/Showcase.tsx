import { Edge, Node } from "@xyflow/react";
import dynamic from "next/dynamic";

// Both render on the client only: the flow needs the browser to measure
// itself and the reply shows the current time.
const FlowExample = dynamic(() => import("../flow/FlowExample"), {
  ssr: false,
});
const DiscordReplyMockup = dynamic(() => import("./DiscordReplyMockup"), {
  ssr: false,
});

export interface ShowcaseContent {
  heading: string;
  body: string;
  flow: { nodes: Node[]; edges: Edge[] };
  trigger: string;
  reply: string;
}

export default function Showcase({ content }: { content: ShowcaseContent }) {
  return (
    <section className="container py-12 md:py-16">
      <h2 className="text-3xl md:text-4xl font-bold text-center">
        {content.heading}
      </h2>
      <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
        {content.body}
      </p>
      <div className="grid lg:grid-cols-2 gap-6 mt-10 items-center">
        <div className="h-[420px] rounded-lg border bg-muted/50">
          <FlowExample nodes={content.flow.nodes} edges={content.flow.edges} />
        </div>
        <DiscordReplyMockup trigger={content.trigger} content={content.reply} />
      </div>
    </section>
  );
}
