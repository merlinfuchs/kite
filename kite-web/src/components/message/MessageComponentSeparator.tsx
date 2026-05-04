import { MessageComponentSeparator } from "@/lib/message/schema";
import { useMessageStore } from "@/lib/message/useMessageStore";

interface Props {
  component: MessageComponentSeparator;
  containerIndex: number;
  componentIndex: number;
}

export default function MessageComponentSeparator({
  component,
  containerIndex,
  componentIndex,
}: Props) {
  const store = useMessageStore();

  return (
    <div className="space-y-3 p-3 bg-dark-3 rounded border border-dark-6">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-gray-300">Separator</label>
        <button
          onClick={() =>
            store.deleteComponentFromContainer(containerIndex, componentIndex)
          }
          className="text-red-500 hover:text-red-400 text-sm"
        >
          Delete
        </button>
      </div>

      <div className="space-y-2">
        <label className="flex items-center space-x-2">
          <input
            type="checkbox"
            checked={component.divider ?? true}
            onChange={(e) =>
              store.setSeparatorDivider(
                containerIndex,
                componentIndex,
                e.target.checked
              )
            }
            className="rounded bg-dark-2 border-dark-6 text-blurple focus:ring-blurple"
          />
          <span className="text-sm text-gray-300">Show divider line</span>
        </label>
      </div>

      <div className="space-y-2">
        <label className="block text-sm font-medium text-gray-300">
          Spacing
        </label>
        <div className="flex gap-2">
          <button
            onClick={() =>
              store.setSeparatorSpacing(containerIndex, componentIndex, 1)
            }
            className={`flex-1 px-3 py-2 rounded text-sm transition-colors ${
              (component.spacing ?? 1) === 1
                ? "bg-blurple text-white"
                : "bg-dark-2 text-gray-300 hover:bg-dark-4"
            }`}
          >
            Small
          </button>
          <button
            onClick={() =>
              store.setSeparatorSpacing(containerIndex, componentIndex, 2)
            }
            className={`flex-1 px-3 py-2 rounded text-sm transition-colors ${
              component.spacing === 2
                ? "bg-blurple text-white"
                : "bg-dark-2 text-gray-300 hover:bg-dark-4"
            }`}
          >
            Large
          </button>
        </div>
      </div>

      {/* Preview */}
      <div className="pt-2">
        <div className="text-xs text-gray-500 mb-1">Preview:</div>
        <div className={component.spacing === 2 ? "py-4" : "py-2"}>
          {component.divider && (
            <div className="border-t border-dark-6"></div>
          )}
        </div>
      </div>
    </div>
  );
}