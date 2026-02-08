import {
  MessageComponentMediaGallery,
  MessageComponentMediaGalleryItem,
} from "@/lib/message/schema";
import { useMessageStore } from "@/lib/message/useMessageStore";
import { getUniqueId } from "@/lib/utils";
import {
  ChevronDown,
  ChevronUp,
  Plus,
  Trash2,
  MoveUp,
  MoveDown,
} from "lucide-react";
import { useState } from "react";

interface Props {
  component: MessageComponentMediaGallery;
  containerIndex: number;
  componentIndex: number;
}

export default function MessageComponentMediaGallery({
  component,
  containerIndex,
  componentIndex,
}: Props) {
  const store = useMessageStore();
  const [collapsed, setCollapsed] = useState(false);

  const addItem = () => {
    const newItem: MessageComponentMediaGalleryItem = {
      id: getUniqueId(),
      media: { url: "" },
    };
    store.addMediaGalleryItem(containerIndex, componentIndex, newItem);
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
          <span>Media Gallery</span>
          <span className="text-xs text-gray-500">
            ({component.items.length}/10)
          </span>
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
          <div className="flex items-center justify-between">
            <p className="text-xs text-gray-500">
              Add multiple images to create a gallery
            </p>
            <button
              onClick={addItem}
              disabled={component.items.length >= 10}
              className="text-xs px-2 py-1 bg-blurple hover:bg-blurple-dark rounded disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-1"
            >
              <Plus className="w-3 h-3" />
              <span>Add Item</span>
            </button>
          </div>

          {/* Gallery Items */}
          <div className="space-y-3">
            {component.items.map((item, itemIndex) => (
              <div
                key={item.id}
                className="p-3 bg-dark-2 rounded border border-dark-6 space-y-2"
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-gray-400">
                    Item {itemIndex + 1}
                  </span>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() =>
                        store.moveMediaGalleryItemUp(
                          containerIndex,
                          componentIndex,
                          itemIndex
                        )
                      }
                      disabled={itemIndex === 0}
                      className="p-1 text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
                      title="Move up"
                    >
                      <MoveUp className="w-3 h-3" />
                    </button>
                    <button
                      onClick={() =>
                        store.moveMediaGalleryItemDown(
                          containerIndex,
                          componentIndex,
                          itemIndex
                        )
                      }
                      disabled={itemIndex === component.items.length - 1}
                      className="p-1 text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
                      title="Move down"
                    >
                      <MoveDown className="w-3 h-3" />
                    </button>
                    <button
                      onClick={() =>
                        store.deleteMediaGalleryItem(
                          containerIndex,
                          componentIndex,
                          itemIndex
                        )
                      }
                      className="p-1 text-red-500 hover:text-red-400"
                      title="Delete"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="block text-xs text-gray-400">
                    Media URL *
                  </label>
                  <input
                    type="text"
                    value={item.media.url}
                    onChange={(e) =>
                      store.setMediaGalleryItemUrl(
                        containerIndex,
                        componentIndex,
                        itemIndex,
                        e.target.value
                      )
                    }
                    placeholder="https://example.com/image.png"
                    className="w-full bg-dark-4 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
                  />
                </div>

                <div className="space-y-2">
                  <label className="block text-xs text-gray-400">
                    Description (Optional)
                  </label>
                  <input
                    type="text"
                    value={item.description || ""}
                    onChange={(e) =>
                      store.setMediaGalleryItemDescription(
                        containerIndex,
                        componentIndex,
                        itemIndex,
                        e.target.value
                      )
                    }
                    placeholder="Image description"
                    className="w-full bg-dark-4 rounded p-2 text-sm border border-dark-6 focus:border-blurple focus:outline-none"
                    maxLength={100}
                  />
                </div>

                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={item.spoiler ?? false}
                    onChange={(e) =>
                      store.setMediaGalleryItemSpoiler(
                        containerIndex,
                        componentIndex,
                        itemIndex,
                        e.target.checked
                      )
                    }
                    className="rounded bg-dark-4 border-dark-6 text-blurple focus:ring-blurple"
                  />
                  <span className="text-xs text-gray-400">
                    Mark as spoiler
                  </span>
                </label>
              </div>
            ))}

            {component.items.length === 0 && (
              <p className="text-sm text-gray-500 text-center py-4">
                No items in gallery. Click "Add Item" to add an image.
              </p>
            )}
          </div>
        </>
      )}
    </div>
  );
}