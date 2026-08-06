import { MessageComponentTextDisplay } from "@/lib/message/schema";
import { useCurrentMessage } from "@/lib/message/state";

interface Props {
  component: MessageComponentTextDisplay;
  containerIndex: number;
  componentIndex: number;
}

export default function MessageComponentTextDisplay({
  component,
  containerIndex,
  componentIndex,
}: Props) {
  const [
    deleteComponentFromContainer,
    setStandaloneTextDisplayContent,
  ] = useCurrentMessage((state) => [
    state.deleteComponentFromContainer,
    state.setStandaloneTextDisplayContent,
  ]);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-gray-300">
          Text Display
        </label>
        <button
          onClick={() =>
            deleteComponentFromContainer(containerIndex, componentIndex)
          }
          className="text-red-500 hover:text-red-400 text-sm"
        >
          Delete
        </button>
      </div>

      <textarea
        value={component.content}
        onChange={(e) =>
          setStandaloneTextDisplayContent(
            containerIndex,
            componentIndex,
            e.target.value
          )
        }
        placeholder="Enter text content..."
        className="w-full bg-dark-2 rounded p-3 text-sm resize-none border border-dark-6 focus:border-blurple focus:outline-none"
        rows={3}
        maxLength={4000}
      />

      <div className="text-xs text-gray-500 text-right">
        {component.content.length} / 4000 characters
      </div>
    </div>
  );
}