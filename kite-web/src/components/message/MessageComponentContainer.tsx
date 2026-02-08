import { MessageComponentContainer } from "@/lib/message/schema";
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
import MessageComponentSection from "./MessageComponentSection";
import MessageComponentTextDisplay from "./MessageComponentTextDisplay";
import MessageComponentMediaGallery from "./MessageComponentMediaGallery";
import MessageComponentFile from "./MessageComponentFile";
import MessageComponentSeparator from "./MessageComponentSeparator";

interface Props {
  component: MessageComponentContainer;
  containerIndex: number;
}

export default function MessageComponentContainer({
  component,
  containerIndex,
}: Props) {
  const store = useMessageStore();
  const [collapsed, setCollapsed] = useState(false);
  const [showAddMenu, setShowAddMenu] = useState(false);

  const addComponent = (type: number) => {
    switch (type) {
      case 1: // Action Row
        store.addComponentToContainer(containerIndex, {
          id: getUniqueId(),
          type: 1,
          components: [],
        });
        break;
      case 9: // Section
        store.addComponentToContainer(containerIndex, {
          id: getUniqueId(),
          type: 9,
          components: [],
        });
        break;
      case 10: // Text Display
        store.addComponentToContainer(containerIndex, {
          id: getUniqueId(),
          type: 10,
          content: "",
        });
        break;
      case 12: // Media Gallery
        store.addComponentToContainer(containerIndex, {
          id: getUniqueId(),
          type: 12,
          items: [],
        });
        break;
      case 13: // File
        store.addComponentToContainer(containerIndex, {
          id: getUniqueId(),
          type: 13,
          file: { url: "" },
        });
        break;
      case 14: // Separator
        store.addComponentToContainer(containerIndex, {
          id: getUniqueId(),
          type: 14,
          divider: true,
          spacing: 1,
        });
        break;
    }
    setShowAddMenu(false);
  };

  return (
    <div className="space-y-3 p-4 bg-dark-4 rounded-lg border-2 border-dark-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="flex items-center space-x-2 text-base font-semibold text-gray-200 hover:text-white"
        >
          {collapsed ? (
            <ChevronDown className="w-5 h-5" />
          ) : (
            <ChevronUp className="w-5 h-5" />
          )}
          <span>Container</span>
          <span className="text-sm text-gray-500">
            ({component.components.length}/10)
          </span>
        </button>
        <button
          onClick={() => store.deleteComponentRow(containerIndex)}
          className="text-red-500 hover:text-red-400"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      {!collapsed && (
        <>
          {/* Container Settings */}
          <div className="grid grid-cols-2 gap-3 pb-3 border-b border-dark-6">
            <div className="space-y-2">
              <label className="block text-xs font-medium text-gray-400">
                Accent Color (Optional)
              </label>
              <div className="flex gap-2">
                <input
                  type="color"
                  value={
                    component.accent_color
                      ? `#${component.accent_color.toString(16).padStart(6, "0")}`
                      : "#5865F2"
                  }
                  onChange={(e) => {
                    const hex = e.target.value.replace("#", "");
                    const decimal = parseInt(hex, 16);
                    store.setContainerAccentColor(containerIndex, decimal);
                  }}
                  className="w-12 h-8 rounded cursor-pointer"
                />
                <input
                  type="text"
                  value={
                    component.accent_color
                      ? `#${component.accent_color.toString(16).padStart(6, "0").toUpperCase()}`
                      : ""
                  }
                  onChange={(e) => {
                    const hex = e.target.value.replace("#", "");
                    if (hex.length === 6) {
                      const decimal = parseInt(hex, 16);
                      if (!isNaN(decimal)) {
                        store.setContainerAccentColor(containerIndex, decimal);
                      }
                    } else if (hex.length === 0) {
                      store.setContainerAccentColor(containerIndex, undefined);
                    }
                  }}
                  placeholder="#5865F2"
                  className="flex-1 bg-dark-2 rounded p-2 text-xs border border-dark-6 focus:border-blurple focus:outline-none uppercase"
                  maxLength={7}
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="block text-xs font-medium text-gray-400">
                Options
              </label>
              <label className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={component.spoiler ?? false}
                  onChange={(e) =>
                    store.setContainerSpoiler(containerIndex, e.target.checked)
                  }
                  className="rounded bg-dark-2 border-dark-6 text-blurple focus:ring-blurple"
                />
                <span className="text-xs text-gray-300">Mark as spoiler</span>
              </label>
            </div>
          </div>

          {/* Add Component Menu */}
          <div className="relative">
            <button
              onClick={() => setShowAddMenu(!showAddMenu)}
              disabled={component.components.length >= 10}
              className="w-full px-3 py-2 bg-blurple hover:bg-blurple-dark rounded disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2 text-sm font-medium"
            >
              <Plus className="w-4 h-4" />
              <span>Add Component</span>
            </button>

            {showAddMenu && (
              <div className="absolute top-full left-0 right-0 mt-1 bg-dark-2 rounded border border-dark-6 shadow-lg z-10">
                <button
                  onClick={() => addComponent(1)}
                  className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:bg-dark-3 first:rounded-t last:rounded-b"
                >
                  Action Row
                </button>
                <button
                  onClick={() => addComponent(9)}
                  className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:bg-dark-3"
                >
                  Section
                </button>
                <button
                  onClick={() => addComponent(10)}
                  className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:bg-dark-3"
                >
                  Text Display
                </button>
                <button
                  onClick={() => addComponent(12)}
                  className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:bg-dark-3"
                >
                  Media Gallery
                </button>
                <button
                  onClick={() => addComponent(13)}
                  className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:bg-dark-3"
                >
                  File
                </button>
                <button
                  onClick={() => addComponent(14)}
                  className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:bg-dark-3 last:rounded-b"
                >
                  Separator
                </button>
              </div>
            )}
          </div>

          {/* Container Components */}
          <div className="space-y-3">
            {component.components.map((comp, compIndex) => (
              <div key={comp.id} className="relative">
                <div className="absolute -left-2 top-0 bottom-0 flex flex-col gap-1">
                  <button
                    onClick={() =>
                      store.moveContainerComponentUp(containerIndex, compIndex)
                    }
                    disabled={compIndex === 0}
                    className="p-1 text-gray-400 hover:text-white bg-dark-3 rounded disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Move up"
                  >
                    <MoveUp className="w-3 h-3" />
                  </button>
                  <button
                    onClick={() =>
                      store.moveContainerComponentDown(
                        containerIndex,
                        compIndex
                      )
                    }
                    disabled={compIndex === component.components.length - 1}
                    className="p-1 text-gray-400 hover:text-white bg-dark-3 rounded disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Move down"
                  >
                    <MoveDown className="w-3 h-3" />
                  </button>
                </div>

                <div className="ml-4">
                  {comp.type === 1 && (
                    <div className="p-3 bg-dark-3 rounded border border-dark-6">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-gray-300">
                          Action Row
                        </span>
                        <button
                          onClick={() =>
                            store.deleteComponentFromContainer(
                              containerIndex,
                              compIndex
                            )
                          }
                          className="text-red-500 hover:text-red-400"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                      <p className="text-xs text-gray-500">
                        Action rows with buttons/select menus will be editable
                        here in a future update.
                      </p>
                    </div>
                  )}
                  {comp.type === 9 && (
                    <MessageComponentSection
                      component={comp}
                      containerIndex={containerIndex}
                      componentIndex={compIndex}
                    />
                  )}
                  {comp.type === 10 && (
                    <MessageComponentTextDisplay
                      component={comp}
                      containerIndex={containerIndex}
                      componentIndex={compIndex}
                    />
                  )}
                  {comp.type === 12 && (
                    <MessageComponentMediaGallery
                      component={comp}
                      containerIndex={containerIndex}
                      componentIndex={compIndex}
                    />
                  )}
                  {comp.type === 13 && (
                    <MessageComponentFile
                      component={comp}
                      containerIndex={containerIndex}
                      componentIndex={compIndex}
                    />
                  )}
                  {comp.type === 14 && (
                    <MessageComponentSeparator
                      component={comp}
                      containerIndex={containerIndex}
                      componentIndex={compIndex}
                    />
                  )}
                </div>
              </div>
            ))}

            {component.components.length === 0 && (
              <p className="text-sm text-gray-500 text-center py-6">
                No components in container. Click "Add Component" to add one.
              </p>
            )}
          </div>
        </>
      )}
    </div>
  );
}