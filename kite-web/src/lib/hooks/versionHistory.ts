import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import {
  getItemVersion,
  ItemVersionTarget,
  useItemVersionsQuery,
} from "../api/queries";
import { ItemVersion } from "../types/wire.gen";

export interface VersionHistory {
  versions?: ItemVersion[];
  isLoading: boolean;
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
  target?: ItemVersionTarget;
  hasUnsavedChanges: boolean;
  apply: (content: Record<string, any>) => void;
}): VersionHistory {
  const versionsQuery = useItemVersionsQuery(
    target ?? { appId: "", commandId: "" },
    !!target
  );
  const versions = useMemo(
    () =>
      versionsQuery.data?.success
        ? versionsQuery.data.data.filter((v): v is ItemVersion => !!v)
        : undefined,
    [versionsQuery.data]
  );

  const [restoringId, setRestoringId] = useState<string | null>(null);
  // The version the last "Undo save" went back to, so pressing it again goes
  // further back instead of undoing the undo.
  const [undoChainId, setUndoChainId] = useState<string | null>(null);

  const restoreId = useCallback(
    async (versionId: string, isUndo: boolean) => {
      if (!target) return;
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
    versions,
    isLoading: versionsQuery.isLoading,
    restoringId,
    canUndoSave: !!undoTarget,
    undoSave,
    restore,
    resetUndoChain,
  };
}
