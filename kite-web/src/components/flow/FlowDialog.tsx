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
import { FlowLogEntriesFilter, useFlowLogEntries } from "@/lib/hooks/api";

function InnerFlowDialog({
  flowData,
  logFilter,
  context,
  onChange,
}: {
  flowData: FlowData;
  logFilter: FlowLogEntriesFilter;
  context: FlowContextType;
  onChange: (d: FlowData) => void;
}) {
  const { getNodes, getEdges } = useReactFlow<NodeType>();
  // Only mounted while the dialog is open, so closed dialogs don't poll.
  const logs = useFlowLogEntries(logFilter);

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
  logFilter,
  context,
}: {
  flowData: FlowData;
  logFilter: FlowLogEntriesFilter;
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
          <style jsx global>{`
            @media (max-width: 767px) {
              [data-sonner-toaster][data-y-position="top"] {
                top: 56px !important;
                --offset: 56px !important;
              }
            }
          `}</style>
          <DialogTitle className="hidden">Flow Editor</DialogTitle>
          <DialogDescription className="hidden">
            Define what happens.
          </DialogDescription>
          <InnerFlowDialog
            flowData={flowData}
            logFilter={logFilter}
            context={context}
            onChange={onChange}
          />
        </ReactFlowProvider>
      </DialogContent>
    </Dialog>
  );
}
