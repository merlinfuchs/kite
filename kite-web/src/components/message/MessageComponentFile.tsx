import { MessageComponentFile } from "@/lib/message/schema";
import { useCurrentMessage } from "@/lib/message/state";

interface Props {
  component: MessageComponentFile;
  containerIndex: number;
  componentIndex: number;
}

export default function MessageComponentFile({
  component,
  containerIndex,
  componentIndex,
}: Props) {
  const [deleteComponentFromContainer, setFileUrl, setFileSpoiler] =
  useCurrentMessage((state) => [
    state.deleteComponentFromContainer,
    state.setFileUrl,
    state.setFileSpoiler,
  ]);

  return (
    <div className="space-y-3 p-3 bg-dark-3 rounded border border-dark-6">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-gray-300">File</label>
        <button
          onClick={() =>
            deleteComponentFromContainer(containerIndex, componentIndex)
          }
          className="text-red-500 hover:text-red-400 text-sm"
        >
          Delete
        </button>
      </div>

      <div className="space-y-2">
        <label className="block text-sm font-medium text-gray-300">
          File URL
        </label>
        <input
          type="text"
          value={component.file.url}
          onChange={(e) =>
            setFileUrl(containerIndex, componentIndex, e.target.value)
          }
          placeholder="https://example.com/file.pdf"
          className="w-full bg-dark-2 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
        />
        <p className="text-xs text-gray-500">
          Enter the URL of the file you want to display
        </p>
      </div>

      <div className="space-y-2">
        <label className="flex items-center space-x-2">
          <input
            type="checkbox"
            checked={component.spoiler ?? false}
            onChange={(e) =>
              setFileSpoiler(
                containerIndex,
                componentIndex,
                e.target.checked
              )
            }
            className="rounded bg-dark-2 border-dark-6 text-blurple focus:ring-blurple"
          />
          <span className="text-sm text-gray-300">Mark as spoiler</span>
        </label>
      </div>
    </div>
  );
}