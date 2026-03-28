import { MessageComponentSection } from "@/lib/message/schema";
import { useMessageStore } from "@/lib/message/useMessageStore";
import { getUniqueId } from "@/lib/utils";
import { ChevronDown, ChevronUp, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

interface Props {
  component: MessageComponentSection;
  containerIndex: number;
  componentIndex: number;
}

export default function MessageComponentSection({
  component,
  containerIndex,
  componentIndex,
}: Props) {
  const store = useMessageStore();
  const [collapsed, setCollapsed] = useState(false);

  const addTextDisplay = () => {
    store.addTextDisplayToSection(containerIndex, componentIndex, {
      id: getUniqueId(),
      type: 10,
      content: "",
    });
  };

  const setAccessoryType = (type: "thumbnail" | "button" | null) => {
    if (type === null) {
      store.setSectionAccessory(containerIndex, componentIndex, undefined);
    } else if (type === "thumbnail") {
      store.setSectionAccessory(containerIndex, componentIndex, {
        id: getUniqueId(),
        type: 11,
        media: { url: "" },
      });
    } else if (type === "button") {
      store.setSectionAccessory(containerIndex, componentIndex, {
        id: getUniqueId(),
        type: 2,
        style: 1,
        label: "",
        flow_source_id: getUniqueId().toString(),
      });
    }
  };

  return (
    <div className="space-y-3 p-3 bg-dark-3 rounded border border-dark-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="flex items-center space-x-2 text-sm font-medium text-gray-300 hover:text-white"
        >
          {collapsed ? (
            <ChevronDown className="w-4 h-4" />
          ) : (
            <ChevronUp className="w-4 h-4" />
          )}
          <span>Section</span>
        </button>
        <button
          onClick={() =>
            store.deleteComponentFromContainer(containerIndex, componentIndex)
          }
          className="text-red-500 hover:text-red-400"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      {!collapsed && (
        <>
          {/* Text Displays */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-gray-300">
                Text Content
              </label>
              <button
                onClick={addTextDisplay}
                disabled={component.components.length >= 10}
                className="text-xs px-2 py-1 bg-blurple hover:bg-blurple-dark rounded disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-1"
              >
                <Plus className="w-3 h-3" />
                <span>Add Text</span>
              </button>
            </div>

            {component.components.map((textDisplay, index) => (
              <div key={textDisplay.id} className="space-y-1">
                <div className="flex items-center justify-between">
                  <label className="text-xs text-gray-500">
                    Text {index + 1}
                  </label>
                  <button
                    onClick={() =>
                      store.deleteTextDisplayFromSection(
                        containerIndex,
                        componentIndex,
                        index
                      )
                    }
                    className="text-red-500 hover:text-red-400"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
                <textarea
                  value={textDisplay.content}
                  onChange={(e) =>
                    store.setTextDisplayContent(
                      containerIndex,
                      componentIndex,
                      index,
                      e.target.value
                    )
                  }
                  placeholder="Enter text..."
                  className="w-full bg-dark-2 rounded p-2 text-sm resize-none border border-dark-6 focus:border-blurple focus:outline-none"
                  rows={2}
                  maxLength={4000}
                />
                <div className="text-xs text-gray-500 text-right">
                  {textDisplay.content.length} / 4000
                </div>
              </div>
            ))}

            {component.components.length === 0 && (
              <p className="text-sm text-gray-500 text-center py-2">
                No text displays. Click &quot;Add Text&quot; to add one.
              </p>
            )}
          </div>

          {/* Accessory */}
          <div className="space-y-2 pt-2 border-t border-dark-6">
            <label className="block text-sm font-medium text-gray-300">
              Accessory (Optional)
            </label>
            <div className="flex gap-2">
              <button
                onClick={() => setAccessoryType("thumbnail")}
                className={`flex-1 px-3 py-2 rounded text-sm transition-colors ${
                  component.accessory?.type === 11
                    ? "bg-blurple text-white"
                    : "bg-dark-2 text-gray-300 hover:bg-dark-4"
                }`}
              >
                Thumbnail
              </button>
              <button
                onClick={() => setAccessoryType("button")}
                className={`flex-1 px-3 py-2 rounded text-sm transition-colors ${
                  component.accessory?.type === 2
                    ? "bg-blurple text-white"
                    : "bg-dark-2 text-gray-300 hover:bg-dark-4"
                }`}
              >
                Button
              </button>
              {component.accessory && (
                <button
                  onClick={() => setAccessoryType(null)}
                  className="px-3 py-2 bg-red-600 hover:bg-red-700 rounded text-sm text-white"
                >
                  Remove
                </button>
              )}
            </div>

            {/* Thumbnail Accessory Fields */}
            {component.accessory?.type === 11 && (
              <div className="space-y-2 mt-2">
                <label className="block text-xs text-gray-400">
                  Thumbnail URL
                </label>
                <input
                  type="text"
                  value={component.accessory.media.url}
                  onChange={(e) =>
                    store.setSectionAccessory(containerIndex, componentIndex, {
                      ...component.accessory!,
                      media: { url: e.target.value },
                    })
                  }
                  placeholder="https://example.com/image.png"
                  className="w-full bg-dark-2 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
                />
                <label className="block text-xs text-gray-400">
                  Description (Optional)
                </label>
                <input
                  type="text"
                  value={component.accessory.description || ""}
                  onChange={(e) =>
                    store.setSectionAccessory(containerIndex, componentIndex, {
                      ...component.accessory!,
                      description: e.target.value,
                    })
                  }
                  placeholder="Image description"
                  className="w-full bg-dark-2 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
                  maxLength={100}
                />
                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={component.accessory.spoiler ?? false}
                    onChange={(e) =>
                      store.setSectionAccessory(containerIndex, componentIndex, {
                        ...component.accessory!,
                        spoiler: e.target.checked,
                      })
                    }
                    className="rounded bg-dark-2 border-dark-6 text-blurple focus:ring-blurple"
                  />
                  <span className="text-xs text-gray-400">Mark as spoiler</span>
                </label>
              </div>
            )}

            {/* Button Accessory Fields */}
            {component.accessory?.type === 2 && (
              <div className="space-y-2 mt-2">
                <label className="block text-xs text-gray-400">
                  Button Label
                </label>
                <input
                  type="text"
                  value={component.accessory.label}
                  onChange={(e) =>
                    store.setSectionAccessory(containerIndex, componentIndex, {
                      ...component.accessory!,
                      label: e.target.value,
                    })
                  }
                  placeholder="Click me"
                  className="w-full bg-dark-2 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
                  maxLength={80}
                />
                <label className="block text-xs text-gray-400">
                  Button Style
                </label>
                <select
                  value={component.accessory.style}
                  onChange={(e) =>
                    store.setSectionAccessory(containerIndex, componentIndex, {
                      ...component.accessory!,
                      style: parseInt(e.target.value) as 1 | 2 | 3 | 4 | 5,
                    })
                  }
                  className="w-full bg-dark-2 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
                >
                  <option value={1}>Primary (Blurple)</option>
                  <option value={2}>Secondary (Gray)</option>
                  <option value={3}>Success (Green)</option>
                  <option value={4}>Danger (Red)</option>
                  <option value={5}>Link</option>
                </select>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}