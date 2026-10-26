import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import {
  getItemVersion,
  ItemVersionTarget,
  useItemVersionsQuery,
} from "../api/queries";
import { ItemVersion } from "../types/wire.gen";
import { useResponseData } from "./api";

export interface VersionHistory {
  // The versions are only loaded while the history is open.
  open: boolean;
  setOpen: (open: boolean) => void;
  versions?: ItemVersion[];
  isLoading: boolean;
  hasUnsavedChanges: boolean;
  // The version being restored, if any.
  restoringId: string | null;
  canUndoSave: boolean;
  undoSave: () => void;
  restore: (version: ItemVersion) => void;
  // Called on edits, so "Undo save" starts from the current save again.
  resetUndoChain: () => void;
}

// Loads the versions of a command, event listener or message and restores
// them. apply puts the content of a version into the editor and saves it, so
// a restore is a new version that can be undone too.
export function useVersionHistory({
  target,
  hasUnsavedChanges,
  apply,
}: {
  target: ItemVersionTarget;
  hasUnsavedChanges: boolean;
  apply: (content: Record<string, any>) => void;
}): VersionHistory {
  const [open, setOpen] = useState(false);
  const versionsQuery = useItemVersionsQuery(target, open);
  // The list has no missing entries, the type just allows them.
  const versions = useResponseData(versionsQuery) as ItemVersion[] | undefined;

  const [restoringId, setRestoringId] = useState<string | null>(null);
  // The version the last "Undo save" went back to, so pressing it again goes
  // further back instead of undoing the undo.
  const [undoChainId, setUndoChainId] = useState<string | null>(null);

  const restoreId = useCallback(
    async (versionId: string, isUndo: boolean) => {
      if (
        hasUnsavedChanges &&
        !confirm("Restoring replaces your unsaved changes. Continue?")
      ) {
        return;
      }

      setRestoringId(versionId);
      try {
        const res = await getItemVersion(target, versionId);
        if (!res.success) {
          toast.error(
            `Failed to load save: ${res.error.message} (${res.error.code})`
          );
          return;
        }
        if (!res.data.content) return;

        apply(res.data.content);
        setUndoChainId(isUndo ? versionId : null);
      } catch (err) {
        toast.error(`Failed to load save: ${err}`);
      } finally {
        setRestoringId(null);
      }
    },
    [target, hasUnsavedChanges, apply]
  );

  // The newest version is the current save, so undoing goes to the one after
  // it, or after the one the previous undo went to.
  const undoTarget = useMemo(() => {
    if (!versions) return undefined;

    const chainIndex = undoChainId
      ? versions.findIndex((v) => v.id === undoChainId)
      : -1;
    return versions[chainIndex === -1 ? 1 : chainIndex + 1];
  }, [versions, undoChainId]);

  const undoSave = useCallback(() => {
    if (undoTarget) restoreId(undoTarget.id, true);
  }, [undoTarget, restoreId]);

  const restore = useCallback(
    (version: ItemVersion) => restoreId(version.id, false),
    [restoreId]
  );

  const resetUndoChain = useCallback(() => setUndoChainId(null), []);

  return {
    open,
    setOpen,
    versions,
    isLoading: versionsQuery.isLoading,
    hasUnsavedChanges,
    restoringId,
    canUndoSave: !!undoTarget,
    undoSave,
    restore,
    resetUndoChain,
  };
}

// useEditorSession is made once per opened editor and sent with its saves, so
// they go into one version of the history, as the editor's undo covers the
// states in between. Opening it again starts a new version. Restores are saved
// without it, so they stay their own version and can be undone later.
export function useEditorSession() {
  const [editorSession] = useState(() => crypto.randomUUID());
  return editorSession;
}
