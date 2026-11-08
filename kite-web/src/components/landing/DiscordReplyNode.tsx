import { Handle, NodeProps, Position } from "@xyflow/react";
import DiscordReplyMockup from "./DiscordReplyMockup";

export interface DiscordReplyNodeData extends Record<string, unknown> {
  command?: string;
  event?: string;
  content: string;
}

// The bot's reply in Discord as the last node of an example flow, so it
// scales with the blocks above it.
export default function DiscordReplyNode({ data }: NodeProps) {
  const reply = data as DiscordReplyNodeData;

  return (
    <div className="w-[380px] shadow-md rounded-lg">
      <Handle
        type="target"
        position={Position.Top}
        isConnectable={false}
        className="!opacity-0"
      />
      <DiscordReplyMockup
        command={reply.command}
        event={reply.event}
        content={reply.content}
      />
    </div>
  );
}
