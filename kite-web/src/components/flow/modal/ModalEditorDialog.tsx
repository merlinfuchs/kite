import {
  ModalInputType,
  changeModalInputType,
  duplicateModalComponent,
  modalComponentNumber,
  modalIdentifierFromLabel,
  modalInputHasOptions,
  modalInputHasPlaceholder,
  modalInputHasValueLimits,
  modalInputTypes,
  modalMaxComponents,
  modalMaxFiles,
  modalOptionCounts,
  modalTextLimits,
  newModalInput,
  newModalOption,
  nextModalInputNumber,
  normalizeModalComponents,
} from "@/lib/flow/modal";
import {
  History,
  pushHistory,
  redoHistory,
  undoHistory,
} from "@/lib/flow/history";
import { useHookedTheme } from "@/lib/hooks/theme";
import {
  EmojiData,
  ModalComponentData,
  ModalComponentOptionData,
  ModalData,
} from "@/lib/types/flow.gen";
import { discordEmojiUrl } from "@/tools/common/utils/discordCdn";
import { cn } from "@/lib/utils";
import {
  AtSignIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  ChevronUpIcon,
  CircleAlertIcon,
  CircleDotIcon,
  CodeIcon,
  CopyIcon,
  HashIcon,
  ListChecksIcon,
  ListIcon,
  LucideIcon,
  PaintbrushIcon,
  RedoIcon,
  ShieldIcon,
  SmileIcon,
  SquareCheckIcon,
  TextCursorInputIcon,
  TrashIcon,
  UndoIcon,
  UploadIcon,
  UserIcon,
  ViewIcon,
  WandSparklesIcon,
  XIcon,
} from "lucide-react";
import {
  ReactNode,
  RefObject,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { toast } from "sonner";
import { FlowReadOnlyContext, PreviewInert } from "@/lib/flow/context";
import EmojiPicker from "../../common/EmojiPicker";
import PlaceholderInput from "../../common/PlaceholderInput";
import Twemoji from "../../common/Twemoji";
import ConfirmDialog from "../../common/ConfirmDialog";
import MessageCollapsibleSection from "../../message/MessageCollapsibleSection";
import MessageControlsButton from "../../message/MessageControlsButton";
import { ChannelTypesInput } from "../../message/MessageComponentSelectMenu";
import { Button } from "../../ui/button";
import { Card } from "../../ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "../../ui/dialog";
import { Drawer, DrawerContent, DrawerTrigger } from "../../ui/drawer";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../../ui/dropdown-menu";
import { Input } from "../../ui/input";
import { Label } from "../../ui/label";
import { ScrollArea } from "../../ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../ui/select";
import { Switch } from "../../ui/switch";
import { Textarea } from "../../ui/textarea";
import { ToggleGroup, ToggleGroupItem } from "../../ui/toggle-group";
import { Tooltip, TooltipContent, TooltipTrigger } from "../../ui/tooltip";
import FlowPlaceholderExplorer from "../FlowPlaceholderExplorer";
import ModalJSONDialog from "./ModalJSONDialog";
import ModalPreview from "./ModalPreview";

const inputTypeIcons: Record<ModalInputType, LucideIcon> = {
  text_input: TextCursorInputIcon,
  string_select: ListIcon,
  user_select: UserIcon,
  role_select: ShieldIcon,
  mentionable_select: AtSignIcon,
  channel_select: HashIcon,
  radio_group: CircleDotIcon,
  checkbox_group: ListChecksIcon,
  checkbox: SquareCheckIcon,
  file_upload: UploadIcon,
};

function inputTypeLabel(type?: string) {
  return (modalInputTypes.find((t) => t.value === type) ?? modalInputTypes[0])
    .label;
}

function componentField(i: number) {
  return `modal_data.components.${i}`;
}

function inputField(i: number) {
  return `${componentField(i)}.components.0`;
}

// The indexes of the components that have an error, so their cards and
// preview can point at it.
function componentErrorIndexes(errors: Record<string, string>) {
  const res = new Set<number>();
  for (const key of Object.keys(errors)) {
    const match = key.match(/^modal_data\.components\.(\d+)\./);
    if (match) res.add(parseInt(match[1]));
  }
  return res;
}

function parseOptionalInt(v: string) {
  const n = parseInt(v);
  return isNaN(n) ? undefined : n;
}

let nextListKey = 0;

// Keys for the components and options, which have no IDs, so the inputs and
// the open card stay with the item when it's moved. Moves, inserts and
// removals have to go through these, other changes, like an undo, only keep
// the keys of the first items.
function useListKeys(length: number) {
  const keys = useRef<number[]>([]);
  while (keys.current.length < length) keys.current.push(nextListKey++);
  keys.current.length = length;

  return {
    keys: keys.current,
    move: (i: number, j: number) => {
      const k = keys.current;
      [k[i], k[j]] = [k[j], k[i]];
    },
    insert: (i: number) => {
      const key = nextListKey++;
      keys.current.splice(i, 0, key);
      return key;
    },
    remove: (i: number) => {
      keys.current.splice(i, 1);
    },
  };
}

export default function ModalEditorDialog({
  modal,
  onChange,
  errors,
}: {
  modal: ModalData | undefined;
  onChange: (modal: ModalData) => void;
  errors: Record<string, string>;
}) {
  // Older modals are converted when they are edited, so the editor only
  // deals with labels and text displays.
  const components = useMemo(
    () => normalizeModalComponents(modal?.components),
    [modal?.components]
  );

  const inputCount = components.filter((c) => c.type !== "text_display").length;
  const textCount = components.length - inputCount;
  const errorCount = Object.keys(errors).filter((k) =>
    k.startsWith("modal_data")
  ).length;

  return (
    <Dialog>
      <div className="space-y-2">
        <DialogTrigger asChild>
          <Button className="w-full">Configure Modal</Button>
        </DialogTrigger>
        <div className="text-sm text-muted-foreground flex items-center gap-2">
          <div className="truncate flex-auto">
            {inputCount} {inputCount === 1 ? "input" : "inputs"}
            {textCount > 0 &&
              `, ${textCount} ${textCount === 1 ? "text" : "texts"}`}
          </div>
          {errorCount > 0 && (
            <div className="flex items-center gap-1 text-red-600 dark:text-red-400 flex-none">
              <CircleAlertIcon className="h-4 w-4" />
              {errorCount} {errorCount === 1 ? "issue" : "issues"}
            </div>
          )}
        </div>
      </div>
      <DialogContent
        className="h-full max-h-none rounded-none sm:h-[90dvh] sm:rounded-lg w-full md:max-w-[90dvw] xl:max-w-7xl p-0 !animate-none overflow-hidden"
        // Opening shouldn't select the title, so typing doesn't replace it.
        // The dialog is focused instead, so the undo shortcut applies to it
        // and not to the flow behind it.
        onOpenAutoFocus={(e) => {
          e.preventDefault();
          (e.currentTarget as HTMLElement | null)?.focus();
        }}
      >
        <ModalEditorBody
          modal={modal}
          components={components}
          onChange={onChange}
          errors={errors}
        />
      </DialogContent>
    </Dialog>
  );
}

// Edits within this time are undone together, so undo doesn't go back one
// letter at a time while typing.
const historyGroupMs = 600;

// Undo and redo of the modal editor, like the message editor has. The history
// lives as long as the editor is open.
function useModalHistory(
  current: ModalData,
  apply: (modal: ModalData) => void,
  containerRef: RefObject<HTMLElement | null>
) {
  const [history, setHistory] = useState<History<ModalData>>({
    past: [],
    future: [],
  });
  const lastChange = useRef(0);

  const change = useCallback(
    (next: ModalData) => {
      const now = Date.now();
      // Undo and redo reset lastChange, so the undone states are dropped
      // here too.
      if (
        now - lastChange.current > historyGroupMs ||
        history.past.length === 0
      ) {
        setHistory(pushHistory(history, current));
      }
      lastChange.current = now;
      apply(next);
    },
    [history, current, apply]
  );

  const restore = useCallback(
    (res: [History<ModalData>, ModalData] | null) => {
      if (!res) return;
      setHistory(res[0]);
      lastChange.current = 0;
      apply(res[1]);
    },
    [apply]
  );
  const undo = useCallback(
    () => restore(undoHistory(history, current)),
    [restore, history, current]
  );
  const redo = useCallback(
    () => restore(redoHistory(history, current)),
    [restore, history, current]
  );

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (!e.ctrlKey && !e.metaKey) return;
      const isUndo = (e.key === "z" || e.key === "Z") && !e.shiftKey;
      const isRedo =
        e.key === "y" || ((e.key === "z" || e.key === "Z") && e.shiftKey);
      if (!isUndo && !isRedo) return;

      // Dialogs opened on top, e.g. the JSON code, handle it themselves.
      const dialog = containerRef.current?.closest("[role=dialog]");
      if (!dialog?.contains(e.target as Node)) return;

      e.preventDefault();
      // Runs before the message editor's handler, so it doesn't also undo
      // when the flow of a button is edited inside the message editor.
      e.stopPropagation();
      isUndo ? undo() : redo();
    }
    document.addEventListener("keydown", onKeyDown, true);
    return () => document.removeEventListener("keydown", onKeyDown, true);
  }, [containerRef, undo, redo]);

  return {
    change,
    undo,
    redo,
    canUndo: history.past.length > 0,
    canRedo: history.future.length > 0,
  };
}

function ModalEditorBody({
  modal,
  components,
  onChange: applyChange,
  errors,
}: {
  modal: ModalData | undefined;
  components: ModalComponentData[];
  onChange: (modal: ModalData) => void;
  errors: Record<string, string>;
}) {
  const currentModal = useMemo(
    () => ({ title: modal?.title, components }),
    [modal?.title, components]
  );
  const containerRef = useRef<HTMLDivElement>(null);
  const history = useModalHistory(currentModal, applyChange, containerRef);
  const onChange = history.change;

  const componentKeys = useListKeys(components.length);
  const readOnly = useContext(FlowReadOnlyContext);
  const [activeKey, setActiveKey] = useState<number | null>(null);
  const activeIndex =
    activeKey === null ? -1 : componentKeys.keys.indexOf(activeKey);

  const cardRefs = useRef(new Map<number, HTMLDivElement>());
  const scrollToKey = useRef<number | null>(null);
  useEffect(() => {
    if (scrollToKey.current === null) return;
    const el = cardRefs.current.get(scrollToKey.current);
    scrollToKey.current = null;
    // Wait for the card to open, so the whole card is scrolled into view.
    requestAnimationFrame(() =>
      el?.scrollIntoView({ block: "nearest", behavior: "smooth" })
    );
  }, [activeKey]);

  const select = useCallback((key: number | null, scroll = false) => {
    if (scroll) scrollToKey.current = key;
    setActiveKey(key);
    setHighlight(key !== null);
  }, []);

  // Whether the open component is outlined in the preview. Clicking anywhere
  // but on it removes the outline, the card stays open.
  const [highlight, setHighlight] = useState(false);
  useEffect(() => {
    if (!highlight) return;
    function onPointerDown(e: PointerEvent) {
      const target = e.target as Element | null;
      if (!target?.closest?.(".kite-modal-component-active")) {
        setHighlight(false);
      }
    }
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [highlight]);

  const setComponents = useCallback(
    (newComponents: ModalComponentData[]) =>
      onChange({ title: modal?.title, components: newComponents }),
    [onChange, modal?.title]
  );

  const updateComponent = useCallback(
    (i: number, newData: Partial<ModalComponentData>) =>
      setComponents(
        components.map((c, j) => (j === i ? { ...c, ...newData } : c))
      ),
    [setComponents, components]
  );

  const updateInput = useCallback(
    (i: number, newData: Partial<ModalComponentData>) => {
      const input = components[i]?.components?.[0] ?? {};
      updateComponent(i, { components: [{ ...input, ...newData }] });
    },
    [updateComponent, components]
  );

  const moveComponent = useCallback(
    (i: number, dir: -1 | 1) => {
      const j = i + dir;
      if (j < 0 || j >= components.length) return;
      const res = [...components];
      [res[i], res[j]] = [res[j], res[i]];
      componentKeys.move(i, j);
      setComponents(res);
    },
    [setComponents, components, componentKeys]
  );

  const removeComponent = useCallback(
    (i: number) => {
      componentKeys.remove(i);
      setComponents(components.filter((_, j) => j !== i));
    },
    [setComponents, components, componentKeys]
  );

  const duplicateComponent = useCallback(
    (i: number) => {
      const copy = duplicateModalComponent(components, i);
      const key = componentKeys.insert(i + 1);
      setComponents([
        ...components.slice(0, i + 1),
        copy,
        ...components.slice(i + 1),
      ]);
      select(key, true);
    },
    [setComponents, components, componentKeys, select]
  );

  const addComponent = useCallback(
    (type: ModalInputType | "text_display") => {
      let component: ModalComponentData;
      if (type === "text_display") {
        component = { type: "text_display", content: "" };
      } else {
        const n = nextModalInputNumber(components);
        component = {
          type: "label",
          label: `Input ${n}`,
          components: [newModalInput(type, `input_${n}`)],
        };
      }
      const key = componentKeys.insert(components.length);
      setComponents([...components, component]);
      select(key, true);
    },
    [setComponents, components, componentKeys, select]
  );

  const clear = (keepTitle: boolean) => {
    onChange({ title: keepTitle ? modal?.title : undefined, components: [] });
    setActiveKey(null);
  };

  const errorIndexes = useMemo(() => componentErrorIndexes(errors), [errors]);
  const componentsError =
    errors["modal_data.components"] || errors["modal_data"];
  const full = components.length >= modalMaxComponents;

  const { theme } = useHookedTheme();
  const preview = (
    <ModalPreview
      modal={currentModal}
      lightTheme={theme === "light"}
      activeIndex={highlight ? activeIndex : null}
      errorIndexes={errorIndexes}
      onSelect={(i) => select(componentKeys.keys[i] ?? null, true)}
    />
  );

  return (
    <div ref={containerRef} className="flex h-full min-h-0 w-full flex-col">
      <DialogTitle className="hidden">Modal Editor</DialogTitle>
      <DialogDescription className="hidden">
        Edit the modal shown to the user.
      </DialogDescription>

      <div className="flex flex-auto min-h-0 flex-col xl:flex-row h-full">
        <ScrollArea className="xl:w-7/12 h-full">
          {/* Previews show the modal without letting it change. */}
          <PreviewInert className="pt-3 pb-24 xl:pb-8 px-3 md:px-5 lg:px-10 space-y-8">
            <div className="flex items-center justify-between space-x-3">
              <div className="flex items-center space-x-3">
                <MessageControlsButton
                  icon={UndoIcon}
                  label="Undo"
                  onClick={history.undo}
                  disabled={!history.canUndo}
                />
                <MessageControlsButton
                  icon={RedoIcon}
                  label="Redo"
                  onClick={history.redo}
                  disabled={!history.canRedo}
                />
              </div>
              <div className="flex items-center space-x-3">
                <ModalJSONDialog modal={currentModal} onChange={onChange}>
                  <MessageControlsButton
                    icon={CodeIcon}
                    label="JSON Code"
                    onClick={() => {}}
                  />
                </ModalJSONDialog>
                <ConfirmDialog
                  title="Are you sure that you want to clear the modal?"
                  description="This will clear the title and all components."
                  onConfirm={() => clear(false)}
                >
                  <MessageControlsButton
                    icon={PaintbrushIcon}
                    label="Clear Modal"
                    onClick={() => {}}
                  />
                </ConfirmDialog>
              </div>
            </div>

            <TextField
              title="Title"
              field="modal_data.title"
              value={modal?.title || ""}
              onChange={(v) => onChange({ title: v || undefined, components })}
              errors={errors}
              limit={modalTextLimits.title}
            />

            <MessageCollapsibleSection
              title="Components"
              count={`${components.length} / ${modalMaxComponents}`}
              className="space-y-4"
              animate={false}
            >
              {components.map((component, i) => {
                const key = componentKeys.keys[i];
                return (
                  <ComponentCard
                    key={key}
                    cardRef={(el) => {
                      if (el) cardRefs.current.set(key, el);
                      else cardRefs.current.delete(key);
                    }}
                    index={i}
                    components={components}
                    // Previews show every component, their cards can't be
                    // opened one by one.
                    open={readOnly || activeKey === key}
                    hasError={errorIndexes.has(i)}
                    onToggle={() => select(activeKey === key ? null : key)}
                    onMove={(dir) => moveComponent(i, dir)}
                    onDuplicate={full ? undefined : () => duplicateComponent(i)}
                    onRemove={() => removeComponent(i)}
                    updateComponent={(d) => updateComponent(i, d)}
                    updateInput={(d) => updateInput(i, d)}
                    errors={errors}
                  />
                );
              })}

              {componentsError && <FieldError error={componentsError} />}

              <div className="space-x-3">
                <AddComponentMenu onAdd={addComponent} disabled={full} />
                <Button variant="outline" onClick={() => clear(true)}>
                  Clear Components
                </Button>
              </div>
            </MessageCollapsibleSection>
          </PreviewInert>
        </ScrollArea>

        <div className="hidden xl:block w-5/12 h-full py-5 pr-5">
          <div className="h-full overflow-y-auto overflow-x-hidden rounded-lg border">
            {preview}
          </div>
        </div>
      </div>

      <Drawer>
        <DrawerTrigger asChild>
          <Button
            size="icon"
            className="fixed bottom-5 right-5 xl:hidden"
            aria-label="Preview modal"
          >
            <ViewIcon />
          </Button>
        </DrawerTrigger>
        <DrawerContent>
          <div className="max-h-[80dvh] overflow-x-hidden overflow-y-auto mt-3">
            {preview}
          </div>
        </DrawerContent>
      </Drawer>
    </div>
  );
}

function AddComponentMenu({
  onAdd,
  disabled,
}: {
  onAdd: (type: ModalInputType | "text_display") => void;
  disabled?: boolean;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild disabled={disabled}>
        <Button className="space-x-2">
          <div>Add Component</div>
          <ChevronDownIcon className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        {modalInputTypes.map((t) => (
          <DropdownMenuItem key={t.value} onClick={() => onAdd(t.value)}>
            {t.label}
          </DropdownMenuItem>
        ))}
        <DropdownMenuItem onClick={() => onAdd("text_display")}>
          Text Display
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function ComponentCard({
  cardRef,
  index,
  components,
  open,
  hasError,
  onToggle,
  onMove,
  onDuplicate,
  onRemove,
  updateComponent,
  updateInput,
  errors,
}: {
  cardRef: (el: HTMLDivElement | null) => void;
  index: number;
  components: ModalComponentData[];
  open: boolean;
  hasError: boolean;
  onToggle: () => void;
  onMove: (dir: -1 | 1) => void;
  onDuplicate?: () => void;
  onRemove: () => void;
  updateComponent: (data: Partial<ModalComponentData>) => void;
  updateInput: (data: Partial<ModalComponentData>) => void;
  errors: Record<string, string>;
}) {
  const component = components[index];
  const isText = component.type === "text_display";
  const input = component.components?.[0] ?? {};
  const number = modalComponentNumber(components, index);

  // Like MessageNodeActions, but actions that can't be taken are disabled
  // instead of hidden, so the buttons don't move while clicking through.
  const actions = [
    {
      label: "Move up",
      Icon: ChevronUpIcon,
      iconClass: "h-6 w-6",
      onClick: index > 0 ? () => onMove(-1) : undefined,
    },
    {
      label: "Move down",
      Icon: ChevronDownIcon,
      iconClass: "h-6 w-6",
      onClick: index < components.length - 1 ? () => onMove(1) : undefined,
    },
    {
      label: "Duplicate",
      Icon: CopyIcon,
      iconClass: "h-5 w-5",
      onClick: onDuplicate,
    },
    {
      label: "Remove",
      Icon: TrashIcon,
      iconClass: "h-5 w-5",
      onClick: onRemove,
    },
  ];

  return (
    <Card ref={cardRef} className="px-4 py-3 scroll-mt-3">
      <div className="flex items-center">
        <button
          type="button"
          className="flex items-center space-x-1 flex-auto min-w-0 text-lg text-foreground/90 text-left"
          onClick={onToggle}
          aria-expanded={open}
        >
          <ChevronRightIcon
            className={cn(
              "h-6 w-6 flex-none transition-transform text-muted-foreground",
              open && "rotate-90"
            )}
          />
          <div className="flex-none">
            {isText ? "Text" : "Input"} {number}
          </div>
          <div className="pl-1 text-sm italic font-light text-muted-foreground truncate">
            {isText ? "" : component.label || inputTypeLabel(input.type)}
          </div>
          {hasError && (
            <div className="pl-1 flex-none">
              <CircleAlertIcon className="h-5 w-5 text-red-600 dark:text-red-400" />
            </div>
          )}
        </button>
        <div className="flex-none flex items-center space-x-1">
          {actions.map(({ label, Icon, iconClass, onClick }) => (
            <Button
              key={label}
              type="button"
              variant="ghost"
              size="icon"
              className="h-9 w-9"
              disabled={!onClick}
              onClick={onClick}
              aria-label={label}
            >
              <Icon className={iconClass} />
            </Button>
          ))}
        </div>
      </div>

      {open && (
        <div className="mt-3 space-y-5">
          {isText ? (
            <TextField
              type="textarea"
              title="Content"
              field={`${componentField(index)}.content`}
              value={component.content || ""}
              onChange={(v) => updateComponent({ content: v || undefined })}
              errors={errors}
              limit={modalTextLimits.content}
            />
          ) : (
            <LabelEditor
              index={index}
              components={components}
              updateComponent={updateComponent}
              updateInput={updateInput}
              errors={errors}
            />
          )}
          <OtherErrors index={index} errors={errors} />
        </div>
      )}
    </Card>
  );
}

// The settings of the components that have their own input in the editor.
// Errors of other paths, e.g. a wrong type, are listed below the settings so
// they don't go unnoticed.
const shownErrorFields = [
  "label",
  "description",
  "content",
  "components.0.type",
  "components.0.custom_id",
  "components.0.style",
  "components.0.placeholder",
  "components.0.value",
  "components.0.min_length",
  "components.0.max_length",
  "components.0.min_values",
  "components.0.max_values",
  "components.0.channel_types",
  "components.0.required",
  "components.0.default",
  "components.0.options",
];

function OtherErrors({
  index,
  errors,
}: {
  index: number;
  errors: Record<string, string>;
}) {
  const prefix = `${componentField(index)}.`;
  const other = Object.entries(errors).filter(([key]) => {
    if (!key.startsWith(prefix)) return false;
    const rest = key.slice(prefix.length);
    return (
      !shownErrorFields.includes(rest) &&
      !rest.startsWith("components.0.options.")
    );
  });
  if (!other.length) return null;

  return (
    <div className="space-y-1">
      {other.map(([key, error]) => (
        <FieldError key={key} error={error} />
      ))}
    </div>
  );
}

function LabelEditor({
  index,
  components,
  updateComponent,
  updateInput,
  errors,
}: {
  index: number;
  components: ModalComponentData[];
  updateComponent: (data: Partial<ModalComponentData>) => void;
  updateInput: (data: Partial<ModalComponentData>) => void;
  errors: Record<string, string>;
}) {
  const component = components[index];
  const input = component.components?.[0] ?? {};
  const type = (input.type || "text_input") as ModalInputType;
  const field = componentField(index);
  const iField = inputField(index);
  const isFile = type === "file_upload";

  return (
    <>
      <div className="flex space-x-3">
        <div className="flex-auto min-w-0">
          <FieldLabel title="Type" />
          <Select
            value={type}
            onValueChange={(v) =>
              updateComponent({
                components: [changeModalInputType(input, v as ModalInputType)],
              })
            }
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {modalInputTypes.map((t) => {
                const Icon = inputTypeIcons[t.value];
                return (
                  <SelectItem key={t.value} value={t.value}>
                    <div className="flex items-center gap-2">
                      <Icon className="h-4 w-4 text-muted-foreground" />
                      {t.label}
                    </div>
                  </SelectItem>
                );
              })}
            </SelectContent>
          </Select>
          <FieldError error={errors[`${iField}.type`]} />
        </div>
        {type === "checkbox" ? (
          <SwitchField
            title="Checked"
            field={`${iField}.default`}
            checked={!!input.default}
            onChange={(v) => updateInput({ default: v || undefined })}
            errors={errors}
          />
        ) : (
          <SwitchField
            title="Required"
            field={`${iField}.required`}
            checked={!!input.required}
            onChange={(v) => updateInput({ required: v || undefined })}
            errors={errors}
          />
        )}
      </div>

      <div className="grid gap-5 md:gap-3 md:grid-cols-2">
        <TextField
          title="Label"
          field={`${field}.label`}
          value={component.label || ""}
          onChange={(v) => updateComponent({ label: v || undefined })}
          errors={errors}
          limit={modalTextLimits.label}
        />
        <IdentifierField
          index={index}
          components={components}
          updateInput={updateInput}
          errors={errors}
        />
      </div>

      <TextField
        title="Description"
        field={`${field}.description`}
        value={component.description || ""}
        onChange={(v) => updateComponent({ description: v || undefined })}
        errors={errors}
        limit={modalTextLimits.description}
      />

      {type === "text_input" && (
        <div>
          <FieldLabel title="Style" />
          <ToggleGroup
            type="single"
            variant="outline"
            className="justify-start"
            value={input.style === 2 ? "2" : "1"}
            onValueChange={(v) => v && updateInput({ style: parseInt(v) })}
          >
            <ToggleGroupItem value="1">Short</ToggleGroupItem>
            <ToggleGroupItem value="2">Paragraph</ToggleGroupItem>
          </ToggleGroup>
          <FieldError error={errors[`${iField}.style`]} />
        </div>
      )}

      {modalInputHasPlaceholder(type) && (
        <TextField
          title="Placeholder"
          field={`${iField}.placeholder`}
          value={input.placeholder || ""}
          onChange={(v) => updateInput({ placeholder: v || undefined })}
          errors={errors}
          limit={
            type === "text_input"
              ? modalTextLimits.placeholder
              : modalTextLimits.select_placeholder
          }
        />
      )}

      {type === "text_input" && (
        <>
          <TextField
            type={input.style === 2 ? "textarea" : "text"}
            title="Pre-filled Value"
            field={`${iField}.value`}
            value={input.value || ""}
            onChange={(v) => updateInput({ value: v || undefined })}
            errors={errors}
            limit={modalTextLimits.value}
          />
          <div className="flex space-x-3">
            <NumberField
              title="Min Length"
              field={`${iField}.min_length`}
              value={input.min_length}
              onChange={(v) => updateInput({ min_length: v })}
              placeholder="0"
              min={0}
              max={4000}
              errors={errors}
            />
            <NumberField
              title="Max Length"
              field={`${iField}.max_length`}
              value={input.max_length}
              onChange={(v) => updateInput({ max_length: v })}
              placeholder="4000"
              min={1}
              max={4000}
              errors={errors}
            />
          </div>
        </>
      )}

      {modalInputHasValueLimits(type) && (
        <div className="flex space-x-3">
          <NumberField
            title={isFile ? "Min Files" : "Min Picks"}
            field={`${iField}.min_values`}
            value={input.min_values}
            onChange={(v) => updateInput({ min_values: v })}
            placeholder={input.required ? "1" : "0"}
            min={1}
            max={isFile ? modalMaxFiles : 25}
            errors={errors}
          />
          <NumberField
            title={isFile ? "Max Files" : "Max Picks"}
            field={`${iField}.max_values`}
            value={input.max_values}
            onChange={(v) => updateInput({ max_values: v })}
            placeholder={
              type === "checkbox_group"
                ? (input.options?.length ?? 1).toString()
                : "1"
            }
            min={1}
            max={isFile ? modalMaxFiles : 25}
            errors={errors}
          />
        </div>
      )}

      {type === "channel_select" && (
        <div>
          <ChannelTypesInput
            values={input.channel_types ?? []}
            onChange={(v) =>
              updateInput({ channel_types: v.length ? v : undefined })
            }
          />
          <FieldError error={errors[`${iField}.channel_types`]} />
        </div>
      )}

      {modalInputHasOptions(type) && (
        <OptionsEditor
          field={`${iField}.options`}
          type={type as keyof typeof modalOptionCounts}
          options={input.options ?? []}
          onChange={(options) => updateInput({ options })}
          errors={errors}
        />
      )}
    </>
  );
}

function IdentifierField({
  index,
  components,
  updateInput,
  errors,
}: {
  index: number;
  components: ModalComponentData[];
  updateInput: (data: Partial<ModalComponentData>) => void;
  errors: Record<string, string>;
}) {
  const component = components[index];
  const id = component.components?.[0]?.custom_id || "";
  const placeholder = `{{input('${id}')}}`;

  const copy = () => {
    navigator.clipboard
      .writeText(placeholder)
      .then(() => toast.success(`Copied ${placeholder}`))
      .catch(() => toast.error("Couldn't copy the placeholder"));
  };

  return (
    <div className="min-w-0">
      <FieldLabel
        title="Identifier"
        count={`${id.length} / ${modalTextLimits.custom_id}`}
      />
      <div className="flex space-x-1">
        <Input
          value={id}
          onChange={(e) =>
            updateInput({ custom_id: e.target.value || undefined })
          }
        />
        <IconButton
          label="Make from label"
          disabled={!component.label}
          onClick={() =>
            updateInput({
              custom_id: modalIdentifierFromLabel(
                component.label || "",
                components,
                index
              ),
            })
          }
        >
          <WandSparklesIcon className="h-4 w-4" />
        </IconButton>
        <IconButton label={`Copy ${placeholder}`} disabled={!id} onClick={copy}>
          <CopyIcon className="h-4 w-4" />
        </IconButton>
      </div>
      <div className="text-muted-foreground text-sm mt-1.5 truncate">
        Read the answer with {id ? placeholder : "{{input('identifier')}}"}
      </div>
      <FieldError error={errors[`${inputField(index)}.custom_id`]} />
    </div>
  );
}

function OptionsEditor({
  field,
  type,
  options,
  onChange,
  errors,
}: {
  field: string;
  type: keyof typeof modalOptionCounts;
  options: ModalComponentOptionData[];
  onChange: (options: ModalComponentOptionData[]) => void;
  errors: Record<string, string>;
}) {
  const [min, max] = modalOptionCounts[type];
  const optionKeys = useListKeys(options.length);

  const update = (i: number, data: Partial<ModalComponentOptionData>) =>
    onChange(options.map((o, j) => (j === i ? { ...o, ...data } : o)));

  const setDefault = (i: number, value: boolean) => {
    // A radio group has one default at most, so picking one unpicks the rest.
    if (type === "radio_group" && value) {
      onChange(
        options.map((o, j) => ({ ...o, default: j === i ? true : undefined }))
      );
      return;
    }
    update(i, { default: value || undefined });
  };

  const move = (i: number, dir: -1 | 1) => {
    const j = i + dir;
    if (j < 0 || j >= options.length) return;
    const res = [...options];
    [res[i], res[j]] = [res[j], res[i]];
    optionKeys.move(i, j);
    onChange(res);
  };

  return (
    <div>
      <FieldLabel title="Options" count={`${options.length} / ${max}`} />
      <div className="text-muted-foreground text-sm mb-3">
        The value is what input() returns when the option is picked. Leave it
        empty to use the label.
      </div>

      <div className="space-y-3">
        {options.map((option, i) => {
          // Every error of the option, e.g. also of its emoji's name.
          const optionErrors = Object.entries(errors)
            .filter(
              ([key]) =>
                key === `${field}.${i}` || key.startsWith(`${field}.${i}.`)
            )
            .map(([, error]) => error);

          return (
            <div key={optionKeys.keys[i]} className="space-y-2">
              <div className="flex gap-2 items-center">
                {type === "string_select" && (
                  <OptionEmojiButton
                    emoji={option.emoji}
                    onChange={(emoji) => update(i, { emoji })}
                  />
                )}
                <div className="flex-auto grid grid-cols-3 gap-2">
                  <PlaceholderInput
                    value={option.label || ""}
                    onChange={(v) => update(i, { label: v || undefined })}
                    placeholder="Label"
                  />
                  <PlaceholderInput
                    value={option.value || ""}
                    onChange={(v) => update(i, { value: v || undefined })}
                    placeholder={option.label || "Value"}
                  />
                  <PlaceholderInput
                    value={option.description || ""}
                    onChange={(v) => update(i, { description: v || undefined })}
                    placeholder="Description"
                  />
                </div>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="flex-none">
                      <Switch
                        checked={!!option.default}
                        onCheckedChange={(v) => setDefault(i, v)}
                      />
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>Picked by default</TooltipContent>
                </Tooltip>
                <div className="flex flex-none">
                  <IconButton
                    label="Move up"
                    disabled={i === 0}
                    onClick={() => move(i, -1)}
                  >
                    <ChevronUpIcon className="h-5 w-5" />
                  </IconButton>
                  <IconButton
                    label="Move down"
                    disabled={i === options.length - 1}
                    onClick={() => move(i, 1)}
                  >
                    <ChevronDownIcon className="h-5 w-5" />
                  </IconButton>
                  <IconButton
                    label="Remove"
                    disabled={options.length <= min}
                    onClick={() => {
                      optionKeys.remove(i);
                      onChange(options.filter((_, j) => j !== i));
                    }}
                  >
                    <XIcon className="h-4 w-4" />
                  </IconButton>
                </div>
              </div>
              {optionErrors.map((error, j) => (
                <FieldError key={j} error={error} />
              ))}
            </div>
          );
        })}
      </div>

      <Button
        variant="outline"
        className="mt-3"
        disabled={options.length >= max}
        onClick={() => onChange([...options, newModalOption(options)])}
      >
        Add Option
      </Button>
      <FieldError error={errors[field]} />
    </div>
  );
}

// Kept at module level, react-twemoji re-parses its children whenever the
// options object changes.
const optionEmojiOptions = { className: "h-5 w-5" };

function OptionEmojiButton({
  emoji,
  onChange,
}: {
  emoji?: EmojiData;
  onChange: (emoji: EmojiData | undefined) => void;
}) {
  return (
    <div className="flex-none relative">
      <EmojiPicker
        onEmojiSelect={(e) =>
          onChange(e.native ? { name: e.name } : { id: e.id, name: e.name })
        }
      >
        <Button
          type="button"
          size="icon"
          variant="outline"
          aria-label="Pick emoji"
        >
          {emoji?.id ? (
            <img src={discordEmojiUrl(emoji.id)} alt="" className="h-5 w-5" />
          ) : emoji?.name ? (
            <Twemoji options={optionEmojiOptions}>{emoji.name}</Twemoji>
          ) : (
            <SmileIcon className="h-5 w-5 text-muted-foreground" />
          )}
        </Button>
      </EmojiPicker>
      {emoji && (
        <button
          type="button"
          aria-label="Remove emoji"
          className="absolute -top-1.5 -right-1.5 rounded-full bg-muted text-muted-foreground hover:text-foreground p-0.5"
          onClick={() => onChange(undefined)}
        >
          <XIcon className="h-3 w-3" />
        </button>
      )}
    </div>
  );
}

function IconButton({
  label,
  children,
  disabled,
  onClick,
}: {
  label: string;
  children: ReactNode;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            disabled={disabled}
            onClick={onClick}
            aria-label={label}
          >
            {children}
          </Button>
        </span>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

// The label of an input, styled like the inputs of the message editor.
function FieldLabel({ title, count }: { title: string; count?: ReactNode }) {
  return (
    <div className="flex space-x-2 mb-2 items-center">
      <Label className="text-base text-slate-800 dark:text-slate-200">
        {title}
      </Label>
      {count !== undefined && (
        <div className="text-sm italic font-light text-muted-foreground">
          {count}
        </div>
      )}
    </div>
  );
}

function TextField({
  type = "text",
  title,
  field,
  value,
  onChange,
  errors,
  limit,
}: {
  type?: "text" | "textarea";
  title: string;
  field: string;
  value: string;
  onChange: (value: string) => void;
  errors: Record<string, string>;
  limit?: number;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Inserts the picked placeholder at the cursor, like BaseInput in
  // FlowNodeEditor does.
  const onPlaceholderSelect = useCallback(
    (p: string) => {
      const element =
        type === "textarea" ? textareaRef.current : inputRef.current;
      if (!element) return;

      const start = element.selectionStart ?? element.value.length;
      const end = element.selectionEnd ?? element.value.length;
      onChange(
        element.value.substring(0, start) +
          `{{${p}}}` +
          element.value.substring(end)
      );
    },
    [type, onChange]
  );

  return (
    <div className="w-full min-w-0">
      <FieldLabel
        title={title}
        count={limit !== undefined ? `${value.length} / ${limit}` : undefined}
      />
      <div className="relative">
        {type === "textarea" ? (
          <Textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            ref={textareaRef}
          />
        ) : (
          <PlaceholderInput value={value} onChange={onChange} ref={inputRef} />
        )}
        <FlowPlaceholderExplorer onSelect={onPlaceholderSelect} />
      </div>
      <FieldError error={errors[field]} />
    </div>
  );
}

function NumberField({
  title,
  field,
  value,
  onChange,
  placeholder,
  min,
  max,
  errors,
}: {
  title: string;
  field: string;
  value: number | undefined;
  onChange: (value: number | undefined) => void;
  placeholder?: string;
  min: number;
  max: number;
  errors: Record<string, string>;
}) {
  return (
    <div className="w-full min-w-0">
      <FieldLabel title={title} />
      <Input
        type="number"
        inputMode="numeric"
        min={min}
        max={max}
        value={value?.toString() ?? ""}
        placeholder={placeholder}
        onChange={(e) => onChange(parseOptionalInt(e.target.value))}
      />
      <FieldError error={errors[field]} />
    </div>
  );
}

function SwitchField({
  title,
  field,
  checked,
  onChange,
  errors,
}: {
  title: string;
  field: string;
  checked: boolean;
  onChange: (value: boolean) => void;
  errors: Record<string, string>;
}) {
  return (
    <div className="flex-none">
      <FieldLabel title={title} />
      <Switch checked={checked} onCheckedChange={onChange} />
      <FieldError error={errors[field]} />
    </div>
  );
}

function FieldError({ error }: { error?: string }) {
  if (!error) return null;
  return (
    <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
      <CircleAlertIcon className="h-5 w-5 flex-none" />
      <div>{error}</div>
    </div>
  );
}
