import { FlowReadOnlyContext } from "@/lib/flow/context";
import { NodeData, NodeProps } from "@/lib/flow/dataSchema";
import {
  fitGroupToMembers,
  getGroupData,
  getGroupMemberIds,
  groupColors,
  GroupColor,
  groupLabelMaxLength,
  groupMaxSize,
  groupMinHeight,
  groupMinWidth,
  isGroupNode,
  setGroupCollapsed,
} from "@/lib/flow/groups";
import { isNoteNode } from "@/lib/flow/notes";
import { cn } from "@/lib/utils";
import {
  Node,
  NodeResizer,
  useReactFlow,
  useStore,
  useStoreApi,
} from "@xyflow/react";
import {
  CheckIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  EllipsisIcon,
  MousePointerSquareDashedIcon,
  PencilIcon,
  ScanIcon,
  Trash2Icon,
  UngroupIcon,
} from "lucide-react";
import {
  KeyboardEvent,
  memo,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

// A frame around blocks with a label, which can be collapsed to its header.
// Only the header and the resize handles take the pointer, so the blocks and
// the canvas inside the frame can be used as if it wasn't there.
export default memo(function FlowNodeGroup({ id, data, selected }: NodeProps) {
  const { label, color, collapsed, hiddenIds } = getGroupData({ data });
  const readOnly = useContext(FlowReadOnlyContext);
  const { setNodes, deleteElements, getNodes, updateNodeData } =
    useReactFlow<Node<NodeData>>();
  const store = useStoreApi<Node<NodeData>>();

  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(label);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing) inputRef.current?.select();
  }, [editing]);

  // The blocks in a collapsed group, without the notes and groups in it.
  const hiddenBlockCount = useStore(
    useCallback(
      (s) =>
        hiddenIds.filter((hid) => {
          const node = s.nodeLookup.get(hid);
          return node && !isGroupNode(node) && !isNoteNode(node);
        }).length,
      [hiddenIds]
    )
  );

  const toggleCollapsed = useCallback(
    () => setNodes((nodes) => setGroupCollapsed(nodes, id, !collapsed)),
    [id, collapsed, setNodes]
  );

  const startEditing = useCallback(() => {
    if (readOnly) return;
    setDraft(label);
    setEditing(true);
  }, [label, readOnly]);

  const finishEditing = useCallback(() => {
    setEditing(false);
    const value = draft.trim().slice(0, groupLabelMaxLength);
    if (value && value !== label) updateNodeData(id, { group_label: value });
  }, [id, draft, label, updateNodeData]);

  const onInputKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      // Typing in the name mustn't move or delete the group.
      e.stopPropagation();
      if (e.key === "Enter") {
        e.preventDefault();
        finishEditing();
      } else if (e.key === "Escape") {
        e.preventDefault();
        setEditing(false);
      }
    },
    [finishEditing]
  );

  const ungroup = useCallback(
    () => deleteElements({ nodes: [{ id }] }),
    [id, deleteElements]
  );

  const deleteWithBlocks = useCallback(() => {
    const nodes = getNodes();
    const group = nodes.find((n) => n.id === id);
    if (!group) return;
    const memberIds = getGroupMemberIds(group, nodes);
    deleteElements({ nodes: [id, ...memberIds].map((nid) => ({ id: nid })) });
  }, [id, getNodes, deleteElements]);

  const selectBlocks = useCallback(() => {
    const nodes = getNodes();
    const group = nodes.find((n) => n.id === id);
    if (!group) return;
    store.getState().addSelectedNodes(getGroupMemberIds(group, nodes));
  }, [id, getNodes, store]);

  const fitToBlocks = useCallback(
    () => setNodes((nodes) => fitGroupToMembers(nodes, id)),
    [id, setNodes]
  );

  const labelElement = editing ? (
    <input
      ref={inputRef}
      value={draft}
      maxLength={groupLabelMaxLength}
      onChange={(e) => setDraft(e.target.value)}
      onBlur={finishEditing}
      onKeyDown={onInputKeyDown}
      placeholder="Group name"
      className="nodrag nopan min-w-0 flex-auto bg-background/80 rounded px-1.5 py-0.5 text-sm font-semibold outline-none ring-1"
      style={{ color, ["--tw-ring-color" as string]: color }}
    />
  ) : (
    <div
      className="min-w-0 truncate text-sm font-semibold px-1.5 py-0.5"
      style={{ color }}
      onDoubleClick={(e) => {
        e.stopPropagation();
        startEditing();
      }}
      title={readOnly ? label : "Double-click to rename"}
    >
      {label}
    </div>
  );

  const menu = readOnly ? null : (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className="nodrag rounded p-1 text-muted-foreground hover:text-foreground hover:bg-foreground/10"
          title="Group options"
        >
          <EllipsisIcon className="size-4" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuItem onSelect={startEditing}>
          <PencilIcon className="size-4 mr-2" />
          Rename
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuLabel className="text-xs text-muted-foreground font-normal">
          Color
        </DropdownMenuLabel>
        <div className="flex flex-wrap gap-1.5 px-2 pb-2">
          {(Object.keys(groupColors) as GroupColor[]).map((key) => (
            <button
              key={key}
              className="size-6 rounded-full flex items-center justify-center ring-offset-2 ring-offset-popover hover:ring-2"
              style={{
                backgroundColor: groupColors[key],
                ["--tw-ring-color" as string]: groupColors[key],
              }}
              onClick={() => updateNodeData(id, { group_color: key })}
              title={key[0].toUpperCase() + key.slice(1)}
            >
              {groupColors[key] === color && (
                <CheckIcon className="size-3.5 text-white" />
              )}
            </button>
          ))}
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={toggleCollapsed}>
          {collapsed ? (
            <ChevronDownIcon className="size-4 mr-2" />
          ) : (
            <ChevronUpIcon className="size-4 mr-2" />
          )}
          {collapsed ? "Expand" : "Collapse"}
        </DropdownMenuItem>
        {!collapsed && (
          <>
            <DropdownMenuItem onSelect={selectBlocks}>
              <MousePointerSquareDashedIcon className="size-4 mr-2" />
              Select blocks
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={fitToBlocks}>
              <ScanIcon className="size-4 mr-2" />
              Fit to blocks
            </DropdownMenuItem>
          </>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={ungroup}>
          <UngroupIcon className="size-4 mr-2" />
          Ungroup
        </DropdownMenuItem>
        <DropdownMenuItem
          onSelect={deleteWithBlocks}
          className="text-destructive focus:text-destructive"
        >
          <Trash2Icon className="size-4 mr-2" />
          Delete group and blocks
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const collapseButton = (
    <button
      className="nodrag rounded p-1 hover:bg-foreground/10"
      style={{ color }}
      onClick={toggleCollapsed}
      title={collapsed ? "Expand group" : "Collapse group"}
    >
      {collapsed ? (
        <ChevronDownIcon className="size-4" />
      ) : (
        <ChevronUpIcon className="size-4" />
      )}
    </button>
  );

  if (collapsed) {
    return (
      <div
        className={cn(
          "pointer-events-auto h-full w-full rounded-lg border-2 bg-muted shadow-md flex items-center gap-2 pl-2 pr-1.5",
          !readOnly && "cursor-grab"
        )}
        style={{
          borderColor: selected ? color : `${color}99`,
          backgroundImage: `linear-gradient(${color}26, ${color}26)`,
        }}
        onDoubleClick={(e) => {
          // Renaming has its own double click.
          if (!editing && e.target === e.currentTarget) toggleCollapsed();
        }}
      >
        <div
          className="size-2.5 rounded-full flex-none"
          style={{ backgroundColor: color }}
        />
        {labelElement}
        <div className="text-xs text-muted-foreground flex-none whitespace-nowrap">
          {hiddenBlockCount} {hiddenBlockCount === 1 ? "block" : "blocks"}
        </div>
        <div className="flex-auto" />
        {menu}
        {collapseButton}
      </div>
    );
  }

  return (
    <>
      {!readOnly && (
        <NodeResizer
          isVisible={selected}
          minWidth={groupMinWidth}
          minHeight={groupMinHeight}
          maxWidth={groupMaxSize}
          maxHeight={groupMaxSize}
          color={color}
          handleClassName="!size-2.5 !rounded-sm !border-0"
          lineClassName="!border-transparent"
        />
      )}
      <div
        className="h-full w-full rounded-xl border-2 relative"
        style={{
          borderColor: selected ? color : `${color}b3`,
          backgroundColor: `${color}14`,
        }}
      >
        <div
          className={cn(
            "pointer-events-auto absolute inset-x-0 top-0 h-11 flex items-center gap-1 pl-1.5 pr-1.5 rounded-t-[10px]",
            !readOnly && "cursor-grab"
          )}
          style={{ backgroundColor: `${color}1f` }}
        >
          {labelElement}
          <div className="flex-auto" />
          {menu}
          {collapseButton}
        </div>
      </div>
    </>
  );
});
