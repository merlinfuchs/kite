import { FlowData, NodeData } from "@/lib/flow/data";
import debounce from "@/lib/util/debounce";
import {
  ArrowLeftIcon,
  ArrowPathIcon,
  ArrowUpIcon,
  CheckIcon,
} from "@heroicons/react/24/solid";
import { useEffect } from "react";
import { useReactFlow } from "reactflow";

interface Props {
  hasUnsavedChanges: boolean;
  isSaving: boolean;
  onSave: (d: FlowData) => void;
  isDeploying: boolean;
  onDeploy: (d: FlowData) => void;
  onExit: () => void;
}

export default function FlowNav({
  hasUnsavedChanges,
  isSaving,
  onSave,
  isDeploying,
  onDeploy,
  onExit,
}: Props) {
  const { getEdges, getNodes } = useReactFlow<NodeData>();

  function save() {
    onSave({
      nodes: getNodes(),
      edges: getEdges(),
    });
  }

  function deploy() {
    onDeploy({
      nodes: getNodes(),
      edges: getEdges(),
    });
  }

  useEffect(() => {
    async function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        save();
      }

      if (e.key === "p" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        save();
        deploy();
      }
    }

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onSave]);

  return (
    <div className="h-12 bg-dark-2 flex items-center space-x-8 px-4 select-none">
      <button
        className="flex space-x-2 text-gray-300 hover:text-white items-center"
        onClick={onExit}
      >
        <ArrowLeftIcon className="h-5 w-5" />
        <div>Back to Server</div>
      </button>
      {isSaving ? (
        <div
          className="flex space-x-2 text-gray-300 hover:text-white items-center"
          onClick={save}
        >
          <ArrowPathIcon className="h-5 w-5 animate-spin" />
          <div>Saving Changes</div>
        </div>
      ) : hasUnsavedChanges ? (
        <button
          className="flex space-x-2 text-gray-300 hover:text-white items-center"
          onClick={save}
        >
          <div className="h-3 w-3 rounded-full bg-white"></div>
          <div>Save Changes</div>
        </button>
      ) : (
        <div className="flex space-x-2 text-gray-300 hover:text-white items-center">
          <CheckIcon className="h-5 w-5" />
          <div>No Unsaved Changes</div>
        </div>
      )}
      {isDeploying ? (
        <div className="flex space-x-2 text-gray-300 hover:text-white items-center">
          <ArrowPathIcon className="h-5 w-5 animate-spin" />
          <div>Deploying to Server</div>
        </div>
      ) : (
        <button
          className="flex space-x-2 text-gray-300 hover:text-white items-center"
          onClick={deploy}
        >
          <ArrowUpIcon className="h-5 w-5" />
          <div>Deploy to Server</div>
        </button>
      )}
    </div>
  );
}
