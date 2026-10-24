import { FlowData } from "@/lib/flow/dataSchema";
import { ReactFlowProvider } from "@xyflow/react";
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
import {
  FlowParametersProvider,
  useGetFlowData,
} from "@/lib/flow/parametersStore";

function InnerFlowDialog({
  flowData,
  logFilter,
  context,
  viewportKey,
  onChange,
}: {
  flowData: FlowData;
  logFilter: FlowLogEntriesFilter;
  context: FlowContextType;
  viewportKey?: string;
  onChange: (d: FlowData) => void;
}) {
  const getFlowData = useGetFlowData();
  // Only mounted while the dialog is open, so closed dialogs don't poll.
  const logs = useFlowLogEntries(logFilter);

  const handleChange = useCallback(() => {
    onChange(getFlowData());
  }, [getFlowData, onChange]);

  return (
    <Flow
      flowData={flowData}
      logs={logs}
      context={context}
      viewportKey={viewportKey}
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
  viewportKey,
}: {
  flowData: FlowData;
  logFilter: FlowLogEntriesFilter;
  onClose: (data: FlowData) => void;
  context: FlowContextType;
  viewportKey?: string;
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
          <FlowParametersProvider initial={flowData.parameters}>
            <InnerFlowDialog
              flowData={flowData}
              logFilter={logFilter}
              context={context}
              viewportKey={viewportKey}
              onChange={onChange}
            />
          </FlowParametersProvider>
        </ReactFlowProvider>
      </DialogContent>
    </Dialog>
  );
}
