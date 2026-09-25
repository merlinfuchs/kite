import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { optionColor } from "@/lib/flow/nodes";
import { EventTypeScheduleCron } from "@/lib/types/flow.gen";

export default function FlowNodeEntryEvent(props: NodeProps) {
  const isSchedule = props.data.event_type === EventTypeScheduleCron;
  const eventName = props.data.event_type?.split("_").join(" ") || "";

  return (
    <FlowNodeBase
      {...props}
      title={isSchedule ? "Run on schedule" : `Listen for ${eventName}`}
      description={
        isSchedule
          ? `Runs the flow on the schedule ${
              props.data.event_schedule_cron || ""
            } (UTC). Drop different actions here!`
          : `Listens for ${eventName} events to trigger the flow. Drop different actions here!`
      }
      highlight={true}
      showConnectedMarker={false}
    >
      {!isSchedule && (
        <FlowNodeHandle
          type="target"
          position={Position.Top}
          color={optionColor}
        />
      )}
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
