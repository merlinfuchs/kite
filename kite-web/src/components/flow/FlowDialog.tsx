import { FlowData, NodeType } from "@/lib/flow/dataSchema";
import { ReactFlowProvider, useReactFlow } from "@xyflow/react";
import { useCallback, useRef } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { LogEntry } from "@/lib/types/wire.gen";

function InnerFlowDialog({
  flowData,
  logs,
  context,
  onChange,
}: {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  onChange: (d: FlowData) => void;
}) {
  const { getNodes, getEdges } = useReactFlow<NodeType>();

  const handleChange = useCallback(() => {
    onChange({
      nodes: getNodes(),
      edges: getEdges(),
    });
  }, [getNodes, getEdges, onChange]);

  return (
    <Flow
      flowData={flowData}
      logs={logs}
      context={context}
      onChange={handleChange}
    />
  );
}

export default function FlowDialog({
  children,
  onClose,
  flowData,
  logs,
  context,
}: {
  flowData: FlowData;
  logs?: LogEntry[];
  onClose: (data: FlowData) => void;
  context: FlowContextType;
  children: React.ReactNode;
}) {
  const dataRef = useRef(flowData);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        onClose(dataRef.current);
      }
    },
    [onClose]
  );

  const onChange = useCallback((data: FlowData) => {
    dataRef.current = data;
  }, []);

  return (
    <Dialog onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      {/* We disable animations for the dialog, because react flow doesn't handle it well */}
      <DialogContent className="h-[90dvh] w-full max-w-[90dvw] xl:max-w-7xl p-0 !animate-none">
        <ReactFlowProvider>
          <DialogTitle className="hidden">Flow Editor</DialogTitle>
          <DialogDescription className="hidden">
            Define what happens.
          </DialogDescription>
          <InnerFlowDialog
            flowData={flowData}
            logs={logs}
            context={context}
            onChange={onChange}
          />
        </ReactFlowProvider>
      </DialogContent>
    </Dialog>
  );
}
