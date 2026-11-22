import {
  createContext,
  FormEvent,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";
import {
  EllipsisIcon,
  FolderIcon,
  FolderInputIcon,
  FolderMinusIcon,
  FolderPlusIcon,
  PencilIcon,
} from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import LoadingButton from "../common/LoadingButton";
import { FolderItemKind, useFolderUpdateMutation } from "@/lib/api/mutations";
import { useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { GroupableItem, ItemGroup, itemGroup } from "@/lib/flow/itemGroups";

type FolderItem = GroupableItem & { id: string };

const maxFolderLength = 100;

// useItemGroupOf returns the group an item is listed under.
export function useItemGroupOf() {
  const installs = useBundleInstalls();
  return (item: GroupableItem) => itemGroup(item, installs);
}

// The list the folder menus act on: the kind of its items, and all of them,
// not just the ones a search shows.
const FolderItemsContext = createContext<{
  kind: FolderItemKind;
  items: FolderItem[];
}>({ kind: "messages", items: [] });

export function FolderItemsProvider({
  kind,
  items,
  children,
}: {
  kind: FolderItemKind;
  items: (FolderItem | null | undefined)[];
  children: ReactNode;
}) {
  return (
    <FolderItemsContext.Provider
      value={{ kind, items: items.flatMap((i) => (i ? [i] : [])) }}
    >
      {children}
    </FolderItemsContext.Provider>
  );
}

function useUpdateFolder(kind: FolderItemKind) {
  const mutation = useFolderUpdateMutation(useAppId(), kind);

  function update(ids: string[], folder: string | null, onDone?: () => void) {
    mutation.mutate(
      { ids, folder },
      {
        onSuccess(res) {
          if (res.success) {
            onDone?.();
          } else {
            toast.error(
              `Failed to update folder: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return { update, pending: mutation.isPending };
}

// useMoveToFolder is the Move to folder item of an item's menu, and the dialog
// that names a new folder. The dialog goes outside the menu, which unmounts its
// content when it closes.
export function useMoveToFolder(item: FolderItem) {
  const [open, setOpen] = useState(false);
  const { kind, items } = useContext(FolderItemsContext);
  const groupOf = useItemGroupOf();
  const { update, pending } = useUpdateFolder(kind);

  const current = groupOf(item)?.name;
  // Moving an item to a template's group gives it a folder of that name, so
  // it's listed with the template's other items.
  const targets = [...new Set(items.flatMap((i) => groupOf(i)?.name ?? []))]
    .filter((name) => name !== current)
    .sort((a, b) => a.localeCompare(b));

  const menu = (
    <DropdownMenuSub>
      <DropdownMenuSubTrigger>
        <FolderInputIcon className="h-4 w-4 mr-2 text-muted-foreground" />
        Move to folder
      </DropdownMenuSubTrigger>
      <DropdownMenuSubContent className="max-h-80 overflow-y-auto">
        {targets.map((name) => (
          <DropdownMenuItem key={name} onSelect={() => update([item.id], name)}>
            <FolderIcon className="h-4 w-4 mr-2 text-muted-foreground" />
            <span className="truncate max-w-56">{name}</span>
          </DropdownMenuItem>
        ))}
        {targets.length > 0 && <DropdownMenuSeparator />}
        {item.folder && (
          <DropdownMenuItem onSelect={() => update([item.id], null)}>
            <FolderMinusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
            No folder
          </DropdownMenuItem>
        )}
        <DropdownMenuItem onSelect={() => setOpen(true)}>
          <FolderPlusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
          New folder
        </DropdownMenuItem>
      </DropdownMenuSubContent>
    </DropdownMenuSub>
  );

  const dialog = (
    <FolderNameDialog
      open={open}
      onOpenChange={setOpen}
      title="New folder"
      description="Items with the same folder are listed together."
      submitLabel="Move to folder"
      pending={pending}
      onSubmit={(name) => update([item.id], name, () => setOpen(false))}
    />
  );

  return { menu, dialog };
}

// FolderActions renames or removes a folder. Removing it keeps its items and
// lists them without a folder.
export function FolderActions({ group }: { group: ItemGroup }) {
  const [open, setOpen] = useState(false);
  const { kind, items } = useContext(FolderItemsContext);
  const { update, pending } = useUpdateFolder(kind);

  if (group.template) return null;

  const ids = items.filter((i) => i.folder === group.name).map((i) => i.id);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            size="icon"
            variant="ghost"
            className="h-8 w-8"
            aria-label={`Options of the ${group.name} folder`}
          >
            <EllipsisIcon className="h-4 w-4 text-muted-foreground" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => setOpen(true)}>
            <PencilIcon className="h-4 w-4 mr-2 text-muted-foreground" />
            Rename folder
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => update(ids, null)}>
            <FolderMinusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
            Remove folder
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <FolderNameDialog
        open={open}
        onOpenChange={setOpen}
        title="Rename folder"
        initialName={group.name}
        submitLabel="Rename folder"
        pending={pending}
        onSubmit={(name) => update(ids, name, () => setOpen(false))}
      />
    </>
  );
}

function FolderNameDialog({
  open,
  onOpenChange,
  title,
  description,
  initialName = "",
  submitLabel,
  pending,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: ReactNode;
  initialName?: string;
  submitLabel: string;
  pending: boolean;
  onSubmit: (name: string) => void;
}) {
  const [name, setName] = useState(initialName);

  useEffect(() => {
    if (open) setName(initialName);
  }, [open, initialName]);

  function submit(e: FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed || pending) return;
    onSubmit(trimmed);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <form onSubmit={submit} className="grid gap-4">
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Folder name"
            aria-label="Folder name"
            maxLength={maxFolderLength}
            autoFocus
          />
          <DialogFooter>
            <LoadingButton
              type="submit"
              loading={pending}
              disabled={!name.trim()}
            >
              {submitLabel}
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
